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

package commands

import (
	"context"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/cliutils2"
	"go.stplr.dev/stplr/internal/usecase/search"
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
				Name:    "query",
				Aliases: []string{"q"},
				Usage:   gotext.Get("Search by query (CEL syntax)"),
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   gotext.Get("[DEPRECATED] Search by name"),
			},
			&cli.StringFlag{
				Name:    "description",
				Aliases: []string{"d"},
				Usage:   gotext.Get("[DEPRECATED] Search by description"),
			},
			&cli.StringFlag{
				Name:    "repository",
				Aliases: []string{"repo"},
				Usage:   gotext.Get("[DEPRECATED] Search by repository"),
			},
			&cli.StringFlag{
				Name:    "provides",
				Aliases: []string{"p"},
				Usage:   gotext.Get("[DEPRECATED] Search by provides"),
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   gotext.Get("Format output using a Go template"),
			},
		},
		Action: cliutils2.ReadonlyAction(func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForSearchAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			packageName := ""
			if c.Args().Len() > 0 {
				packageName = c.Args().First()
			}

			return search.New(d.Searcher, d.Info).Run(ctx, search.Options{
				Name:        c.String("name"),
				Description: c.String("description"),
				Repository:  c.String("repository"),
				Provides:    c.String("provides"),
				Format:      c.String("format"),
				PackageName: packageName,
				Query:       c.String("query"),
				All:         c.Bool("all"),
			})
		}),
	}
}
