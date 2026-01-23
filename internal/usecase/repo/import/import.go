// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package repo_import

import (
	"context"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/repoutils"
	"go.stplr.dev/stplr/pkg/types"
)

type Repos interface {
	ModifySlice(ctx context.Context, c func(repos []types.Repo) ([]types.Repo, error), pull bool) error
	HasRepo(name string) bool
}

type useCase struct {
	cfg *config.ALRConfig
	r   Repos
}

func New(cfg *config.ALRConfig, p Repos) *useCase {
	return &useCase{cfg, p}
}

type Options struct {
	Name           string
	ConfigContent  string
	NoPull         bool
	IgnoreExisting bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	if u.r.HasRepo(opts.Name) {
		if opts.IgnoreExisting {
			return nil
		}
		return errors.NewI18nError(
			gotext.Get("Repository %q already exists", opts.Name),
		)
	}

	r, err := repoutils.RepoFromConfigString(opts.ConfigContent)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Failed to read repo configuration"))
	}

	if opts.Name != "" {
		r.Name = opts.Name
	}

	err = u.r.ModifySlice(ctx, func(repos []types.Repo) ([]types.Repo, error) {
		repos = append(repos, *r)
		return repos, nil
	}, !opts.NoPull)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Failed to import repository"))
	}

	return nil
}
