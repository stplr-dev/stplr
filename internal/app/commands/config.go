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
	"fmt"
	"slices"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/cliutils2"
	"go.stplr.dev/stplr/internal/usecase/config/get"
	"go.stplr.dev/stplr/internal/usecase/config/set"
	"go.stplr.dev/stplr/internal/usecase/config/show"
)

func ConfigCmd() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: gotext.Get("Manage config"),
		Commands: []*cli.Command{
			ShowCmd(),
			SetConfig(),
			GetConfig(),
		},
	}
}

func ShowCmd() *cli.Command {
	return &cli.Command{
		Name:  "show",
		Usage: gotext.Get("Show config"),
		Action: func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForConfigShowAction(ctx)
			if err != nil {
				return err
			}
			defer f()
			return show.New(d.Config).Run(ctx)
		},
	}
}

var configKeys = []string{
	"rootCmd",
	"useRootCmd",
	"pagerStyle",
	"autoPull",
	"logLevel",
	"ignorePkgUpdates",
	"forbidSkipInChecksums",
	"forbidBuildCommand",
}

func SetConfig() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     gotext.Get("Set config value"),
		ArgsUsage: gotext.Get("<key> <value>"),
		ShellComplete: cliutils.BashCompleteWithError(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() == 0 {
				for _, key := range configKeys {
					fmt.Println(key)
				}
				return nil
			}
			return nil
		}),
		Action: cliutils2.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			field := c.Args().Get(0)

			repoKeys := []string{"repo", "repos"}
			if slices.Contains(repoKeys, field) {
				return errors.NewI18nError(gotext.Get("use 'repo add/remove' commands to manage repositories"))
			}

			d, f, err := deps.ForConfigSetAction(ctx)
			if err != nil {
				return err
			}
			defer f()
			return set.New(d.Config).Run(ctx, set.Options{
				Field: field,
				Value: c.Args().Get(1),
			})
		}),
	}
}

func GetConfig() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     gotext.Get("Get config value"),
		ArgsUsage: gotext.Get("<key>"),
		ShellComplete: cliutils.BashCompleteWithError(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() == 0 {
				for _, key := range configKeys {
					fmt.Println(key)
				}
				return nil
			}
			return nil
		}),
		Action: cliutils2.ReadonlyAction(func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForConfigGetAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			// If no key passed we fallback to show command
			if c.Args().Len() == 0 {
				return show.New(d.Config).Run(ctx)
			}

			key := c.Args().Get(0)
			return get.New(d.Config).Run(ctx, key)
		}),
	}
}
