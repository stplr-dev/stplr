// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
// Copyright (C) 2025 The ALR Authors
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

package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"
	"go.elara.ws/vercmp"
	"golang.org/x/exp/maps"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	database "go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/internal/search"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/distro"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

func prepareInstallerAndScripter() (installer build.InstallerExecutor, scripter build.ScriptExecutor, cleanup func(), err error) {
	var installerClose func()
	var scripterClose func()

	installer, installerClose, err = build.GetSafeInstaller()
	if err != nil {
		return nil, nil, nil, err
	}

	if utils.IsRoot() {
		if err := utils.ExitIfCantDropCapsToBuilderUserNoPrivs(); err != nil {
			installerClose()
			return nil, nil, nil, err
		}
	}

	scripter, scripterClose, err = build.GetSafeScriptExecutor()
	if err != nil {
		installerClose()
		return nil, nil, nil, err
	}

	cleanup = func() {
		scripterClose()
		installerClose()
	}

	return installer, scripter, cleanup, nil
}

func UpgradeCmd() *cli.Command {
	return &cli.Command{
		Name:    "upgrade",
		Usage:   gotext.Get("Upgrade all installed packages"),
		Aliases: []string{"up"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   gotext.Get("Build package from scratch even if there's an already built package available"),
			},
		},
		Action: utils.RootNeededAction(func(c *cli.Context) error {
			installer, scripter, cleanup, err := prepareInstallerAndScripter()
			if err != nil {
				return err
			}
			defer cleanup()

			ctx := c.Context

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithRepos().
				WithDistroInfo().
				WithManager().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			builder, err := build.NewMainBuilder(
				deps.Cfg,
				deps.Manager,
				deps.Repos,
				scripter,
				installer,
			)
			if err != nil {
				return err
			}

			updates, err := checkForUpdates(ctx, deps.Manager, deps.DB, deps.Info)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error checking for updates"), err)
			}

			if len(updates) > 0 {
				err = builder.InstallALRPackages(
					ctx,
					&build.BuildArgs{
						Opts: &types.BuildOpts{
							Clean:       c.Bool("clean"),
							Interactive: c.Bool("interactive"),
						},
						Info:       deps.Info,
						PkgFormat_: build.GetPkgFormat(deps.Manager),
					},
					mapUptatesInfoToPackages(updates),
				)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error checking for updates"), err)
				}
			} else {
				slog.Info(gotext.Get("There is nothing to do."))
			}

			return nil
		}),
	}
}

func mapUptatesInfoToPackages(updates []UpdateInfo) []alrsh.Package {
	var pkgs []alrsh.Package
	for _, info := range updates {
		pkgs = append(pkgs, *info.Package)
	}
	return pkgs
}

type UpdateInfo struct {
	Package *alrsh.Package

	FromVersion string
	ToVersion   string
}

func checkForUpdates(
	ctx context.Context,
	mgr manager.Manager,
	db *database.Database,
	info *distro.OSRelease,
) ([]UpdateInfo, error) {
	installed, err := mgr.ListInstalled(nil)
	if err != nil {
		return nil, err
	}

	pkgNames := maps.Keys(installed)

	s := search.New(db)

	var out []UpdateInfo
	for _, pkgName := range pkgNames {
		matches := build.RegexpALRPackageName.FindStringSubmatch(pkgName)
		if matches != nil {
			packageName := matches[build.RegexpALRPackageName.SubexpIndex("package")]
			repoName := matches[build.RegexpALRPackageName.SubexpIndex("repo")]

			pkgs, err := s.Search(
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

			repoVer := pkg.Version
			releaseStr := overrides.ReleasePlatformSpecific(pkg.Release, info)

			if pkg.Release != 0 && pkg.Epoch == 0 {
				repoVer = fmt.Sprintf("%s-%s", pkg.Version, releaseStr)
			} else if pkg.Release != 0 && pkg.Epoch != 0 {
				repoVer = fmt.Sprintf("%d:%s-%s", pkg.Epoch, pkg.Version, releaseStr)
			}

			c := vercmp.Compare(repoVer, installed[pkgName])

			if c == 1 {
				out = append(out, UpdateInfo{
					Package:     &pkg,
					FromVersion: installed[pkgName],
					ToVersion:   repoVer,
				})
			}
		}

	}

	return out, nil
}
