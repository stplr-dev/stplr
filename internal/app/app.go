// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package app

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/commands"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/manager"
)

func GetApp() *cli.Command {
	cmds := []*cli.Command{
		commands.InstallCmd(),
		commands.RemoveCmd(),
		commands.UpgradeCmd(),
		commands.InfoCmd(),
		commands.ListCmd(),
		commands.BuildCmd(),
		commands.RefreshCmd(),
		commands.FixCmd(),
		commands.HelperCmd(),
		commands.VersionCmd(),
		commands.SearchCmd(),
		commands.RepoCmd(),
		commands.ConfigCmd(),
		commands.MigrateCmd(),
		commands.SupportCmd(),
		// Internal commands
		commands.InternalPluginProvider(),
		commands.InternalPluginProviderRoot(),
	}

	return &cli.Command{
		Name:  "stplr",
		Usage: gotext.Get("Command-line interface for Stapler, a universal Linux package build system"),
		Description: gotext.Get("Stapler is a universal Linux package build and distribution system designed for cross-distribution software delivery.\n\n" +
			"Packages are distributed through Stapler repositories. The application ships without any repositories configured, " +
			"so you need to add one or more repositories before installing software.\n\n" +
			"Getting started:\n" +
			"  stplr repo add [name] [url]    Add a Stapler repository\n" +
			"  stplr install [package]        Install a package from configured repositories\n" +
			"  stplr search [query]           Search for packages\n\n" +
			"Learn more (including community repositories): https://stplr.dev/docs/intro"),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "pm-args",
				Aliases: []string{"P"},
				Usage:   gotext.Get("Arguments to be passed on to the package manager"),
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Value:   isatty.IsTerminal(os.Stdin.Fd()),
				Usage:   gotext.Get("Enable interactive questions and prompts"),
			},
		},
		Commands: cmds,
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			slog.Debug("cli started", "args", os.Args)
			if trimmed := strings.TrimSpace(c.String("pm-args")); trimmed != "" {
				args := strings.Split(trimmed, " ")
				manager.Args = append(manager.Args, args...)
			}
			return ctx, nil
		},
		EnableShellCompletion: true,
		ExitErrHandler: func(ctx context.Context, c *cli.Command, err error) {
			cliutils.HandleExitCoder(ctx, c, err)
		},
	}
}
