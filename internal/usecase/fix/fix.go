// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2025 The Stapler Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package fix

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/pkg/types"
)

type ReposPuller interface {
	Pull(ctx context.Context, repos []types.Repo) error
}

type ReposPullerGetter func(ctx context.Context) (ReposPuller, deps.Cleanup, error)

type Config interface {
	GetPaths() *config.Paths
}

type useCase struct {
	config Config
	init   ReposPullerGetter
}

func New(config Config, init ReposPullerGetter) *useCase {
	return &useCase{config: config, init: init}
}

func (u *useCase) reinit(ctx context.Context) error {
	r, f, err := u.init(ctx)
	if err != nil {
		return err
	}
	defer f()
	// TODO: replace with rereader
	return r.Pull(ctx, nil)
}

func (u *useCase) Run(ctx context.Context) error {
	paths := u.config.GetPaths()

	slog.Info(gotext.Get("Clearing cache directory"))

	dir, err := os.Open(paths.CacheDir)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Unable to open cache directory"))
	}
	defer dir.Close()

	entries, err := dir.Readdirnames(-1)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Unable to read cache directory contents"))
	}

	for _, entry := range entries {
		fullPath := filepath.Join(paths.CacheDir, entry)

		if err := makeWritableRecursive(fullPath); err != nil {
			slog.Debug("Failed to make path writable", "path", fullPath, "error", err)
		}

		err = os.RemoveAll(fullPath)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Unable to remove cache item (%s)", entry))
		}
	}

	slog.Info(gotext.Get("Rebuilding cache"))

	err = os.MkdirAll(paths.CacheDir, 0o755)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Unable to create new cache directory"))
	}

	if err := u.reinit(ctx); err != nil {
		return err
	}

	slog.Info(gotext.Get("Done"))

	return nil
}

func makeWritableRecursive(path string) error {
	return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		newMode := info.Mode() | 0o200
		if d.IsDir() {
			newMode |= 0o100
		}

		return os.Chmod(path, newMode)
	})
}
