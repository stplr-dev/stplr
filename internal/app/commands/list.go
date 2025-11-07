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

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/cliutils2"
	"go.stplr.dev/stplr/internal/usecase/list"
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
		Action: cliutils2.ReadonlyAction(func(ctx context.Context, c *cli.Command) error {
			if err := cliutils.ExitIfRootCantDropCapsNoPrivs(); err != nil {
				return err
			}

			d, f, err := deps.ForListAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			return list.New(d.Updater, d.DB, d.Config, d.Info, output.FromContext(ctx)).Run(ctx, list.Options{
				Upgradable: c.Bool("upgradable"),
				Installed:  c.Bool("installed"),
				Format:     c.String("format"),
			})
		}),
	}
}
