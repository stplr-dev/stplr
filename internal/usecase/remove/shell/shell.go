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

package shell

import (
	"context"
	"fmt"
	"os"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type PackageGetter interface {
	GetPkgs(ctx context.Context, where string, args ...any) ([]staplerfile.Package, error)
}

type PackageLister interface {
	ListInstalled(
		opts *manager.Opts,
	) (map[string]string, error)
}

type useCase struct {
	mgr       PackageLister
	pkgGetter PackageGetter
	stdout    *os.File
}

func New(mgr PackageLister, pkgGetter PackageGetter) *useCase {
	return &useCase{
		mgr:       mgr,
		pkgGetter: pkgGetter,
		stdout:    os.Stdout,
	}
}

func (u *useCase) Run(ctx context.Context) error {
	installedMap := map[string]string{}
	installedList, err := u.mgr.ListInstalled(&manager.Opts{})
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error listing installed packages"))
	}
	for pkgName, version := range installedList {
		matches := build.RegexpALRPackageName.FindStringSubmatch(pkgName)
		if matches != nil {
			packageName := matches[build.RegexpALRPackageName.SubexpIndex("package")]
			repoName := matches[build.RegexpALRPackageName.SubexpIndex("repo")]
			installedMap[fmt.Sprintf("%s/%s", repoName, packageName)] = version
		}
	}

	res, err := u.pkgGetter.GetPkgs(ctx, "true")
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error getting packages"))
	}

	for _, pkg := range res {
		pkgFullName := pkg.FormatFullName()
		_, ok := installedMap[pkgFullName]
		if !ok {
			continue
		}
		fmt.Fprintln(u.stdout, pkgFullName)
	}

	return nil
}
