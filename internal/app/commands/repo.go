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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"
	"golang.org/x/exp/slices"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/types"
)

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
		Name:      "remove",
		Usage:     gotext.Get("Remove an existing repository"),
		Aliases:   []string{"rm"},
		ArgsUsage: gotext.Get("<name>"),
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.NArg() == 0 {
				// Get repo names from config
				deps, err := appbuilder.New(ctx).WithConfig().Build()
				if err != nil {
					return
				}
				defer deps.Defer()

				for _, repo := range deps.Cfg.Repos() {
					fmt.Println(repo.Name)
				}
			}
		},
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return cliutils.FormatCliExit("missing args", nil)
			}
			name := c.Args().Get(0)

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			cfg := deps.Cfg

			found := false
			index := 0
			reposSlice := cfg.Repos()
			for i, repo := range reposSlice {
				if repo.Name == name {
					index = i
					found = true
				}
			}
			if !found {
				return cliutils.FormatCliExit(gotext.Get("Repo \"%s\" does not exist", name), nil)
			}

			cfg.SetRepos(slices.Delete(reposSlice, index, index+1))

			err = os.RemoveAll(filepath.Join(cfg.GetPaths().RepoDir, name))
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error removing repo directory"), err)
			}
			err = cfg.System.Save()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
			}

			if err := utils.ExitIfCantDropCapsToBuilderUser(); err != nil {
				return err
			}

			deps, err = appbuilder.
				New(ctx).
				UseConfig(cfg).
				WithDB().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			err = deps.DB.DeletePkgs(ctx, "repository = ?", name)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error removing packages from database"), err)
			}

			return nil
		}),
	}
}

func AddRepoCmd() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     gotext.Get("Add a new repository"),
		ArgsUsage: gotext.Get("<name> <url>"),
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			name := c.Args().Get(0)
			repoURL := c.Args().Get(1)

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			cfg := deps.Cfg

			reposSlice := cfg.Repos()
			for _, repo := range reposSlice {
				if repo.URL == repoURL || repo.Name == name {
					return cliutils.FormatCliExit(gotext.Get("Repo \"%s\" already exists", repo.Name), nil)
				}
			}

			newRepo := types.Repo{
				Name: name,
				URL:  repoURL,
			}

			r, close, err := build.GetSafeReposExecutor()
			if err != nil {
				return err
			}
			defer close()

			newRepo, err = r.PullOneAndUpdateFromConfig(ctx, &newRepo)
			if err != nil {
				return err
			}

			reposSlice = append(reposSlice, newRepo)
			cfg.SetRepos(reposSlice)

			err = cfg.System.Save()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
			}

			return nil
		}),
	}
}

func SetRepoRefCmd() *cli.Command {
	return &cli.Command{
		Name:      "set-ref",
		Usage:     gotext.Get("Set the reference of the repository"),
		ArgsUsage: gotext.Get("<name> <ref>"),
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.NArg() == 0 {
				// Get repo names from config
				deps, err := appbuilder.New(ctx).WithConfig().Build()
				if err != nil {
					return
				}
				defer deps.Defer()

				for _, repo := range deps.Cfg.Repos() {
					fmt.Println(repo.Name)
				}
			}
		},
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			name := c.Args().Get(0)
			ref := c.Args().Get(1)

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithReposNoPull().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			repos := deps.Cfg.Repos()
			newRepos := []types.Repo{}
			for _, repo := range repos {
				if repo.Name == name {
					repo.Ref = ref
				}
				newRepos = append(newRepos, repo)
			}
			deps.Cfg.SetRepos(newRepos)
			err = deps.Cfg.System.Save()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
			}

			err = deps.Repos.Pull(ctx, newRepos)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error pulling repositories"), err)
			}

			return nil
		}),
	}
}

func SetUrlCmd() *cli.Command {
	return &cli.Command{
		Name:      "set-url",
		Usage:     gotext.Get("Set the main url of the repository"),
		ArgsUsage: gotext.Get("<name> <url>"),
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.NArg() == 0 {
				// Get repo names from config
				deps, err := appbuilder.New(ctx).WithConfig().Build()
				if err != nil {
					return
				}
				defer deps.Defer()

				for _, repo := range deps.Cfg.Repos() {
					fmt.Println(repo.Name)
				}
			}
		},
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			name := c.Args().Get(0)
			repoUrl := c.Args().Get(1)

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithReposNoPull().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			repos := deps.Cfg.Repos()
			newRepos := []types.Repo{}
			for _, repo := range repos {
				if repo.Name == name {
					repo.URL = repoUrl
				}
				newRepos = append(newRepos, repo)
			}
			deps.Cfg.SetRepos(newRepos)
			err = deps.Cfg.System.Save()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
			}

			err = deps.Repos.Pull(ctx, newRepos)
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error pulling repositories"), err)
			}

			return nil
		}),
	}
}

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
		Name:      "add",
		Usage:     gotext.Get("Add a mirror URL to repository"),
		ArgsUsage: gotext.Get("<name> <url>"),
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.NArg() == 0 {
				deps, err := appbuilder.New(ctx).WithConfig().Build()
				if err != nil {
					return
				}
				defer deps.Defer()

				for _, repo := range deps.Cfg.Repos() {
					fmt.Println(repo.Name)
				}
			}
		},
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			name := c.Args().Get(0)
			url := c.Args().Get(1)

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithReposNoPull().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			repos := deps.Cfg.Repos()
			for i, repo := range repos {
				if repo.Name == name {
					repos[i].Mirrors = append(repos[i].Mirrors, url)
					break
				}
			}
			deps.Cfg.SetRepos(repos)
			err = deps.Cfg.System.Save()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
			}

			return nil
		}),
	}
}

