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

package commands

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"text/template"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/internal/utils"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
)

func ListCmd() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Usage:   gotext.Get("List Stapler repo packages"),
		Aliases: []string{"ls"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "installed",
				Aliases: []string{"I"},
			},
			&cli.BoolFlag{
				Name:    "upgradable",
				Aliases: []string{"U"},
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   gotext.Get("Format output using a Go template"),
			},
		},
		Action: utils.ReadonlyAction(func(ctx context.Context, c *cli.Command) error {
			if err := utils.ExitIfRootCantDropCapsNoPrivs(); err != nil {
				return err
			}

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithManager().
				WithDistroInfo().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			cfg := deps.Cfg
			db := deps.DB
			mgr := deps.Manager
			info := deps.Info

			if c.Bool("upgradable") {
				updates, err := checkForUpdates(ctx, mgr, db, info)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error getting packages for upgrade"), err)
				}
				if len(updates) == 0 {
					slog.Info(gotext.Get("No packages for upgrade"))
					return nil
				}

				format := c.String("format")
				if format == "" {
					format = "{{.Package.Repository}}/{{.Package.Name}} {{.FromVersion}} -> {{.ToVersion}}\n"
				}
				tmpl, err := template.New("format").Parse(format)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error parsing format template"), err)
				}

				for _, updateInfo := range updates {
					err = tmpl.Execute(os.Stdout, updateInfo)
					if err != nil {
						return cliutils.FormatCliExit(gotext.Get("Error executing template"), err)
					}
				}

				return nil
			}

			// TODO: refactor code below

			where := "true"
			args := []any(nil)
			if c.NArg() > 0 {
				where = "name LIKE ? OR json_array_contains(provides, ?)"
				args = []any{c.Args().First(), c.Args().First()}
			}

			result, err := db.GetPkgs(ctx, where, args...)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error getting packages"), err)
			}

			type verInfo struct {
				Version string
				Release int
			}

			installedAlrPackages := map[string]verInfo{}
			if c.Bool("installed") {
				mgr := manager.Detect()
				if mgr == nil {
					return cli.Exit(gotext.Get("Unable to detect a supported package manager on the system"), 1)
				}

				installed, err := mgr.ListInstalled(&manager.Opts{})
				if err != nil {
					slog.Error(gotext.Get("Error listing installed packages"), "err", err)
					return cli.Exit(err, 1)
				}

				for pkgName, version := range installed {
					matches := build.RegexpALRPackageName.FindStringSubmatch(pkgName)
					if matches != nil {
						packageName := matches[build.RegexpALRPackageName.SubexpIndex("package")]
						repoName := matches[build.RegexpALRPackageName.SubexpIndex("repo")]

						verInfo := verInfo{
							Version: version,
							Release: 0,
						}

						if i := strings.LastIndex(version, "-"); i != -1 {
							verInfo.Version = version[:i]
							verInfo.Release, err = overrides.ParseReleasePlatformSpecific(version[i+1:], info)
							if err != nil {
								slog.Error(gotext.Get("Failed to parse release"), "err", err)
								return cli.Exit(err, 1)
							}
						}

						installedAlrPackages[fmt.Sprintf("%s/%s", repoName, packageName)] = verInfo
					}
				}
			}

			for _, pkg := range result {
				if slices.Contains(cfg.IgnorePkgUpdates(), pkg.Name) {
					continue
				}

				type packageInfo struct {
					Package *alrsh.Package
				}

				pkgInfo := &packageInfo{}
				pkgInfo.Package = &pkg
				if c.Bool("installed") {
					instVersion, ok := installedAlrPackages[fmt.Sprintf("%s/%s", pkg.Repository, pkg.Name)]
					if !ok {
						continue
					} else {
						pkg.Version = instVersion.Version
						pkg.Release = instVersion.Release
					}
				}

				format := c.String("format")
				if format == "" {
					format = "{{.Package.Repository}}/{{.Package.Name}} {{.Package.Version}}-{{.Package.Release}}\n"
				}
				tmpl, err := template.New("format").Parse(format)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error parsing format template"), err)
				}
				err = tmpl.Execute(os.Stdout, pkgInfo)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error executing template"), err)
				}
			}

			return nil
		}),
	}
}
