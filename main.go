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
	"os/signal"
	"strings"
	"syscall"

	"github.com/leonelquinteros/gotext"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/translations"

	"go.stplr.dev/stplr/internal/logger"
)

func VersionCmd() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: gotext.Get("Print the current Stapler version and exit"),
		Action: func(ctx *cli.Context) error {
			println(config.Version)
			return nil
		},
	}
}

func GetApp() *cli.App {
	return &cli.App{
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
		Commands: []*cli.Command{
			InstallCmd(),
			RemoveCmd(),
			UpgradeCmd(),
			InfoCmd(),
			ListCmd(),
			BuildCmd(),
			LegacyAddRepoCmd(),
			LegacyRemoveRepoCmd(),
			RefreshCmd(),
			FixCmd(),
			HelperCmd(),
			VersionCmd(),
			SearchCmd(),
			RepoCmd(),
			ConfigCmd(),
			// Internal commands
			InternalBuildCmd(),
			InternalInstallCmd(),
			InternalReposCmd(),
			InternalCoplyFiles(),
		},
		Before: func(c *cli.Context) error {
			if trimmed := strings.TrimSpace(c.String("pm-args")); trimmed != "" {
				args := strings.Split(trimmed, " ")
				manager.Args = append(manager.Args, args...)
			}
			return nil
		},
		EnableBashCompletion: true,
		ExitErrHandler: func(cCtx *cli.Context, err error) {
			cliutils.HandleExitCoder(err)
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
	logger, ok := slog.Default().Handler().(*logger.Logger)
	if !ok {
		panic("unexpected")
	}
	logger.SetLevel(level)
}

func main() {
	logger.SetupDefault()
	setLogLevel(os.Getenv("STPLR_LOG_LEVEL"))
	translations.Setup()

	ctx := context.Background()

	app := GetApp()
	cfg := config.New()
	err := cfg.Load()
	if err != nil {
		slog.Error(gotext.Get("Error loading config"), "err", err)
		os.Exit(1)
	}
	setLogLevel(cfg.LogLevel())

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cliutils.Localize(app)

	err = app.RunContext(ctx, os.Args)
	if err != nil {
		slog.Error(gotext.Get("Error while running app"), "err", err)
	}
}
