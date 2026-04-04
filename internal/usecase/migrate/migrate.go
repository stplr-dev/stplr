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

package migrate

import (
	"context"
	"log/slog"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/service/repos"
)

type resetter interface {
	Reset(ctx context.Context) error
}

type ReposGetter func() (*repos.Repos, deps.Cleanup, error)

type useCase struct {
	cfg           *config.ALRConfig
	dbResetter    resetter
	cacheResetter resetter
	reposGetter   ReposGetter
}

func New(cfg *config.ALRConfig, dbResetter, cacheResetter resetter, reposGetter ReposGetter) *useCase {
	return &useCase{cfg, dbResetter, cacheResetter, reposGetter}
}

func (u *useCase) Run(ctx context.Context) error {
	if err := u.migrateInlineRepos(ctx); err != nil {
		slog.Warn("failed to migrate inline repos", "err", err)
	}

	if err := u.dbResetter.Reset(ctx); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error resetting database"))
	}

	if err := u.cacheResetter.Reset(ctx); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error resetting cache"))
	}

	r, cleanup, err := u.reposGetter()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := r.RereadAll(ctx, repos.WithDeleteFailed()); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Failed to reread all"))
	}

	return nil
}

// migrateInlineRepos moves [[repo]] entries from stplr.toml to individual files
// in /etc/stplr/repos.d/.
func (u *useCase) migrateInlineRepos(ctx context.Context) error {
	inline := u.cfg.InlineRepos()
	if len(inline) == 0 {
		return nil
	}

	migrated := 0
	for _, repo := range inline {
		if err := u.cfg.AddRepo(repo); err != nil {
			slog.Warn("failed to migrate inline repo", "repo", repo.Name, "err", err)
			continue
		}
		migrated++
	}

	if migrated == 0 {
		return nil
	}

	slog.Info("migrated inline repos to repos.d", "count", migrated)
	return u.cfg.ClearInlineRepos()
}
