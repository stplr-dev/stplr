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
	"log/slog"

	"github.com/leonelquinteros/gotext"

	stdErrors "errors"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliprompts"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/types"
)

type builder interface {
	InstallPkgs(ctx context.Context, input build.InstallInput, pkgs []string) ([]*commonbuild.BuiltDep, error)
}

type useCase struct {
	builder builder
	mgr     manager.Manager
	info    *distro.OSRelease
}

func New(builder builder, mgr manager.Manager, info *distro.OSRelease) *useCase {
	return &useCase{
		builder: builder,
		mgr:     mgr,
		info:    info,
	}
}

type Options struct {
	Pkgs        []string
	Clean       bool
	Interactive bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	slog.Info("trying install", "pkgs", opts.Pkgs, "interactive", opts.Interactive)

	_, err := u.builder.InstallPkgs(
		ctx,
		&build.BuildArgs{
			Opts: &types.BuildOpts{
				Clean:       opts.Clean,
				Interactive: opts.Interactive,
			},
			Info:       u.info,
			PkgFormat_: build.GetPkgFormat(u.mgr),
		},
		opts.Pkgs,
	)
	if stdErrors.Is(err, build.ErrLicenseAgreementWasDeclined) {
		slog.Info("license agreement was declined", "pkgs", opts.Pkgs)
		return errors.NewI18nError(gotext.Get("License agreement was declined"))
	}
	if stdErrors.Is(err, cliprompts.ErrUserChoseNotContinue) {
		slog.Info("user chose not to continue after reading the script", "pkgs", opts.Pkgs)
		return errors.NewI18nError(gotext.Get("User chose not to continue after reading script"))
	}
	var ctxErr *build.BuildContextError
	if stdErrors.As(err, &ctxErr) {
		msg := gotext.Get(
			"Error when installing the package. Report the issue here: %s\nError trace",
			ctxErr.ReportUrl,
		)
		return errors.WrapIntoI18nError(ctxErr.Unwrap(), msg)
	}
	if err != nil {
		slog.Error("error when installing", "pkgs", opts.Pkgs, "err", err)
		return errors.WrapIntoI18nError(err, gotext.Get("Error when installing the package"))
	}

	return nil
}
