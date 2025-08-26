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
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/osutils"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

func BuildCmd() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: gotext.Get("Build a local package"),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "script",
				Aliases: []string{"s"},
				Value:   "Staplerfile",
				Usage:   gotext.Get("Path to the build script"),
			},
			&cli.StringFlag{
				Name:    "subpackage",
				Aliases: []string{"sb"},
				Usage:   gotext.Get("Specify subpackage in script (for multi package script only)"),
			},
			&cli.StringFlag{
				Name:    "package",
				Aliases: []string{"p"},
				Usage:   gotext.Get("Name of the package to build and its repo (example: default/go-bin)"),
			},
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   gotext.Get("Build package from scratch even if there's an already built package available"),
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithReposNoPull().
				WithDistroInfo().
				WithManager().
				Build()
			if err != nil {
				return cli.Exit(err, 1)
			}
			defer deps.Defer()

			if deps.Cfg.ForbidBuildCommand() {
				return cliutils.FormatCliExit(gotext.Get("Your settings do not allow build command"), nil)
			}

			if err := utils.EnsureIsPrivilegedGroupMemberOrRoot(); err != nil {
				return err
			}

			wd, err := os.Getwd()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error getting working directory"), err)
			}

			origUid, origGid := syscall.Getuid(), syscall.Getgid()

			userMode := utils.IsNotRoot()

			var script string
			var packages []string

			var res []*build.BuiltDep

			var scriptArgs *build.BuildPackageFromScriptArgs
			var dbArgs *build.BuildPackageFromDbArgs

			buildArgs := &build.BuildArgs{
				Opts: &types.BuildOpts{
					Clean:       c.Bool("clean"),
					Interactive: c.Bool("interactive"),
				},
				PkgFormat_: build.GetPkgFormat(deps.Manager),
				Info:       deps.Info,
			}

			switch {
			case c.IsSet("script"):
				script, err = filepath.Abs(c.String("script"))
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Cannot get absolute script path"), err)
				}

				subpackage := c.String("subpackage")

				if subpackage != "" {
					packages = append(packages, subpackage)
				}

				scriptArgs = &build.BuildPackageFromScriptArgs{
					Script:    script,
					Packages:  packages,
					BuildArgs: *buildArgs,
				}
			case c.IsSet("package"):
				// TODO: handle multiple packages
				packageInput := c.String("package")

				pkgs, _, err := deps.Repos.FindPkgs(ctx, []string{packageInput})
				if err != nil {
					return cliutils.FormatCliExit("failed to find pkgs", err)
				}

				pkg := cliutils.FlattenPkgs(ctx, pkgs, "build", c.Bool("interactive"))

				if len(pkg) < 1 {
					return cliutils.FormatCliExit(gotext.Get("Package not found"), nil)
				}

				if pkg[0].BasePkgName != "" {
					packages = append(packages, pkg[0].Name)
				}

				dbArgs = &build.BuildPackageFromDbArgs{
					Package:   &pkg[0],
					Packages:  packages,
					BuildArgs: *buildArgs,
				}
			default:
				return cliutils.FormatCliExit(gotext.Get("Nothing to build"), nil)
			}

			var file *staplerfile.ScriptFile

			if scriptArgs != nil {
				file, err = staplerfile.ReadFromLocal(scriptArgs.Script)
				if err != nil {
					return err
				}
			}

			needCopier := utils.IsRoot()

			var copier build.ScriptCopier
			var cleanup func()

			if needCopier {
				copier, cleanup, err = build.GetSafeScriptCopier()
				if err != nil {
					return cliutils.FormatCliExit("failed to init copier", err)
				}
				defer cleanup()

				if scriptArgs != nil {
					scriptArgs.Script, err = copier.Copy(ctx, file, deps.Info)
					if err != nil {
						cleanup()
						return err
					}
					defer func() {
						err = os.RemoveAll(filepath.Dir(scriptArgs.Script))
						if err != nil {
							panic(err)
						}
					}()
				}
			}

			installer, scripter, cleanup, err := prepareInstallerAndScripter()
			if err != nil {
				return err
			}
			defer cleanup()

			if userMode {
				paths := deps.Cfg.GetPaths()

				userCacheDir, err := os.UserCacheDir()
				if err != nil {
					return err
				}

				paths.CacheDir = filepath.Join(userCacheDir, "stplr")
				paths.PkgsDir = filepath.Join(paths.CacheDir, "pkgs")
			}

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

			if scriptArgs != nil {
				res, err = builder.BuildPackageFromScript(
					ctx,
					scriptArgs,
				)
			} else if dbArgs != nil {
				res, err = builder.BuildPackageFromDb(
					ctx,
					dbArgs,
				)
			}

			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error building package"), err)
			}

			for _, pkg := range res {
				name := filepath.Base(pkg.Path)
				if needCopier {
					if err = copier.CopyOut(ctx, pkg.Path, filepath.Join(wd, name), origUid, origGid); err != nil {
						return cliutils.FormatCliExit(gotext.Get("Error moving the package"), err)
					}
				} else {
					if err = osutils.CopyFile(pkg.Path, filepath.Join(wd, name)); err != nil {
						return cliutils.FormatCliExit(gotext.Get("Error moving the package"), err)
					}
				}

			}

			slog.Info(gotext.Get("Done"))

			return nil
		},
	}
}
