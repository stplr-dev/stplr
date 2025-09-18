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
	"os"
	"path/filepath"
	"slices"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/utils"
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

	found := false
	index := 0
	reposSlice := u.cfg.Repos()
	for i, repo := range reposSlice {
		if repo.Name == name {
			index = i
			found = true
		}
	}
	if !found {
		return cliutils.FormatCliExit(gotext.Get("Repo \"%s\" does not exist", name), nil)
	}

	u.cfg.SetRepos(slices.Delete(reposSlice, index, index+1))

	err := os.RemoveAll(filepath.Join(u.cfg.GetPaths().RepoDir, name))
	if err != nil {
		return cliutils.FormatCliExit(gotext.Get("Error removing repo directory"), err)
	}
	err = u.cfg.System.Save()
	if err != nil {
		return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
	}

	if err := utils.ExitIfCantDropCapsToBuilderUser(); err != nil {
		return err
	}

	err = u.pp.DeletePkgs(ctx, "repository = ?", name)
	if err != nil {
		return cliutils.FormatCliExit(gotext.Get("Error removing packages from database"), err)
	}

	return nil
}
