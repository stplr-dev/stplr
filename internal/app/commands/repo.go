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
	"errors"
	"fmt"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/app/deps"

	"go.stplr.dev/stplr/internal/cliutils2"
	"go.stplr.dev/stplr/internal/usecase/repo/add"
	"go.stplr.dev/stplr/internal/usecase/repo/remove"
	"go.stplr.dev/stplr/internal/usecase/repo/setref"
	"go.stplr.dev/stplr/internal/usecase/repo/seturl"
)

var errMissingArgs = errors.New("missing args")

func ShellCompleteRepoName(ctx context.Context, c *cli.Command) {
	if c.NArg() == 0 {
		// Get repo names from config
		d, f, err := deps.ForConfigGetAction(ctx)
		if err != nil {
			return
		}
		defer f()

		for _, repo := range d.Config.Repos() {
			fmt.Println(repo.Name)
		}
	}
}

func RepoCmd() *cli.Command {
	return &cli.Command{
		Name:  "repo",
		Usage: gotext.Get("Manage repos"),
		Commands: []*cli.Command{
			RemoveRepoCmd(),
			AddRepoCmd(),
			SetRepoRefCmd(),
			RepoMirrorCmd(),
			SetUrlCmd(),
		},
	}
}

func RemoveRepoCmd() *cli.Command {
	return &cli.Command{
		Name:          "remove",
		Usage:         gotext.Get("Remove an existing repository"),
		Aliases:       []string{"rm"},
		ArgsUsage:     gotext.Get("<name>"),
		ShellComplete: ShellCompleteRepoName,
		Action: cliutils2.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return errMissingArgs
			}

			d, f, err := deps.ForRepoRemoveAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			return remove.New(d.Config, d.DB).Run(ctx, remove.Options{Name: c.Args().Get(0)})
		}),
	}
}

func AddRepoCmd() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     gotext.Get("Add a new repository"),
		ArgsUsage: gotext.Get("<name> <url>"),
		Action: cliutils2.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return errMissingArgs
			}

			d, f, err := deps.ForRepoAddAction(ctx)
			if err != nil {
				return err
			}
			defer f()

			return add.New(d.Config, d.Repos).Run(ctx, add.Options{
				Name: c.Args().Get(0),
				URL:  c.Args().Get(1),
			})
		}),
	}
}

func SetRepoRefCmd() *cli.Command {
	return &cli.Command{
		Name:          "set-ref",
		Usage:         gotext.Get("Set the reference of the repository"),
		ArgsUsage:     gotext.Get("<name> <ref>"),
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

			return setref.
				New(d.Repos).
				Run(ctx, setref.Options{
					Name: c.Args().Get(0),
					Ref:  c.Args().Get(1),
				})
		}),
	}
}

func SetUrlCmd() *cli.Command {
	return &cli.Command{
		Name:          "set-url",
		Usage:         gotext.Get("Set the main url of the repository"),
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

			return seturl.New(d.Repos).Run(ctx, seturl.Options{
				Name: c.Args().Get(0),
				URL:  c.Args().Get(1),
			})
		}),
	}
}
