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
	"fmt"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/types"
)

func InstallCmd() *cli.Command {
	return &cli.Command{
		Name:    "install",
		Usage:   gotext.Get("Install a new package"),
		Aliases: []string{"in"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   gotext.Get("Build package from scratch even if there's an already built package available"),
			},
		},
		Action: utils.RootNeededAction(func(c *cli.Context) error {
			args := c.Args()
			if args.Len() < 1 {
				return cliutils.FormatCliExit(gotext.Get("Command install expected at least 1 argument, got %d", args.Len()), nil)
			}

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

			_, err = builder.InstallPkgs(
				ctx,
				&build.BuildArgs{
					Opts: &types.BuildOpts{
						Clean:       c.Bool("clean"),
						Interactive: c.Bool("interactive"),
					},
					Info:       deps.Info,
					PkgFormat_: build.GetPkgFormat(deps.Manager),
				},
				args.Slice(),
			)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error when installing the package"), err)
			}

			return nil
		}),
		BashComplete: cliutils.BashCompleteWithError(func(c *cli.Context) error {
			ctx := c.Context
			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			result, err := deps.DB.GetPkgs(c.Context, "true")
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error getting packages"), err)
			}

			for _, pkg := range result {
				fmt.Println(pkg.Name)
			}

			return nil
		}),
	}
}

func RemoveCmd() *cli.Command {
	return &cli.Command{
		Name:    "remove",
		Usage:   gotext.Get("Remove an installed package"),
		Aliases: []string{"rm"},
		BashComplete: cliutils.BashCompleteWithError(func(c *cli.Context) error {
			ctx := c.Context

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithManager().
				Build()
			if err != nil {
				return cli.Exit(err, 1)
			}
			defer deps.Defer()

			installedAlrPackages := map[string]string{}
			installed, err := deps.Manager.ListInstalled(&manager.Opts{})
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error listing installed packages"), err)
			}
			for pkgName, version := range installed {
				matches := build.RegexpALRPackageName.FindStringSubmatch(pkgName)
				if matches != nil {
					packageName := matches[build.RegexpALRPackageName.SubexpIndex("package")]
					repoName := matches[build.RegexpALRPackageName.SubexpIndex("repo")]
					installedAlrPackages[fmt.Sprintf("%s/%s", repoName, packageName)] = version
				}
			}

			result, err := deps.DB.GetPkgs(c.Context, "true")
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error getting packages"), err)
			}

			for _, pkg := range result {
				_, ok := installedAlrPackages[fmt.Sprintf("%s/%s", pkg.Repository, pkg.Name)]
				if !ok {
					continue
				}
				fmt.Println(pkg.Name)
			}

			return nil
		}),
		Action: utils.RootNeededAction(func(c *cli.Context) error {
			args := c.Args()
			if args.Len() < 1 {
				return cliutils.FormatCliExit(gotext.Get("Command remove expected at least 1 argument, got %d", args.Len()), nil)
			}

			deps, err := appbuilder.
				New(c.Context).
				WithManager().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			if err := deps.Manager.Remove(&manager.Opts{
				NoConfirm: !c.Bool("interactive"),
			}, c.Args().Slice()...); err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error removing packages"), err)
			}

			return nil
		}),
	}
}