func RemoveMirror() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     gotext.Get("Remove mirror from the repository"),
		ArgsUsage: gotext.Get("<name> <url>"),
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			deps, err := appbuilder.New(ctx).WithConfig().Build()
			if err != nil {
				return
			}
			defer deps.Defer()

			if c.NArg() == 0 {
				for _, repo := range deps.Cfg.Repos() {
					fmt.Println(repo.Name)
				}
			}
		},
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
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			name := c.Args().Get(0)
			urlToRemove := c.Args().Get(1)
			ignoreMissing := c.Bool("ignore-missing")
			partialMatch := c.Bool("partial")

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithReposNoPull().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			reposSlice := deps.Cfg.Repos()
			repoIndex := -1
			urlIndicesToRemove := []int{}

			// Находим репозиторий
			for i, repo := range reposSlice {
				if repo.Name == name {
					repoIndex = i
					break
				}
			}

			if repoIndex == -1 {
				if ignoreMissing {
					return nil // Тихо завершаем, если репозиторий не найден
				}
				return cliutils.FormatCliExit(gotext.Get("Repo \"%s\" does not exist", name), nil)
			}

			// Ищем зеркала для удаления
			repo := reposSlice[repoIndex]
			for j, mirror := range repo.Mirrors {
				var match bool
				if partialMatch {
					// Частичное совпадение - проверяем, содержит ли зеркало указанную строку
					match = strings.Contains(mirror, urlToRemove)
				} else {
					// Точное совпадение
					match = mirror == urlToRemove
				}

				if match {
					urlIndicesToRemove = append(urlIndicesToRemove, j)
				}
			}

			if len(urlIndicesToRemove) == 0 {
				if ignoreMissing {
					return nil
				}
				if partialMatch {
					return cliutils.FormatCliExit(gotext.Get("No mirrors containing \"%s\" found in repo \"%s\"", urlToRemove, name), nil)
				} else {
					return cliutils.FormatCliExit(gotext.Get("URL \"%s\" does not exist in repo \"%s\"", urlToRemove, name), nil)
				}
			}

			for i := len(urlIndicesToRemove) - 1; i >= 0; i-- {
				urlIndex := urlIndicesToRemove[i]
				reposSlice[repoIndex].Mirrors = slices.Delete(reposSlice[repoIndex].Mirrors, urlIndex, urlIndex+1)
			}

			deps.Cfg.SetRepos(reposSlice)
			err = deps.Cfg.System.Save()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
			}

			if len(urlIndicesToRemove) > 1 {
				fmt.Println(gotext.Get("Removed %d mirrors from repo \"%s\"\n", len(urlIndicesToRemove), name))
			}

			return nil
		}),
	}
}

func ClearMirrors() *cli.Command {
	return &cli.Command{
		Name:      "clear",
		Aliases:   []string{"rm-all"},
		Usage:     gotext.Get("Remove all mirrors from the repository"),
		ArgsUsage: gotext.Get("<name>"),
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.NArg() == 0 {
				// Get repo names from config
				deps, err := appbuilder.New(ctx).WithConfig().Build()
				if err != nil {
					return
				}
				defer deps.Defer()

				for _, repo := range deps.Cfg.Repos() {
					fmt.Println(repo.Name)
				}
			}
		},
		Action: utils.RootNeededAction(func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return cliutils.FormatCliExit("missing args", nil)
			}

			name := c.Args().Get(0)

			deps, err := appbuilder.
				New(ctx).
				WithConfig().
				WithDB().
				WithReposNoPull().
				Build()
			if err != nil {
				return err
			}
			defer deps.Defer()

			reposSlice := deps.Cfg.Repos()
			repoIndex := -1
			urlIndicesToRemove := []int{}

			// Находим репозиторий
			for i, repo := range reposSlice {
				if repo.Name == name {
					repoIndex = i
					break
				}
			}

			if repoIndex == -1 {
				return cliutils.FormatCliExit(gotext.Get("Repo \"%s\" does not exist", name), nil)
			}

			reposSlice[repoIndex].Mirrors = []string{}

			deps.Cfg.SetRepos(reposSlice)
			err = deps.Cfg.System.Save()
			if err != nil {
				return cliutils.FormatCliExit(gotext.Get("Error saving config"), err)
			}

			if len(urlIndicesToRemove) > 1 {
				fmt.Println(gotext.Get("Removed %d mirrors from repo \"%s\"\n", len(urlIndicesToRemove), name))
			}

			return nil
		}),
	}
}
