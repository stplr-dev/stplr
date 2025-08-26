// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
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
	"os"
	"text/template"

	"github.com/jeandeaual/go-locale"
	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/internal/search"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/distro"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
)

func SearchCmd() *cli.Command {
	return &cli.Command{
		Name:    "search",
		Usage:   gotext.Get("Search packages"),
		Aliases: []string{"s"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   gotext.Get("Show all information, not just for the current distro"),
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   gotext.Get("Search by name"),
			},
			&cli.StringFlag{
				Name:    "description",
				Aliases: []string{"d"},
				Usage:   gotext.Get("Search by description"),
			},
			&cli.StringFlag{
				Name:    "repository",
				Aliases: []string{"repo"},
				Usage:   gotext.Get("Search by repository"),
			},
			&cli.StringFlag{
				Name:    "provides",
				Aliases: []string{"p"},
				Usage:   gotext.Get("Search by provides"),
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   gotext.Get("Format output using a Go template"),
			},
		},
		Action: utils.ReadonlyAction(func(c *cli.Context) error {
			ctx := c.Context

			var names []string
			all := c.Bool("all")

			systemLang, err := locale.GetLanguage()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Can't detect system language"), err)
			}
			if systemLang == "" {
				systemLang = "en"
			}
			if !all {
				info, err := distro.ParseOSRelease(ctx)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error parsing os-release file"), err)
				}
				names, err = overrides.Resolve(
					info,
					overrides.DefaultOpts.
						WithLanguages([]string{systemLang}),
				)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error resolving overrides"), err)
				}
			}

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			database := deps.DB

			s := search.New(database)

			packages, err := s.Search(
				ctx,
				search.NewSearchOptions().
					WithName(c.String("name")).
					WithDescription(c.String("description")).
					WithRepository(c.String("repository")).
					WithProvides(c.String("provides")).
					Build(),
			)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error while executing search"), err)
			}

			format := c.String("format")
			var tmpl *template.Template
			if format != "" {
				tmpl, err = template.New("format").Parse(format)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Error parsing format template"), err)
				}
			}

			for _, pkg := range packages {
				alrsh.ResolvePackage(&pkg, names)
				if tmpl != nil {
					err = tmpl.Execute(os.Stdout, &pkg)
					if err != nil {
						return cliutils.FormatCliExit(gotext.Get("Error executing template"), err)
					}
					fmt.Println()
				} else {
					fmt.Println(pkg.Name)
				}
			}

			return nil
		}),
	}
}
