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
	"go.stplr.dev/stplr/internal/cliprompts"
	"go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/service/repos"
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
	repos   *repos.Repos

	out output.Output
}

func New(builder builder, upd *updater.Updater, manager manager.Manager, db *db.Database, repos *repos.Repos, info *distro.OSRelease, out output.Output) *useCase {
	return &useCase{
		builder: builder,
		mgr:     manager,
		db:      db,
		info:    info,
		upd:     upd,
		repos:   repos,
		out:     out,
	}
}

type Options struct {
	Pkgs        []string
	Clean       bool
	Interactive bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	err := u.repos.PullAll(ctx)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error pulling repositories"))
	}

	updates, err := u.upd.CheckForUpdates(ctx)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error checking for updates"))
	}

	if len(updates) == 0 {
		u.out.Info(gotext.Get("There is nothing to do."))
		return nil
	}

	var (
		succeeded []string
		failed    []struct {
			pkg string
			err error
		}
	)

	for i, update := range updates {
		pkgName := update.Package.Name
		u.out.Info(gotext.Get("Upgrading %d/%d: %s", i+1, len(updates), pkgName))

		if ctx.Err() != nil {
			u.out.Warn(gotext.Get("Stopping upgrade process"))
			break
		}

		pkgCtx, cancel := context.WithCancel(ctx)

		errChan := make(chan error, 1)

		go func() {
			errChan <- u.builder.InstallALRPackages(
				pkgCtx,
				&build.BuildArgs{
					Opts: &types.BuildOpts{
						Clean:       opts.Clean,
						Interactive: opts.Interactive,
					},
					Info:       u.info,
					PkgFormat_: build.GetPkgFormat(u.mgr),
				},
				[]staplerfile.Package{*update.Package},
			)
		}()

		select {
		case err := <-errChan:
			cancel()

			if err != nil {
				u.out.Warn(gotext.Get("Failed to upgrade %s: %v", pkgName, err))
				failed = append(failed, struct {
					pkg string
					err error
				}{pkgName, err})

				if i+1 != len(updates) {
					stopAll, _ := cliprompts.YesNoPrompt(context.Background(), gotext.Get("Do you want to stop the entire upgrade process?"), opts.Interactive, false)
					if stopAll {
						u.out.Info(gotext.Get("Stopping upgrade process"))
						break
					}
					u.out.Info(gotext.Get("Continuing with next package"))
				}
			} else {
				u.out.Info(gotext.Get("Successfully upgraded %s", pkgName))
				succeeded = append(succeeded, pkgName)
			}

		case <-ctx.Done():
			cancel()
			u.out.Warn(gotext.Get("Upgrade process interrupted"))
			<-errChan
			break
		}
	}

	u.printSummary(succeeded, failed)

	if len(failed) > 0 {
		return errors.NewI18nError(gotext.Get("Some packages failed to upgrade"))
	}

	return nil
}

func (u *useCase) printSummary(succeeded []string, failed []struct {
	pkg string
	err error
},
) {
	if len(succeeded) > 0 {
		u.out.Info(gotext.GetN(
			"Successfully upgraded: %d package",
			"Successfully upgraded: %d packages",
			len(succeeded),
			len(succeeded),
		))
		for _, pkg := range succeeded {
			u.out.Info("  - %s", pkg)
		}
	}

	if len(failed) > 0 {
		u.out.Warn(gotext.GetN(
			"Failed to upgrade: %d package",
			"Failed to upgrade: %d packages",
			len(failed),
			len(failed),
		))
		for _, f := range failed {
			u.out.Warn("  - %s: %v", f.pkg, f.err)
		}
	}
}
