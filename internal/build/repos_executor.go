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

package build

import (
	"context"

	"go.stplr.dev/stplr/internal/repos"
	"go.stplr.dev/stplr/pkg/types"
)

type reposExecutor struct{ r *repos.Repos }

func NewRepos(r *repos.Repos) *reposExecutor {
	return &reposExecutor{r}
}

func (r *reposExecutor) PullOneAndUpdateFromConfig(ctx context.Context, repo *types.Repo) (types.Repo, error) {
	if err := r.r.PullOneAndUpdateFromConfig(ctx, repo); err != nil {
		return *repo, err
	}
	return *repo, nil
}
