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
	"go.stplr.dev/stplr/internal/cliutils2"
	"go.stplr.dev/stplr/internal/usecase/support"
)

func SupportCmd() *cli.Command {
	return &cli.Command{
		Name:        "support",
		Usage:       gotext.Get("Dump debug information"),
		Description: gotext.Get("Gathers system, configuration, and log data into a diagnostic archive"),
		Action: cliutils2.ReadonlyAction(func(ctx context.Context, c *cli.Command) error {
			d, f, err := deps.ForSupportAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			return support.New(d.Out).Run(ctx)
		}),
	}
}
