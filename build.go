// This file was originally part of the project "LURE - Linux User REpository", created by Elara Musayelyan.
// It has been modified as part of "ALR - Any Linux Repository" by Евгений Храмов.
//
// ALR - Any Linux Repository
// Copyright (C) 2025 Евгений Храмов
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
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"plemya-x.ru/alr/internal/config"
	"plemya-x.ru/alr/internal/osutils"
	"plemya-x.ru/alr/internal/types"
	"plemya-x.ru/alr/pkg/build"
	"plemya-x.ru/alr/pkg/loggerctx"
	"plemya-x.ru/alr/pkg/manager"
	"plemya-x.ru/alr/pkg/repos"
)

var buildCmd = &cli.Command{
	Name:  "build",
	Usage: "Build a local package",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "script",
			Aliases: []string{"s"},
			Value:   "alr.sh",
			Usage:   "Path to the build script",
		},
		&cli.StringFlag{
			Name:    "package",
			Aliases: []string{"p"},
			Usage:   "Name of the package to build and its repo (example: default/go-bin)",
		},
		&cli.BoolFlag{
			Name:    "clean",
			Aliases: []string{"c"},
			Usage:   "Build package from scratch even if there's an already built package available",
		},
	},
	Action: func(c *cli.Context) error {
		ctx := c.Context
		log := loggerctx.From(ctx)

		script := c.String("script")
		if c.String("package") != "" {
			script = filepath.Join(config.GetPaths(ctx).RepoDir, c.String("package"), "alr.sh")
		}

		err := repos.Pull(ctx, config.Config(ctx).Repos)
		if err != nil {
			log.Fatal("Error pulling repositories").Err(err).Send()
		}

		mgr := manager.Detect()
		if mgr == nil {
			log.Fatal("Unable to detect a supported package manager on the system").Send()
		}

		pkgPaths, _, err := build.BuildPackage(ctx, types.BuildOpts{
			Script:      script,
			Manager:     mgr,
			Clean:       c.Bool("clean"),
			Interactive: c.Bool("interactive"),
		})
		if err != nil {
			log.Fatal("Error building package").Err(err).Send()
		}

		wd, err := os.Getwd()
		if err != nil {
			log.Fatal("Error getting working directory").Err(err).Send()
		}

		for _, pkgPath := range pkgPaths {
			name := filepath.Base(pkgPath)
			err = osutils.Move(pkgPath, filepath.Join(wd, name))
			if err != nil {
				log.Fatal("Error moving the package").Err(err).Send()
			}
		}

		return nil
	},
}
