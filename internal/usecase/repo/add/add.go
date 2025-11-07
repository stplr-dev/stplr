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

package add

import (
	"context"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/config/common"
	"go.stplr.dev/stplr/internal/service/repos"
	"go.stplr.dev/stplr/pkg/types"
)

type useCase struct {
	cfg *config.ALRConfig
	r   *repos.Repos
}

func New(cfg *config.ALRConfig, p *repos.Repos) *useCase {
	return &useCase{cfg, p}
}

type Options struct {
	Name string
	URL  string
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	repos := u.cfg.Repos()
	for _, repo := range repos {
		if repo.URL == opts.URL || repo.Name == opts.Name {
			return errors.NewI18nError(gotext.Get("Repo \"%s\" already exists", repo.Name))
		}
	}

	repo := types.Repo{
		Name: opts.Name,
		URL:  opts.URL,
	}

	newRepo, err := u.r.Pull(ctx, repo)
	if err != nil {
		return err
	}

	repos = append(repos, newRepo)
	u.cfg.SetRepos(repos)
	err = u.cfg.Save(common.SOURCE_SYSTEM)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error saving config"))
	}

	return nil
}
