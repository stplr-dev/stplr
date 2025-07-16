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
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/utils"
)

func FixCmd() *cli.Command {
	return &cli.Command{
		Name:  "fix",
		Usage: gotext.Get("Attempt to fix problems with Stapler"),
		Action: func(c *cli.Context) error {
			if err := utils.ExitIfCantDropCapsToAlrUserNoPrivs(); err != nil {
				return err
			}

			ctx := c.Context

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				Build()
			if err != nil {
				return cli.Exit(err, 1)
			}
			defer deps.Defer()

			cfg := deps.Cfg

			paths := cfg.GetPaths()

			slog.Info(gotext.Get("Clearing cache directory"))

			dir, err := os.Open(paths.CacheDir)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Unable to open cache directory"), err)
			}
			defer dir.Close()

			entries, err := dir.Readdirnames(-1)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Unable to read cache directory contents"), err)
			}

			for _, entry := range entries {
				fullPath := filepath.Join(paths.CacheDir, entry)

				if err := makeWritableRecursive(fullPath); err != nil {
					slog.Debug("Failed to make path writable", "path", fullPath, "error", err)
				}

				err = os.RemoveAll(fullPath)
				if err != nil {
					return cliutils.FormatCliExit(gotext.Get("Unable to remove cache item (%s)", entry), err)
				}
			}

			slog.Info(gotext.Get("Rebuilding cache"))

			err = os.MkdirAll(paths.CacheDir, 0o755)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Unable to create new cache directory"), err)
			}

			deps, err = appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithReposForcePull().
				Build()
			if err != nil {
				return cli.Exit(err, 1)
			}
			defer deps.Defer()

			slog.Info(gotext.Get("Done"))

			return nil
		},
	}
}

func makeWritableRecursive(path string) error {
	return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		newMode := info.Mode() | 0o200
		if d.IsDir() {
			newMode |= 0o100
		}

		return os.Chmod(path, newMode)
	})
}
