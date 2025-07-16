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
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/logger"
	"go.stplr.dev/stplr/internal/manager"
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
		Action: func(c *cli.Context) error {
			logger.SetupForGoPlugin()

			if err := utils.EnsureIsAlrUser(); err != nil {
				return err
			}

			// Before escalating the rights, we made sure that
			// this is an ALR user, so it looks safe.
			err := utils.EscalateToRootUid()
			if err != nil {
				return cliutils.FormatCliExit("cannot escalate to root", err)
			}

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
		},
	}
}

func Mount(target string) (string, func(), error) {
	exe, err := os.Executable()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(exe, "_internal-temporary-mount", target)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", nil, fmt.Errorf("failed to start mount: %w", err)
	}

	scanner := bufio.NewScanner(stdoutPipe)
	var mountPath string
	if scanner.Scan() {
		mountPath = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Process.Kill()
		return "", nil, fmt.Errorf("failed to read mount output: %w", err)
	}

	if mountPath == "" {
		_ = cmd.Process.Kill()
		return "", nil, errors.New("mount failed: no target path returned")
	}

	cleanup := func() {
		slog.Debug("cleanup triggered")
		_, _ = fmt.Fprintln(stdinPipe, "")
		_ = cmd.Wait()
	}

	return mountPath, cleanup, nil
}

func InternalMountCmd() *cli.Command {
	return &cli.Command{
		Name:     "_internal-temporary-mount",
		HideHelp: true,
		Hidden:   true,
		Action: func(c *cli.Context) error {
			logger.SetupForGoPlugin()

			sourceDir := c.Args().First()

			u, err := user.Current()
			if err != nil {
				return cliutils.FormatCliExit("cannot get current user", err)
			}

			_, alrGid, err := utils.GetUidGidStaplerUser()
			if err != nil {
				return cliutils.FormatCliExit("cannot get stapler-builder user", err)
			}

			if _, err := os.Stat(sourceDir); err != nil {
				return cliutils.FormatCliExit(fmt.Sprintf("cannot read %s", sourceDir), err)
			}

			if err := utils.EnuseIsPrivilegedGroupMember(); err != nil {
				return err
			}

			// Before escalating the rights, we made sure that
			// 1. user in wheel group
			// 2. user can access sourceDir
			if err := utils.EscalateToRootUid(); err != nil {
				return err
			}
			if err := syscall.Setgid(alrGid); err != nil {
				return err
			}

			if err := os.MkdirAll(constants.StaRunDir, 0o770); err != nil {
				return cliutils.FormatCliExit(fmt.Sprintf("failed to create %s", constants.StaRunDir), err)
			}

			if err := os.Chown(constants.StaRunDir, 0, alrGid); err != nil {
				return cliutils.FormatCliExit(fmt.Sprintf("failed to chown %s", constants.StaRunDir), err)
			}

			targetDir := filepath.Join(constants.StaRunDir, fmt.Sprintf("bindfs-%d", os.Getpid()))
			// 0750: owner (root) and group (alr)
			if err := os.MkdirAll(targetDir, 0o750); err != nil {
				return cliutils.FormatCliExit("error creating bindfs target directory", err)
			}

			//  chown AlrRunDir/mounts/bindfs-* to (root:alr),
			//  so alr user can access dir
			if err := os.Chown(targetDir, 0, alrGid); err != nil {
				return cliutils.FormatCliExit("failed to chown bindfs directory", err)
			}

			bindfsCmd := exec.Command(
				"bindfs",
				fmt.Sprintf("--map=%s/%s:@%s/@%s", u.Uid, constants.BuilderUser, u.Gid, constants.BuilderGroup),
				sourceDir,
				targetDir,
			)

			bindfsCmd.Stderr = os.Stderr

			if err := bindfsCmd.Run(); err != nil {
				return cliutils.FormatCliExit("failed to strart bindfs", err)
			}

			fmt.Println(targetDir)

			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')

			slog.Debug("start unmount", "dir", targetDir)

			umountCmd := exec.Command("umount", targetDir)
			umountCmd.Stderr = os.Stderr
			if err := umountCmd.Run(); err != nil {
				return cliutils.FormatCliExit(fmt.Sprintf("failed to unmount %s", targetDir), err)
			}

			if err := os.Remove(targetDir); err != nil {
				return cliutils.FormatCliExit(fmt.Sprintf("error removing directory %s", targetDir), err)
			}

			return nil
		},
	}
}
