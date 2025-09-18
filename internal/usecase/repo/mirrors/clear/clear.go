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

package clear

import (
	"context"
	stdErrors "errors"

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
	Name string
}

var errMissingRepo = stdErrors.New("repo does not exist")

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

	repo.Mirrors = []string{}

	return repos, nil
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	return u.r.ModifySlice(ctx, func(repos []types.Repo) ([]types.Repo, error) {
		return u.modify(opts, repos)
	}, false)
}
