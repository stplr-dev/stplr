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

package repos

import (
	"context"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/pkg/types"
)

type ReposPuller interface {
	Pull(ctx context.Context, repos []types.Repo) error
}

type SysDropper interface {
	DropCapsToBuilderUser() error
}

type Repos struct {
	cfg *config.ALRConfig
	sys SysDropper
	rp  ReposPuller
}

func New(cfg *config.ALRConfig, sys SysDropper, rp ReposPuller) *Repos {
	return &Repos{
		cfg: cfg,
		sys: sys,
		rp:  rp,
	}
}

// Modify updates the configured repositories using the provided callback.
// The updated list is saved into the system configuration.
//
// WARNING: Any root-level operations must be performed before this method,
// because process privileges are permanently dropped to the builder user.
//
// The callback can return nil to exclude a repo from the final list.
func (r *Repos) Modify(ctx context.Context, c func(repo types.Repo) *types.Repo) error {
	return r.ModifySlice(ctx, func(repos []types.Repo) ([]types.Repo, error) {
		newRepos := make([]types.Repo, 0, len(repos))
		for _, repo := range repos {
			if r := c(repo); r != nil {
				newRepos = append(newRepos, *r)
			}
		}
		return newRepos, nil
	}, true)
}

// WARNING: Any root-level operations must be performed before this method,
// because process privileges are permanently dropped to the builder user.
func (r *Repos) ModifySlice(ctx context.Context, c func(repos []types.Repo) ([]types.Repo, error), pull bool) error {
	// TODO: Consider rewriting this to use GetSafeConfigModifier or smth else
	// to avoid the side-effect of dropping privileges here.
	// This is may not be ideal, but refactoring is deferred.

	repos := r.cfg.Repos()
	newRepos, err := c(repos)
	if err != nil {
		return err
	}
	r.cfg.SetRepos(newRepos)
	err = r.cfg.System.Save()
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error saving config"))
	}

	err = r.sys.DropCapsToBuilderUser()
	if err != nil {
		return err
	}

	if pull {
		err = r.rp.Pull(ctx, newRepos)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error pulling repositories"))
		}
	}
	return nil
}
