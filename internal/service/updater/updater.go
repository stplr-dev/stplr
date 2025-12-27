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

package updater

import (
	"context"
	"fmt"

	"golang.org/x/exp/maps"

	"go.elara.ws/vercmp"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/search"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/overrides"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type Searcher interface {
	Search(ctx context.Context, opts *search.SearchOptions) ([]staplerfile.Package, error)
}

type Updater struct {
	mgr      manager.Manager
	info     *distro.OSRelease
	searcher Searcher
}

type UpdateInfo struct {
	Package *staplerfile.Package

	FromVersion string
	ToVersion   string
}

func New(mgr manager.Manager, info *distro.OSRelease, searcher Searcher) *Updater {
	return &Updater{mgr, info, searcher}
}

func (u *Updater) CheckForUpdates(
	ctx context.Context,
) ([]UpdateInfo, error) {
	installed, err := u.mgr.ListInstalled(nil)
	if err != nil {
		return nil, err
	}

	pkgNames := maps.Keys(installed)

	var out []UpdateInfo
	for _, pkgName := range pkgNames {
		matches := build.RegexpALRPackageName.FindStringSubmatch(pkgName)
		if matches != nil {
			packageName := matches[build.RegexpALRPackageName.SubexpIndex("package")]
			repoName := matches[build.RegexpALRPackageName.SubexpIndex("repo")]

			pkgs, err := u.searcher.Search(
				ctx,
				search.NewSearchOptions().
					WithName(packageName).
					WithRepository(repoName).
					Build(),
			)
			if err != nil {
				return nil, err
			}

			if len(pkgs) == 0 {
				continue
			}

			pkg := pkgs[0]

			version := u.getRepoVer(&pkg)
			c := vercmp.Compare(version, installed[pkgName])
			if c == 1 {
				out = append(out, UpdateInfo{
					Package:     &pkg,
					FromVersion: installed[pkgName],
					ToVersion:   version,
				})
			}
		}

	}

	return out, nil
}

func (u *Updater) getRepoVer(pkg *staplerfile.Package) string {
	repoVer := pkg.Version
	releaseStr := overrides.ReleasePlatformSpecific(pkg.Release, u.info)

	if pkg.Release != 0 && pkg.Epoch == 0 {
		repoVer = fmt.Sprintf("%s-%s", pkg.Version, releaseStr)
	} else if pkg.Release != 0 && pkg.Epoch != 0 {
		repoVer = fmt.Sprintf("%d:%s-%s", pkg.Epoch, pkg.Version, releaseStr)
	}

	return repoVer
}
