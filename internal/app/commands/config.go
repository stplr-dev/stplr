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
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/utils"
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
		ShellComplete: cliutils.BashCompleteWithError(func(ctx context.Context, c *cli.Command) error {
			return nil
		}),
		Action: func(ctx context.Context, c *cli.Command) error {
			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			content, err := deps.Cfg.ToYAML()
			if err != nil {
				return err
			}
			fmt.Println(content)
			return nil
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
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			key := c.Args().Get(0)
			value := c.Args().Get(1)

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			stringSetters := map[string]func(string){
				"rootCmd":    deps.Cfg.System.SetRootCmd,
				"pagerStyle": deps.Cfg.System.SetPagerStyle,
				"logLevel":   deps.Cfg.System.SetLogLevel,
			}

			boolSetters := map[string]func(bool){
				"useRootCmd":            deps.Cfg.System.SetUseRootCmd,
				"autoPull":              deps.Cfg.System.SetAutoPull,
				"forbidSkipInChecksums": deps.Cfg.System.SetForbidSkipInChecksums,
				"forbidBuildCommand":    deps.Cfg.System.SetForbidBuildCommand,
			}

			switch key {
			case "ignorePkgUpdates":
				var updates []string
				if value != "" {
					updates = strings.Split(value, ",")
					for i := range updates {
						updates[i] = strings.TrimSpace(updates[i])
					}
				}
				deps.Cfg.System.SetIgnorePkgUpdates(updates)

			case "repo", "repos":
				return cliutils.FormatCliExit(
					gotext.Get("use 'repo add/remove' commands to manage repositories"), nil)

			default:
				if setter, ok := stringSetters[key]; ok {
					setter(value)
				} else if setter, ok := boolSetters[key]; ok {
					boolValue, err := strconv.ParseBool(value)
					if err != nil {
						return cliutils.FormatCliExit(
							gotext.Get("invalid boolean value for %s: %s", key, value), err)
					}
					setter(boolValue)
				} else {
					return cliutils.FormatCliExit(gotext.Get("unknown config key: %s", key), nil)
				}
			}

			if err := deps.Cfg.System.Save(); err != nil {
				return cliutils.FormatCliExit(gotext.Get("failed to save config"), err)
			}

			fmt.Println(gotext.Get("Successfully set %s = %s", key, value))
			return nil
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
		Action: utils.ReadonlyAction(func(ctx context.Context, c *cli.Command) error {
			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			if c.Args().Len() == 0 {
				content, err := deps.Cfg.ToYAML()
				if err != nil {
					return cliutils.FormatCliExit("failed to serialize config", err)
				}
				fmt.Print(content)
				return nil
			}

			key := c.Args().Get(0)

			stringGetters := map[string]func() string{
				"rootCmd":    deps.Cfg.RootCmd,
				"pagerStyle": deps.Cfg.PagerStyle,
				"logLevel":   deps.Cfg.LogLevel,
			}

			boolGetters := map[string]func() bool{
				"useRootCmd":            deps.Cfg.UseRootCmd,
				"autoPull":              deps.Cfg.AutoPull,
				"forbidSkipInChecksums": deps.Cfg.ForbidSkipInChecksums,
				"forbidBuildCommand":    deps.Cfg.ForbidBuildCommand,
			}

			switch key {
			case "ignorePkgUpdates":
				updates := deps.Cfg.IgnorePkgUpdates()
				if len(updates) == 0 {
					fmt.Println("[]")
				} else {
					fmt.Println(strings.Join(updates, ", "))
				}
			case "repo", "repos":
				repos := deps.Cfg.Repos()
				if len(repos) == 0 {
					fmt.Println("[]")
				} else {
					repoData, err := yaml.Marshal(repos)
					if err != nil {
						return cliutils.FormatCliExit("failed to serialize repos", err)
					}
					fmt.Print(string(repoData))
				}
			default:
				if getter, ok := boolGetters[key]; ok {
					fmt.Println(getter())
				} else if getter, ok := stringGetters[key]; ok {
					fmt.Println(getter())
				} else {
					return cliutils.FormatCliExit(gotext.Get("unknown config key: %s", key), nil)
				}
			}

			return nil
		}),
	}
}
