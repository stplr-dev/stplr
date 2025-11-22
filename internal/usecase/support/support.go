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

package support

import (
	"context"

	"github.com/leonelquinteros/gotext"
	"github.com/spf13/afero"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/app/output"
)

type useCase struct {
	out output.Output
}

func New(out output.Output) *useCase {
	return &useCase{
		out,
	}
}

func (u *useCase) Run(ctx context.Context) error {
	archivePath := "stplr-support.tar.gz"

	c := newArchiveCreator(afero.NewOsFs())
	if err := c.CreateSupportArchive(ctx, archivePath); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Failed to generate support archive"))
	}

	u.out.Info(gotext.Get("Support archive %s has been created. You can send it to whoever needs it.", archivePath))

	return nil
}
