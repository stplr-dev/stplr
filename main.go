/*
 * LURE - Linux User REpository
 * Copyright (C) 2023 Elara Musayelyan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
	"go.elara.ws/logger"
	"lure.sh/lure/internal/config"
	"lure.sh/lure/internal/db"
	"lure.sh/lure/internal/translations"
	"lure.sh/lure/pkg/loggerctx"
	"lure.sh/lure/pkg/manager"
)

var app = &cli.App{
	Name:  "lure",
	Usage: "Linux User REpository",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "pm-args",
			Aliases: []string{"P"},
			Usage:   "Arguments to be passed on to the package manager",
		},
		&cli.BoolFlag{
			Name:    "interactive",
			Aliases: []string{"i"},
			Value:   isatty.IsTerminal(os.Stdin.Fd()),
			Usage:   "Enable interactive questions and prompts",
		},
	},
	Commands: []*cli.Command{
		installCmd,
		removeCmd,
		upgradeCmd,
		infoCmd,
		listCmd,
		buildCmd,
		addrepoCmd,
		removerepoCmd,
		refreshCmd,
		fixCmd,
		genCmd,
		helperCmd,
		versionCmd,
	},
	Before: func(c *cli.Context) error {
		ctx := c.Context
		log := loggerctx.From(ctx)

		cmd := c.Args().First()
		if cmd != "helper" && !config.Config(ctx).Unsafe.AllowRunAsRoot && os.Geteuid() == 0 {
			log.Fatal("Running LURE as root is forbidden as it may cause catastrophic damage to your system").Send()
		}

		if trimmed := strings.TrimSpace(c.String("pm-args")); trimmed != "" {
			args := strings.Split(trimmed, " ")
			manager.Args = append(manager.Args, args...)
		}

		return nil
	},
	After: func(ctx *cli.Context) error {
		return db.Close()
	},
	EnableBashCompletion: true,
}

var versionCmd = &cli.Command{
	Name:  "version",
	Usage: "Print the current LURE version and exit",
	Action: func(ctx *cli.Context) error {
		println(config.Version)
		return nil
	},
}

func main() {
	ctx := context.Background()
	log := translations.NewLogger(ctx, logger.NewCLI(os.Stderr), config.Language(ctx))
	ctx = loggerctx.With(ctx, log)

	// Set the root command to the one set in the LURE config
	manager.DefaultRootCmd = config.Config(ctx).RootCmd

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Error("Error while running app").Err(err).Send()
	}
}
