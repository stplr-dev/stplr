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
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/commands"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/i18n"
	"go.stplr.dev/stplr/internal/manager"

	"go.stplr.dev/stplr/internal/logger"
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

func setLogLevel(newLevel string) {
	level := slog.LevelInfo
	switch newLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	}
	logger.Level.Set(level)
}

func main() {
	logger.SetupDefault()
	setLogLevel(os.Getenv("STPLR_LOG_LEVEL"))
	i18n.Setup()

	ctx := context.Background()

	out := output.NewConsoleOutput()
	ctx = output.WithOutput(ctx, out)

	app := GetApp()
	cliutils.Localize(app)

	cfg := config.New()
	err := cfg.Load()
	if err != nil {
		out.Error("%s: %v", gotext.Get("Error loading config"), err)
		os.Exit(1)
	}
	setLogLevel(cfg.LogLevel())

	err = app.Run(ctx, os.Args)
	if err != nil {
		out.Error("%s: %v", gotext.Get("Error while running app"), err)
	}
}
