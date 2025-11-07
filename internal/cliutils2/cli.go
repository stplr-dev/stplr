// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
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

package cliutils2

import (
	"context"
	"os"
	"os/exec"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/utils"

	"go.stplr.dev/stplr/internal/config"
)

func runAsRoot(ctx context.Context, rootCmd string, args []string) error {
	// gosec:disable G204
	cmd := exec.CommandContext(ctx, rootCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			cli.OsExiter(exitErr.ExitCode())
		}
	}
	return nil
}

func handleNonRoot(ctx context.Context, out output.Output, cfg *config.ALRConfig) error {
	if !cfg.UseRootCmd() {
		return cli.Exit(gotext.Get("You need to be root to perform this action"), 1)
	}

	executable, err := os.Executable()
	if err != nil {
		return cliutils.FormatCliExit("failed to get executable path", err)
	}

	out.Warn(gotext.Get("This action requires elevated privileges."))
	out.Warn(gotext.Get("Attempting to run as root using '%s'...", cfg.RootCmd()))
	args := append([]string{executable}, os.Args[1:]...)
	return runAsRoot(ctx, cfg.RootCmd(), args)
}

func RootNeededAction(f cli.ActionFunc) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		out := output.FromContext(ctx)

		deps, cleanup, err := deps.ForConfigGetAction(ctx)
		if err != nil {
			return err
		}
		defer cleanup()

		if utils.IsNotRoot() {
			return handleNonRoot(ctx, out, deps.Config)
		}

		return f(ctx, c)
	}
}

func ReadonlyAction(f cli.ActionFunc) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		if utils.IsNotRoot() {
			// TODO: relaunch in userns with HOME hide
		} else {
			if err := cliutils.ExitIfCantDropCapsToBuilderUserNoPrivs(); err != nil {
				return err
			}
		}

		return f(ctx, c)
	}
}
