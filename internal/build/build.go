// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
// Copyright (C) 2025 The ALR Authors
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
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type BuildInput struct {
	Opts        *types.BuildOpts
	Info_       *distro.OSRelease
	PkgFormat_  string
	Script      string
	Repository_ string
	Packages_   []string
}

func (bi *BuildInput) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	if err := encoder.Encode(bi.Opts); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Info_); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.PkgFormat_); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Script); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Repository_); err != nil {
		return nil, err
	}
	if err := encoder.Encode(bi.Packages_); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func (bi *BuildInput) GobDecode(data []byte) error {
	r := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(r)

	if err := decoder.Decode(&bi.Opts); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Info_); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.PkgFormat_); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Script); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Repository_); err != nil {
		return err
	}
	if err := decoder.Decode(&bi.Packages_); err != nil {
		return err
	}

	return nil
}

func (b *BuildInput) Repository() string {
	return b.Repository_
}

func (b *BuildInput) BuildOpts() *types.BuildOpts {
	return b.Opts
}

func (b *BuildInput) OSRelease() *distro.OSRelease {
	return b.Info_
}

func (b *BuildInput) PkgFormat() string {
	return b.PkgFormat_
}

func (b *BuildInput) Packages() []string {
	return b.Packages_
}

type BuildOptsProvider interface {
	BuildOpts() *types.BuildOpts
}

type OsInfoProvider interface {
	OSRelease() *distro.OSRelease
}

type PkgFormatProvider interface {
	PkgFormat() string
}

type RepositoryProvider interface {
	Repository() string
}

// ================================================

func Map[T, R any](items []T, f func(T) R) []R {
	res := make([]R, len(items))
	for i, item := range items {
		res[i] = f(item)
	}
	return res
}

func GetBuiltPaths(deps []*BuiltDep) []string {
	return Map(deps, func(dep *BuiltDep) string {
		return dep.Path
	})
}

func GetBuiltName(deps []*BuiltDep) []string {
	return Map(deps, func(dep *BuiltDep) string {
		return dep.Name
	})
}

type PackageFinder interface {
	FindPkgs(ctx context.Context, pkgs []string) (map[string][]staplerfile.Package, []string, error)
}

type Config interface {
	GetPaths() *config.Paths
	PagerStyle() string
}

type BuiltDep struct {
	Name string
	Path string
}

type OSReleaser interface {
	OSRelease() *distro.OSRelease
}

type PkgFormatter interface {
	PkgFormat() string
}

type RepositoryGetter interface {
	Repository() string
}

type SourcesInput struct {
	Sources   []string
	Checksums []string
}

type FunctionsOutput struct {
	Contents *[]string
}

type BuildArgs struct {
	Opts       *types.BuildOpts
	Info       *distro.OSRelease
	PkgFormat_ string
}

func (b *BuildArgs) BuildOpts() *types.BuildOpts {
	return b.Opts
}

func (b *BuildArgs) OSRelease() *distro.OSRelease {
	return b.Info
}

func (b *BuildArgs) PkgFormat() string {
	return b.PkgFormat_
}

type BuildPackageFromDbArgs struct {
	BuildArgs
	Package  *staplerfile.Package
	Packages []string
}

type BuildPackageFromScriptArgs struct {
	BuildArgs
	Script   string
	Packages []string
}

func (b *Builder) BuildPackageFromDb(
	ctx context.Context,
	args *BuildPackageFromDbArgs,
) ([]*BuiltDep, error) {
	scriptInfo := b.scriptResolver.ResolveScript(ctx, args.Package)

	return b.BuildPackage(ctx, &BuildInput{
		Script:      scriptInfo.Script,
		Repository_: scriptInfo.Repository,
		Packages_:   args.Packages,
		PkgFormat_:  args.PkgFormat(),
		Opts:        args.Opts,
		Info_:       args.Info,
	})
}

