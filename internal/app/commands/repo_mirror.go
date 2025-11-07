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
	"go.stplr.dev/stplr/internal/usecase/repo/mirrors/add"
	mirrorsClear "go.stplr.dev/stplr/internal/usecase/repo/mirrors/clear"
	"go.stplr.dev/stplr/internal/usecase/repo/mirrors/remove"
)

func RepoMirrorCmd() *cli.Command {
	return &cli.Command{
		Name:  "mirror",
		Usage: gotext.Get("Manage mirrors of repos"),
		Commands: []*cli.Command{
			AddMirror(),
			RemoveMirror(),
			ClearMirrors(),
		},
	}
}

func AddMirror() *cli.Command {
	return &cli.Command{
		Name:          "add",
		Usage:         gotext.Get("Add a mirror URL to repository"),
		ArgsUsage:     gotext.Get("<name> <url>"),
		ShellComplete: ShellCompleteRepoName,
		Action: cliutils2.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return errMissingArgs
			}

			d, f, err := deps.ForUniversalReposModificationActionDeps(ctx)
			if err != nil {
				return err
			}
			defer f()

			return add.New(d.Repos).Run(ctx, add.Options{
				Name: c.Args().Get(0),
				URL:  c.Args().Get(1),
			})
		}),
	}
}

func RemoveMirror() *cli.Command {
	return &cli.Command{
		Name:          "remove",
		Aliases:       []string{"rm"},
		Usage:         gotext.Get("Remove mirror from the repository"),
		ArgsUsage:     gotext.Get("<name> <url>"),
		ShellComplete: ShellCompleteRepoName,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "ignore-missing",
				Usage: gotext.Get("Ignore if mirror does not exist"),
			},
			&cli.BoolFlag{
				Name:    "partial",
				Aliases: []string{"p"},
				Usage:   gotext.Get("Match partial URL (e.g., github.com instead of full URL)"),
			},
		},
		Action: cliutils2.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return errMissingArgs
			}

			d, f, err := deps.ForUniversalReposModificationActionDeps(ctx)
			if err != nil {
				return err
			}
			defer f()

			return remove.New(d.Repos).Run(ctx, remove.Options{
				Name:          c.Args().Get(0),
				URL:           c.Args().Get(1),
				IgnoreMissing: c.Bool("ignore-missing"),
				PartialMatch:  c.Bool("partial"),
			})
		}),
	}
}

func ClearMirrors() *cli.Command {
	return &cli.Command{
		Name:          "clear",
		Aliases:       []string{"rm-all"},
		Usage:         gotext.Get("Remove all mirrors from the repository"),
		ArgsUsage:     gotext.Get("<name>"),
		ShellComplete: ShellCompleteRepoName,
		Action: cliutils2.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return errMissingArgs
			}

			name := c.Args().Get(0)

			d, f, err := deps.ForUniversalReposModificationActionDeps(ctx)
			if err != nil {
				return err
			}
			defer f()

			return mirrorsClear.New(d.Repos).Run(ctx, mirrorsClear.Options{
				Name: name,
			})
		}),
	}
}
