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

package build

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	stdErrors "errors"

	"github.com/leonelquinteros/gotext"
	"github.com/spf13/afero"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliprompts"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/copier"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type useCase struct {
	config  *config.ALRConfig
	builder *build.Builder
	info    *distro.OSRelease
	copier  copier.CopierExecutor
	manager manager.Manager
	finder  build.PackageFinder
	fsys    afero.Fs

	cleanups []func()
}

type ConstructOptions struct {
	Config  *config.ALRConfig
	Builder *build.Builder
	Info    *distro.OSRelease
	Copier  copier.CopierExecutor
	Manager manager.Manager
	Finder  build.PackageFinder
}

func New(o ConstructOptions) *useCase {
	return &useCase{
		config:  o.Config,
		builder: o.Builder,
		copier:  o.Copier,
		info:    o.Info,
		manager: o.Manager,
		finder:  o.Finder,
		fsys:    afero.NewOsFs(),
	}
}

type RunOptions struct {
	Subpackage  string
	Directory   string
	Clean       bool
	Interactive bool
	NoSuffix    bool

	Script  string
	Package string
}

func (u *useCase) Run(ctx context.Context, o RunOptions) error {
	if err := u.checks(); err != nil {
		return err
	}

	var err error
	var pkgs []*commonbuild.BuiltDep

	switch {
	case o.Package != "":
		pkgs, err = u.runForDb(ctx, o)
	case o.Script != "":
		pkgs, err = u.runForScript(ctx, o)
	default:
		return fmt.Errorf("either Script or Package must be specified")
	}
	var ctxErr *build.BuildContextError
	if stdErrors.As(err, &ctxErr) {
		msg := gotext.Get(
			"Error when building the package. Report the issue here: %s\nError trace",
			ctxErr.ReportUrl,
		)
		return errors.WrapIntoI18nError(ctxErr.Unwrap(), msg)
	}
	if err != nil {
		return err
	}

	builtPkgs := make([]commonbuild.BuiltDep, len(pkgs))
	for i, pkg := range pkgs {
		builtPkgs[i] = commonbuild.BuiltDep{Name: pkg.Name, Path: pkg.Path}
	}

	if err := u.copier.CopyOut(ctx, builtPkgs); err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error moving the package"))
	}

	return nil
}

func (u *useCase) checks() error {
	if u.config.ForbidBuildCommand() {
		return errors.NewI18nError(gotext.Get("Your settings do not allow build command"))
	}

	if err := cliutils.EnsureIsPrivilegedGroupMemberOrRoot(); err != nil {
		return err
	}

	return nil
}

func (u *useCase) runForDb(ctx context.Context, o RunOptions) ([]*commonbuild.BuiltDep, error) {
	foundPkgs, _, err := u.finder.FindPkgs(ctx, []string{o.Package})
	if err != nil {
		return nil, fmt.Errorf("failed to find pkgs: %w", err)
	}

	pkg := cliprompts.FlattenPkgs(ctx, foundPkgs, "build", o.Interactive)
	var packages []string
	if pkg[0].BasePkgName != "" {
		packages = append(packages, pkg[0].Name)
	}

	return u.builder.BuildPackageFromDb(
		ctx,
		&build.BuildPackageFromDbArgs{
			Package:  &pkg[0],
			Packages: packages,
			BuildArgs: build.BuildArgs{
				Opts: &types.BuildOpts{
					Clean:       o.Clean,
					Interactive: o.Interactive,
					NoSuffix:    o.NoSuffix,
				},
				PkgFormat_: build.GetPkgFormat(u.manager),
				Info:       u.info,
			},
		},
	)
}

func (u *useCase) runForScript(ctx context.Context, o RunOptions) ([]*commonbuild.BuiltDep, error) {
	file, err := staplerfile.ReadFromAferoFS(u.fsys, o.Script)
	if err != nil {
		return nil, err
	}

	script, err := u.copier.Copy(ctx, file, u.info)
	if err != nil {
		return nil, err
	}

	u.cleanups = append(u.cleanups, func() {
		err = os.RemoveAll(filepath.Dir(script))
		if err != nil {
			slog.Warn("failed to cleanup", "err", err)
		}
	})

	var packages []string
	if o.Subpackage != "" {
		packages = []string{o.Subpackage}
	}

	return u.builder.BuildPackageFromScript(
		ctx,
		&build.BuildPackageFromScriptArgs{
			Script:   script,
			Packages: packages,
			BuildArgs: build.BuildArgs{
				Opts: &types.BuildOpts{
					Clean:       o.Clean,
					Interactive: o.Interactive,
					NoSuffix:    o.NoSuffix,
				},
				PkgFormat_: build.GetPkgFormat(u.manager),
				Info:       u.info,
			},
		},
	)
}
