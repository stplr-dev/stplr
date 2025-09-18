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

package remove

import (
	"context"
	stdErrors "errors"
	"fmt"
	"slices"
	"strings"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/pkg/types"
)

type Repos interface {
	ModifySlice(ctx context.Context, c func(repos []types.Repo) ([]types.Repo, error), pull bool) error
}

type useCase struct {
	r Repos
}

func New(r Repos) *useCase {
	return &useCase{r}
}

type Options struct {
	Name          string
	URL           string
	IgnoreMissing bool
	PartialMatch  bool
}

var (
	errMissingRepo   = stdErrors.New("repo does not exist")
	errMissingMirror = stdErrors.New("mirror does not exist")
)

func (u *useCase) modify(opts Options, repos []types.Repo) ([]types.Repo, error) {
	var repo *types.Repo

	for i := range repos {
		if repos[i].Name == opts.Name {
			repo = &repos[i]
			break
		}
	}

	if repo == nil {
		return nil, errMissingRepo
	}

	originalCount := len(repo.Mirrors)
	repo.Mirrors = slices.DeleteFunc(repo.Mirrors, func(mirror string) bool {
		if opts.PartialMatch {
			return strings.Contains(mirror, opts.URL)
		}
		return mirror == opts.URL
	})

	if len(repo.Mirrors) == originalCount {
		return nil, errMissingMirror
	}

	return repos, nil
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	count := 0

	err := u.r.ModifySlice(ctx, func(repos []types.Repo) ([]types.Repo, error) {
		newRepos, err := u.modify(opts, repos)
		if err != nil {
			return nil, err
		}

		count = len(repos) - len(newRepos)

		return newRepos, err
	}, false)
	if err != nil {
		switch err {
		case errMissingRepo:
			if opts.IgnoreMissing {
				return nil
			}
		case errMissingMirror:
			if opts.IgnoreMissing {
				return nil
			}
			if opts.PartialMatch {
				return errors.NewI18nError(gotext.Get("No mirrors containing \"%s\" found in repo \"%s\"", opts.URL, opts.Name))
			}
			return errors.NewI18nError(gotext.Get("URL \"%s\" does not exist in repo \"%s\"", opts.URL, opts.Name))
		default:
			return err
		}
	}

	fmt.Println(gotext.Get("Removed %d mirrors from repo \"%s\"\n", count, opts.Name))

	return nil
}
