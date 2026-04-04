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

package remove

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/config/repomgr"
)

type useCase struct {
	cfg *config.ALRConfig
	pp  PackageProvider
}

type PackageProvider interface {
	DeletePkgs(ctx context.Context, where string, args ...any) error
}

func New(cfg *config.ALRConfig, pp PackageProvider) *useCase {
	return &useCase{cfg, pp}
}

type Options struct {
	Name string
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	name := opts.Name

	err := u.cfg.RemoveRepo(name)
	if err != nil {
		if errors.Is(err, repomgr.ErrSystemRepo) {
			return cliutils.FormatCliExit(gotext.Get("Repo \"%s\" is provided by a system package. Remove the providing package to delete it.", name), nil)
		}
		if errors.Is(err, repomgr.ErrRepoNotFound) {
			return cliutils.FormatCliExit(gotext.Get("Repo \"%s\" does not exist", name), nil)
		}
		return cliutils.FormatCliExit(gotext.Get("Error removing repo"), err)
	}

	if err := os.RemoveAll(filepath.Join(u.cfg.GetPaths().RepoDir, name)); err != nil {
		return cliutils.FormatCliExit(gotext.Get("Error removing repo directory"), err)
	}

	if err := cliutils.ExitIfCantDropCapsToBuilderUser(); err != nil {
		return err
	}

	if err := u.pp.DeletePkgs(ctx, "repository = ?", name); err != nil {
		return cliutils.FormatCliExit(gotext.Get("Error removing packages from database"), err)
	}

	return nil
}
