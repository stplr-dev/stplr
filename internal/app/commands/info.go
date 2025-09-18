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
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/usecase/info"
	"go.stplr.dev/stplr/internal/usecase/info/shell"
	"go.stplr.dev/stplr/internal/utils"
)

func InfoCmd() *cli.Command {
	return &cli.Command{
		Name:  "info",
		Usage: gotext.Get("Print information about a package"),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   gotext.Get("Show all information, not just for the current distro"),
			},
		},
		ShellComplete: cliutils.BashCompleteWithError(func(ctx context.Context, c *cli.Command) error {
			if err := utils.ExitIfRootCantDropCapsNoPrivs(); err != nil {
				return err
			}

			d, f, err := deps.ForInfoShellComp(ctx)
			if err != nil {
				return err
			}
			defer f()

			return shell.New(d.DB).Run(ctx)
		}),
		Action: func(ctx context.Context, c *cli.Command) error {
			args := c.Args()
			if args.Len() < 1 {
				return cli.Exit(gotext.Get("Command info expected at least 1 argument, got %d", args.Len()), 1)
			}

			if err := utils.ExitIfRootCantDropCapsNoPrivs(); err != nil {
				return err
			}

			d, f, err := deps.ForInfoAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			return info.New(d.Repos, d.Info).Run(ctx, info.Options{
				All:         c.Bool("all"),
				Pkgs:        c.Args().Slice(),
				Interactive: c.Bool("interactive"),
			})
		},
	}
}
