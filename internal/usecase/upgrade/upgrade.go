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

package upgrade

import (
	"context"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/service/updater"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type builder interface {
	InstallALRPackages(
		ctx context.Context,
		args build.InstallInput,
		pkgs []staplerfile.Package,
	) error
}

type useCase struct {
	builder builder
	mgr     manager.Manager
	db      *db.Database
	info    *distro.OSRelease
	upd     *updater.Updater

	out output.Output
}

func New(builder builder, upd *updater.Updater, manager manager.Manager, db *db.Database, info *distro.OSRelease, out output.Output) *useCase {
	return &useCase{
		builder: builder,
		mgr:     manager,
		db:      db,
		info:    info,
		upd:     upd,
		out:     out,
	}
}

type Options struct {
	Pkgs        []string
	Clean       bool
	Interactive bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	updates, err := u.upd.CheckForUpdates(ctx)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error checking for updates"))
	}

	if len(updates) > 0 {
		err = u.builder.InstallALRPackages(
			ctx,
			&build.BuildArgs{
				Opts: &types.BuildOpts{
					Clean:       opts.Clean,
					Interactive: opts.Interactive,
				},
				Info:       u.info,
				PkgFormat_: build.GetPkgFormat(u.mgr),
			},
			mapUptatesInfoToPackages(updates),
		)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error checking for updates"))
		}
	} else {
		u.out.Info(gotext.Get("There is nothing to do."))
	}

	return nil
}

func mapUptatesInfoToPackages(updates []updater.UpdateInfo) []staplerfile.Package {
	var pkgs []staplerfile.Package
	for _, info := range updates {
		pkgs = append(pkgs, *info.Package)
	}
	return pkgs
}
