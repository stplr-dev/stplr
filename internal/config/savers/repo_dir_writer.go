// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package savers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.stplr.dev/stplr/internal/constants"
)

type RepoDirWriter struct {
	UserDir      string
	OverridesDir string
}

func NewRepoDirWriter() *RepoDirWriter {
	return &RepoDirWriter{
		UserDir:      constants.UserReposDirPath,
		OverridesDir: constants.RepoOverridesDirPath,
	}
}

func (r *RepoDirWriter) WriteUserRepo(_ context.Context, name string, data []byte) error {
	if err := validateRepoName(name); err != nil {
		return err
	}
	if err := os.MkdirAll(r.UserDir, 0o755); err != nil {
		return fmt.Errorf("create repos dir: %w", err)
	}
	return os.WriteFile(filepath.Join(r.UserDir, name+".toml"), data, 0o644) //gosec:disable G306 -- repo config files in /etc must be world-readable
}

func (r *RepoDirWriter) RemoveUserRepo(_ context.Context, name string) error {
	if err := validateRepoName(name); err != nil {
		return err
	}
	return os.Remove(filepath.Join(r.UserDir, name+".toml"))
}

func (r *RepoDirWriter) WriteOverride(_ context.Context, name string, data []byte) error {
	if err := validateRepoName(name); err != nil {
		return err
	}
	if err := os.MkdirAll(r.OverridesDir, 0o755); err != nil {
		return fmt.Errorf("create overrides dir: %w", err)
	}
	return os.WriteFile(filepath.Join(r.OverridesDir, name+".toml"), data, 0o644) //gosec:disable G306 -- repo config files in /etc must be world-readable
}

func (r *RepoDirWriter) RemoveOverride(_ context.Context, name string) error {
	if err := validateRepoName(name); err != nil {
		return err
	}
	err := os.Remove(filepath.Join(r.OverridesDir, name+".toml"))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func validateRepoName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsRune(name, '/') {
		return fmt.Errorf("invalid repo name: %q", name)
	}
	return nil
}

var _ RepoDirWriterExecutor = (*RepoDirWriter)(nil)
