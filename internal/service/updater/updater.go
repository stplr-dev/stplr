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
	"log/slog"
	"slices"

	"github.com/gobwas/glob"
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

type Manager interface {
	ListInstalled(*manager.Opts) (map[string]string, error)
}

type IgnoreUpdatesProvider interface {
	// returns glob with repo/pkg pattern
	IgnorePkgUpdates() []string
}

type Updater struct {
	cfg      IgnoreUpdatesProvider
	mgr      Manager
	info     *distro.OSRelease
	searcher Searcher
}

type UpdateInfo struct {
	Package *staplerfile.Package

	FromVersion string
	ToVersion   string
}

func New(cfg IgnoreUpdatesProvider, mgr Manager, info *distro.OSRelease, searcher Searcher) *Updater {
	return &Updater{cfg, mgr, info, searcher}
}

func (u *Updater) CheckForUpdates(
	ctx context.Context,
) ([]UpdateInfo, error) {
	installed, err := u.mgr.ListInstalled(nil)
	if err != nil {
		return nil, err
	}

	pkgNames := maps.Keys(installed)
	slices.Sort(pkgNames)

	var out []UpdateInfo

	for _, pkgName := range pkgNames {
		updateInfo, err := u.checkPackageUpdate(ctx, pkgName, installed)
		if err != nil {
			return nil, err
		}
		if updateInfo != nil {
			out = append(out, *updateInfo)
		}
	}

	return out, nil
}

func (u *Updater) checkPackageUpdate(
	ctx context.Context,
	pkgName string,
	installed map[string]string,
) (*UpdateInfo, error) {
	matches := build.RegexpALRPackageName.FindStringSubmatch(pkgName)
	if matches == nil {
		return nil, nil
	}

	packageName := matches[build.RegexpALRPackageName.SubexpIndex("package")]
	repoName := matches[build.RegexpALRPackageName.SubexpIndex("repo")]

	pkg, err := u.findPackage(ctx, packageName, repoName)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		return nil, nil
	}

	if u.shouldIgnorePackage(pkg) {
		return nil, nil
	}

	return u.buildUpdateInfo(pkg, installed[pkgName])
}

func (u *Updater) findPackage(
	ctx context.Context,
	packageName string,
	repoName string,
) (*staplerfile.Package, error) {
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
		return nil, nil
	}

	return &pkgs[0], nil
}

func (u *Updater) shouldIgnorePackage(pkg *staplerfile.Package) bool {
	fullName := pkg.FormatFullName()

	for _, pattern := range u.cfg.IgnorePkgUpdates() {
		matched, err := patternMatch(fullName, pattern)
		if err != nil {
			slog.Debug("failed to match pattern", "err", err)
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

func (u *Updater) buildUpdateInfo(pkg *staplerfile.Package, installedVersion string) (*UpdateInfo, error) {
	version := u.getRepoVer(pkg)

	if vercmp.Compare(version, installedVersion) != 1 {
		return nil, nil
	}

	return &UpdateInfo{
		Package:     pkg,
		FromVersion: installedVersion,
		ToVersion:   version,
	}, nil
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

func patternMatch(fullName, pattern string) (bool, error) {
	g, err := glob.Compile(pattern)
	if err != nil {
		return false, err
	}
	return g.Match(fullName), nil
}
