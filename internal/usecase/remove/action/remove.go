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

package action

import (
	"context"
	"fmt"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/scripter"
	"go.stplr.dev/stplr/internal/service/repos"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type PackageFinder interface {
	FindPkgs(ctx context.Context, pkgs []string) (map[string][]staplerfile.Package, []string, error)
}

type useCase struct {
	mgr    manager.Manager
	finder PackageFinder
}

func New(manager manager.Manager, finder PackageFinder) *useCase {
	return &useCase{mgr: manager, finder: finder}
}

type Options struct {
	Pkgs        []string
	Interactive bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	found, _, err := u.finder.FindPkgs(ctx, opts.Pkgs)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error finding packages"))
	}

	systemPkgsName := make([]string, 0, len(opts.Pkgs))
	for _, pkg := range opts.Pkgs {
		name, repo, ok := repos.ExtractNameAndRepo(pkg)
		if ok {
			systemPkgsName = append(systemPkgsName, scripter.FormatName(name, repo))
			continue
		}

		systemName := pkg
		for _, candidate := range found[pkg] {
			candidateName := scripter.FormatName(candidate.Name, candidate.Repository)
			installed, err := u.mgr.IsInstalled(candidateName)
			if err != nil {
				return fmt.Errorf("check if package %s is installed: %w", candidateName, err)
			}
			if installed {
				systemName = candidateName
				break
			}
		}
		systemPkgsName = append(systemPkgsName, systemName)
	}

	if err := u.mgr.Remove(&manager.Opts{
		NoConfirm: !opts.Interactive,
	}, systemPkgsName...); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error removing packages"))
	}

	return nil
}
