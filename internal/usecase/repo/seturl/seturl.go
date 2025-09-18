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

package seturl

import (
	"context"

	"go.stplr.dev/stplr/pkg/types"
)

type Repos interface {
	Modify(ctx context.Context, c func(repo types.Repo) *types.Repo) error
}

type useCase struct {
	r Repos
}

func New(r Repos) *useCase {
	return &useCase{r}
}

type Options struct {
	Name string
	URL  string
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	return u.r.Modify(ctx, func(repo types.Repo) *types.Repo {
		if repo.Name == opts.Name {
			repo.URL = opts.URL
		}
		return &repo
	})
}
