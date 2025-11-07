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
	"os"
	"os/exec"
	"syscall"

	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"

	"go.stplr.dev/stplr/internal/plugins"
	"go.stplr.dev/stplr/internal/sandbox"
	"go.stplr.dev/stplr/internal/utils"
)

func InternalSandbox() *cli.Command {
	return &cli.Command{
		Name:                  "_internal-sandbox",
		HideHelp:              true,
		Hidden:                true,
		SkipFlagParsing:       true,
		EnableShellCompletion: false,
		Action: func(ctx context.Context, c *cli.Command) error {
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
			cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = os.Environ()
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC,
			}
			return cmd.Run()
		},
	}
}

func InternalPluginProvider() *cli.Command {
	return &cli.Command{
		Name:     plugins.ProviderSubcommand,
		HideHelp: true,
		Hidden:   true,
		Action: func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForPluginsServe(ctx)
			if err != nil {
				return err
			}
			defer f()
			return plugins.Serve(plugins.PluginMap(d.Puller, d.Scripter))
		},
	}
}

func InternalPluginProviderRoot() *cli.Command {
	return &cli.Command{
		Name:     plugins.RootProviderSubcommand,
		HideHelp: true,
		Hidden:   true,
		Action: func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForPluginsServeRoot(ctx)
			if err != nil {
				return err
			}
			defer f()
			return plugins.Serve(plugins.RootPluginMap(
				d.Installer,
				d.Copier,
				d.SystemConfigWriter,
			))
		},
	}
}
