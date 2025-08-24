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

package main

import (
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/logger"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/sandbox"
	"go.stplr.dev/stplr/internal/utils"
)

func InternalBuildCmd() *cli.Command {
	return &cli.Command{
		Name:     "_internal-safe-script-executor",
		HideHelp: true,
		Hidden:   true,
		Action: func(c *cli.Context) error {
			logger.SetupForGoPlugin()

			slog.Debug("start _internal-safe-script-executor", "uid", syscall.Getuid(), "gid", syscall.Getgid())

			if err := utils.ExitIfCantDropCapsToAlrUser(); err != nil {
				return err
			}

			cfg := config.New()
			err := cfg.Load()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error loading config"), err)
			}

			logger := hclog.New(&hclog.LoggerOptions{
				Name:        "plugin",
				Output:      os.Stderr,
				Level:       hclog.Debug,
				JSONFormat:  false,
				DisableTime: true,
			})

			plugin.Serve(&plugin.ServeConfig{
				HandshakeConfig: build.HandshakeConfig,
				Plugins: map[string]plugin.Plugin{
					"script-executor": &build.ScriptExecutorPlugin{
						Impl: build.NewLocalScriptExecutor(cfg),
					},
				},
				Logger: logger,
			})
			return nil
		},
	}
}

func InternalReposCmd() *cli.Command {
	return &cli.Command{
		Name:     "_internal-repos",
		HideHelp: true,
		Hidden:   true,
		Action: utils.RootNeededAction(func(ctx *cli.Context) error {
			logger.SetupForGoPlugin()

			if err := utils.ExitIfCantDropCapsToAlrUser(); err != nil {
				return err
			}

			deps, err := appbuilder.
				New(ctx.Context).
				WithConfig().
				WithDB().
				WithReposNoPull().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			pluginCfg := build.GetPluginServeCommonConfig()
			pluginCfg.Plugins = map[string]plugin.Plugin{
				"repos": &build.ReposExecutorPlugin{
					Impl: build.NewRepos(
						deps.Repos,
					),
				},
			}
			plugin.Serve(pluginCfg)
			return nil
		}),
	}
}

func InternalInstallCmd() *cli.Command {
	return &cli.Command{
		Name:     "_internal-installer",
		HideHelp: true,
		Hidden:   true,
		Action: utils.RootNeededAction(func(c *cli.Context) error {
			logger.SetupForGoPlugin()

			deps, err := appbuilder.
				New(c.Context).
				WithConfig().
				WithDB().
				WithReposNoPull().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			logger := hclog.New(&hclog.LoggerOptions{
				Name:        "plugin",
				Output:      os.Stderr,
				Level:       hclog.Trace,
				JSONFormat:  true,
				DisableTime: true,
			})

			plugin.Serve(&plugin.ServeConfig{
				HandshakeConfig: build.HandshakeConfig,
				Plugins: map[string]plugin.Plugin{
					"installer": &build.InstallerExecutorPlugin{
						Impl: build.NewInstaller(
							manager.Detect(),
						),
					},
				},
				Logger: logger,
			})
			return nil
		}),
	}
}

func InternalCoplyFiles() *cli.Command {
	return &cli.Command{
		Name:     "_internal-script-copier",
		HideHelp: true,
		Hidden:   true,
		Action: utils.RootNeededAction(func(c *cli.Context) error {
			logger.SetupForGoPlugin()
			logger := hclog.New(&hclog.LoggerOptions{
				Name:        "plugin",
				Output:      os.Stderr,
				Level:       hclog.Trace,
				JSONFormat:  true,
				DisableTime: true,
			})
			plugin.Serve(&plugin.ServeConfig{
				HandshakeConfig: build.HandshakeConfig,
				Plugins: map[string]plugin.Plugin{
					"script-copier": &build.ScriptCopierPlugin{
						Impl: build.NewLocalScriptCopierExecutor(),
					},
				},
				Logger: logger,
			})

			return nil
		}),
	}
}

func InternalSandbox() *cli.Command {
	return &cli.Command{
		Name:     "_internal-sandbox",
		HideHelp: true,
		Hidden:   true,
		Action: func(c *cli.Context) error {
			if c.NArg() < 4 {
				return fmt.Errorf("not enough arguments: need srcDir, pkgDir, command")
			}

			cmdArgs := c.Args().Slice()[3:]
			if len(cmdArgs) == 0 {
				return fmt.Errorf("no command specified")
			}

			if err := sandbox.Setup(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)); err != nil {
				return fmt.Errorf("failed to setup sandbox: %w", err)
			}

			if err := utils.NoNewPrivs(); err != nil {
				return fmt.Errorf("failed to drop privileges: %w", err)
			}

			//gosec:disable G204 -- Expected
			return syscall.Exec(cmdArgs[0], cmdArgs, os.Environ())
		},
	}
}
