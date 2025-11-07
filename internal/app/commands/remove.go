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

package commands

import (
	"context"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/cliutils2"
	"go.stplr.dev/stplr/internal/usecase/remove/action"
	"go.stplr.dev/stplr/internal/usecase/remove/shell"
)

func removeCmdActionChecks(_ context.Context, c *cli.Command) error {
	args := c.Args()
	if args.Len() < 1 {
		return errors.NewI18nError(gotext.Get("Command remove expected at least 1 argument, got %d", args.Len()))
	}
	return nil
}

func RemoveCmd() *cli.Command {
	return &cli.Command{
		Name:    "remove",
		Usage:   gotext.Get("Remove an installed package"),
		Aliases: []string{"rm"},
		ShellComplete: cliutils.BashCompleteWithError(func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForRemoveShellComp(ctx)
			if err != nil {
				return err
			}
			defer f()

			return shell.New(d.Mgr, d.DB).Run(ctx)
		}),
		Action: cliutils2.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if err := removeCmdActionChecks(ctx, c); err != nil {
				return err
			}

			d, f, err := deps.ForRemoveAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			return action.New(d.Mgr).Run(ctx, action.Options{
				Pkgs:        c.Args().Slice(),
				Interactive: c.Bool("interactive"),
			})
		}),
	}
}
