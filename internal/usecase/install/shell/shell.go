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

package shell

import (
	"context"
	"fmt"
	"os"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type PackageGetter interface {
	GetPkgs(ctx context.Context, where string, args ...any) ([]staplerfile.Package, error)
}

type useCase struct {
	pkgGetter PackageGetter
	stdout    *os.File
}

func New(pkgGetter PackageGetter) *useCase {
	return &useCase{
		pkgGetter: pkgGetter,
		stdout:    os.Stdout,
	}
}

func (u *useCase) Run(ctx context.Context) error {
	result, err := u.pkgGetter.GetPkgs(ctx, "true")
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error getting packages"))
	}
	for _, pkg := range result {
		fmt.Fprintln(u.stdout, pkg.Name)
	}
	return nil
}
