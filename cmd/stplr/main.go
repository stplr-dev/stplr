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

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/i18n"

	"go.stplr.dev/stplr/internal/logger"
)

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

	p := app.GetApp()
	cliutils.Localize(p)

	cfg := config.New()
	err := cfg.Load()
	if err != nil {
		out.Error("%s: %v", gotext.Get("Error loading config"), err)
		os.Exit(1)
	}
	setLogLevel(cfg.LogLevel())

	err = p.Run(ctx, os.Args)
	if err != nil {
		out.Error("%s: %v", gotext.Get("Error while running app"), err)
	}
}
