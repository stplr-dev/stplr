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
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/errors"
	build "go.stplr.dev/stplr/internal/usecase/build"
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
			&cli.BoolFlag{
				Name:  "no-suffix",
				Usage: gotext.Get("Do not add suffix to package name"),
			},
			&cli.StringSliceFlag{
				Name:  "source",
				Usage: gotext.Get("Override a source by index (format: N:path or N:url, e.g. 0:./foo.deb)"),
			},
			&cli.BoolFlag{
				Name:  "ignore-overriden-source-checksum",
				Usage: gotext.Get("Skip checksum verification for overridden sources"),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForBuildAction(ctx)
			if err != nil {
				return fmt.Errorf("failed to get BuildActionDeps: %w", err)
			}
			defer f()

			overrides, err := parseSourceOverrides(c.StringSlice("source"))
			if err != nil {
				return err
			}

			return build.New(build.ConstructOptions{
				Builder: d.Builder,
				Info:    d.Info,
				Copier:  d.Copier,
				Manager: d.Manager,
				Finder:  d.Repos,
				Config:  d.Config,
			}).Run(ctx, build.RunOptions{
				Script:               c.String("script"),
				Package:              c.String("package"),
				Subpackage:           c.String("subpackage"),
				Clean:                c.Bool("clean"),
				Interactive:          c.Bool("interactive"),
				NoSuffix:             c.Bool("no-suffix"),
				SourceOverrides:      overrides,
				IgnoreSourceChecksum: c.Bool("ignore-overriden-source-checksum"),
			})
		},
	}
}

// parseSourceOverrides parses --source flags of the form "N:path" or "N:url" into a map of index -> path/URL.
// Local paths are resolved to absolute paths; remote URLs (containing "://") are kept as-is.
func parseSourceOverrides(args []string) (map[int]string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	result := make(map[int]string, len(args))
	for _, s := range args {
		idx, value, ok := strings.Cut(s, ":")
		if !ok {
			return nil, errors.NewI18nError(gotext.Get("invalid --source format %q: expected N:path (e.g. 0:./file.deb)", s))
		}
		n, err := strconv.Atoi(idx)
		if err != nil || n < 0 {
			return nil, errors.NewI18nError(gotext.Get("invalid source index %q: must be a non-negative integer", idx))
		}
		if strings.Contains(value, "://") {
			result[n] = value
		} else {
			absPath, err := filepath.Abs(value)
			if err != nil {
				return nil, errors.WrapIntoI18nError(err, gotext.Get("failed to resolve path %q", value))
			}
			result[n] = "local:///" + absPath
		}
	}
	return result, nil
}
