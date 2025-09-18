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

package list

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"text/template"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/internal/service/updater"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

const (
	defaultFormatTemplate = "{{.Package.Repository}}/{{.Package.Name}} {{.Package.Version}}-{{.Package.Release}}\n"
)

type Updater interface {
	CheckForUpdates(
		ctx context.Context,
	) ([]updater.UpdateInfo, error)
}

type PackagesProvider interface {
	GetPkgs(ctx context.Context, where string, args ...any) ([]staplerfile.Package, error)
}

type IgnorePkgProvider interface {
	IgnorePkgUpdates() []string
}

type useCase struct {
	upd            Updater
	pkgs           PackagesProvider
	ignoreProvider IgnorePkgProvider

	info *distro.OSRelease
}

func New(upd Updater, pkgs PackagesProvider, ignoreProvider IgnorePkgProvider, info *distro.OSRelease) *useCase {
	return &useCase{
		upd,
		pkgs,
		ignoreProvider,
		info,
	}
}

type Options struct {
	Upgradable bool
	Installed  bool
	Format     string
}

func (u *useCase) runForUpgradable(ctx context.Context, opts Options) error {
	updates, err := u.upd.CheckForUpdates(ctx)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error getting packages for upgrade"))
	}
	if len(updates) == 0 {
		slog.Info(gotext.Get("No packages for upgrade"))
		return nil
	}

	format := opts.Format
	if format == "" {
		format = "{{.Package.Repository}}/{{.Package.Name}} {{.FromVersion}} -> {{.ToVersion}}\n"
	}
	tmpl, err := template.New("format").Parse(format)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error parsing format template"))
	}

	for _, updateInfo := range updates {
		err = tmpl.Execute(os.Stdout, updateInfo)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error executing template"))
		}
	}

	return nil
}

func (u *useCase) getPackages(ctx context.Context) ([]staplerfile.Package, error) {
	return u.pkgs.GetPkgs(ctx, "true")
}

type VerInfo struct {
	Version string
	Release int
}

type PackageInfo struct {
	Package *staplerfile.Package
}

func (u *useCase) parseVersion(version string) (VerInfo, error) {
	verInfo := VerInfo{Version: version, Release: 0}
	if i := strings.LastIndex(version, "-"); i != -1 {
		verInfo.Version = version[:i]
		release, err := overrides.ParseReleasePlatformSpecific(version[i+1:], u.info)
		if err != nil {
			return VerInfo{}, err
		}
		verInfo.Release = release
	}
	return verInfo, nil
}

func (u *useCase) parseInstalledPackages(installed map[string]string) (map[string]VerInfo, error) {
	installedAlrPackages := make(map[string]VerInfo)

	for pkgName, version := range installed {
		matches := build.RegexpALRPackageName.FindStringSubmatch(pkgName)
		if matches == nil {
			continue
		}

		packageName := matches[build.RegexpALRPackageName.SubexpIndex("package")]
		repoName := matches[build.RegexpALRPackageName.SubexpIndex("repo")]
		key := fmt.Sprintf("%s/%s", repoName, packageName)

		verInfo, err := u.parseVersion(version)
		if err != nil {
			return nil, errors.WrapIntoI18nError(err, gotext.Get("Failed to parse release"))
		}

		installedAlrPackages[key] = verInfo
	}

	return installedAlrPackages, nil
}

func (u *useCase) getInstalledPackages(opts Options) (map[string]VerInfo, error) {
	if !opts.Installed {
		return nil, nil
	}

	mgr := manager.Detect()
	if mgr == nil {
		return nil, errors.NewI18nError(gotext.Get("Unable to detect a supported package manager on the system"))
	}

	installed, err := mgr.ListInstalled(&manager.Opts{})
	if err != nil {
		return nil, errors.WrapIntoI18nError(err, gotext.Get("Error listing installed packages"))
	}

	return u.parseInstalledPackages(installed)
}

func (u *useCase) processAndOutputPackages(opts Options, pkgs []staplerfile.Package, installedPkgs map[string]VerInfo) error {
	format := opts.Format
	if format == "" {
		format = defaultFormatTemplate
	}

	tmpl, err := template.New("format").Parse(format)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error parsing format template"))
	}

	for _, pkg := range pkgs {
		if slices.Contains(u.ignoreProvider.IgnorePkgUpdates(), pkg.Name) {
			continue
		}

		pkgInfo := &PackageInfo{Package: &pkg}
		if opts.Installed {
			if instVersion, ok := installedPkgs[fmt.Sprintf("%s/%s", pkg.Repository, pkg.Name)]; ok {
				pkg.Version = instVersion.Version
				pkg.Release = instVersion.Release
			} else {
				continue
			}
		}

		if err := tmpl.Execute(os.Stdout, pkgInfo); err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error executing template"))
		}
	}

	return nil
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	if opts.Upgradable {
		return u.runForUpgradable(ctx, opts)
	}

	pkgs, err := u.getPackages(ctx)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error getting packages"))
	}

	installedPkgs, err := u.getInstalledPackages(opts)
	if err != nil {
		return err
	}

	return u.processAndOutputPackages(opts, pkgs, installedPkgs)
}
