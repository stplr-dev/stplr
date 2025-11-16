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

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/service/repos"
)

type resetter interface {
	Reset(ctx context.Context) error
}

type ReposGetter func() (*repos.Repos, deps.Cleanup, error)

type useCase struct {
	dbResetter    resetter
	cacheResetter resetter
	reposGetter   ReposGetter
}

func New(dbResetter, cacheResetter resetter, reposGetter ReposGetter) *useCase {
	return &useCase{dbResetter, cacheResetter, reposGetter}
}

func (u *useCase) Run(ctx context.Context) error {
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