func (b *Builder) BuildPackageFromScript(
	ctx context.Context,
	args *BuildPackageFromScriptArgs,
) ([]*BuiltDep, error) {
	return b.BuildPackage(ctx, &BuildInput{
		Script:      args.Script,
		Repository_: "default",
		Packages_:   args.Packages,
		PkgFormat_:  args.PkgFormat(),
		Opts:        args.Opts,
		Info_:       args.Info,
	})
}

type InstallPkgsArgs struct {
	BuildArgs
	AlrPkgs    []staplerfile.Package
	NativePkgs []string
}

func (b *Builder) InstallALRPackages(
	ctx context.Context,
	input InstallInput,
	pkgs []staplerfile.Package,
) error {
	for _, pkg := range pkgs {
		res, err := b.BuildPackageFromDb(
			ctx,
			&BuildPackageFromDbArgs{
				Package:  &pkg,
				Packages: []string{},
				BuildArgs: BuildArgs{
					Opts:       input.BuildOpts(),
					Info:       input.OSRelease(),
					PkgFormat_: input.PkgFormat(),
				},
			},
		)
		if err != nil {
			return err
		}

		err = b.installerExecutor.InstallLocal(
			ctx,
			GetBuiltPaths(res),
			&manager.Opts{
				NoConfirm: !input.BuildOpts().Interactive,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Builder) InstallPkgs(
	ctx context.Context,
	input InstallInput,
	pkgs []string,
) ([]*BuiltDep, error) {
	builtDeps, repoDeps, err := i.BuildALRDeps(ctx, input, pkgs)
	if err != nil {
		return nil, err
	}

	if len(builtDeps) > 0 {
		err = i.installerExecutor.InstallLocal(ctx, GetBuiltPaths(builtDeps), &manager.Opts{
			NoConfirm: !input.BuildOpts().Interactive,
		})
		if err != nil {
			return nil, err
		}
	}

	if len(repoDeps) > 0 {
		err = i.installerExecutor.Install(ctx, repoDeps, &manager.Opts{
			NoConfirm: !input.BuildOpts().Interactive,
		})
		if err != nil {
			return nil, err
		}
	}

	return builtDeps, nil
}

func (b *Builder) BuildALRDeps(ctx context.Context, input InstallInput, depends []string) (buildDeps []*BuiltDep, repoDeps []string, err error) {
	if len(depends) > 0 {
		slog.Info(gotext.Get("Installing dependencies"))

		found, notFound, err := b.repos.FindPkgs(ctx, depends)
		if err != nil {
			return nil, nil, fmt.Errorf("failed FindPkgs: %w", err)
		}
		repoDeps = notFound

		pkgs := cliutils.FlattenPkgs(
			ctx,
			found,
			"install",
			input.BuildOpts().Interactive,
		)
		pkgsMap := groupPackages(pkgs)

		for basePkgName := range pkgsMap {
			pkg := pkgsMap[basePkgName].pkg
			res, err := b.BuildPackageFromDb(
				ctx,
				&BuildPackageFromDbArgs{
					Package:  pkg,
					Packages: pkgsMap[basePkgName].packages,
					BuildArgs: BuildArgs{
						Opts:       input.BuildOpts(),
						Info:       input.OSRelease(),
						PkgFormat_: input.PkgFormat(),
					},
				},
			)
			if err != nil {
				return nil, nil, fmt.Errorf("failed build package from db: %w", err)
			}

			buildDeps = append(buildDeps, res...)
		}
	}

	repoDeps = removeDuplicates(repoDeps)
	buildDeps = removeDuplicates(buildDeps)

	return buildDeps, repoDeps, nil
}

type pkgItem struct {
	pkg      *staplerfile.Package
	packages []string
}

func groupPackages(pkgs []staplerfile.Package) map[string]pkgItem {
	pkgsMap := make(map[string]pkgItem)
	for _, pkg := range pkgs {
		name := pkg.BasePkgName
		if name == "" {
			name = pkg.Name
		}
		it, ok := pkgsMap[name]
		if !ok {
			copy := pkg
			it = pkgItem{pkg: &copy}
		}
		it.packages = append(it.packages, pkg.Name)
		pkgsMap[name] = it // be sure to overwrite it!
	}
	return pkgsMap
}
