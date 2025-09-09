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

	"go.stplr.dev/stplr/internal/services/builder"
	"go.stplr.dev/stplr/internal/sys"
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
		Action: func(ctx context.Context, c *cli.Command) error {
			return builder.
				New(sys.Sys{}).
				Run(ctx, builder.Options{
					Script:      c.String("script"),
					Subpackage:  c.String("subpackage"),
					Package:     c.String("package"),
					Clean:       c.Bool("clean"),
					Interactive: c.Bool("interactive"),
				})
		},
	}
}
