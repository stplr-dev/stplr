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

package action

import (
	"context"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/manager"
)

type useCase struct {
	mgr manager.Manager
}

func New(manager manager.Manager) *useCase {
	return &useCase{mgr: manager}
}

type Options struct {
	Pkgs        []string
	Interactive bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	if err := u.mgr.Remove(&manager.Opts{
		NoConfirm: !opts.Interactive,
	}, opts.Pkgs...); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error removing packages"))
	}

	return nil
}
