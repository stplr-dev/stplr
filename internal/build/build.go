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
	"context"
	"fmt"
	"log/slog"

	"github.com/gobwas/glob"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/cliprompts"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/scripter"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

// ================================================

func GetBuiltPaths(deps []*commonbuild.BuiltDep) []string {
	return scripter.Map(deps, func(dep *commonbuild.BuiltDep) string {
		return dep.Path
	})
}

type PackageFinder interface {
	FindPkgs(ctx context.Context, pkgs []string) (map[string][]staplerfile.Package, []string, error)
	GetRepo(name string) (types.Repo, error)
}

type SourcesInput struct {
	Sources      []string
	Checksums    []string
	NewExtractor bool
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
) ([]*commonbuild.BuiltDep, error) {
	scriptInfo := b.scriptResolver.ResolveScript(ctx, args.Package)

	// very dirty but ok
	// TODO: refactor logic
	name := args.Package.BasePkgName
	if name == "" {
		name = args.Package.Name
	}

	r := staplerfile.NewResolver(args.Info)
	err := r.Init()
	if err != nil {
		return nil, err
	}
	r.Resolve(args.Package)

	if isFirejailExcluded(args.Package, b.cfg, b.out) {
		args.Opts.DisableFirejail = true
	}

	return b.BuildPackage(ctx, &commonbuild.BuildInput{
		BasePkgName: name,
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
) ([]*commonbuild.BuiltDep, error) {
	return b.BuildPackage(ctx, &commonbuild.BuildInput{
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
) ([]*commonbuild.BuiltDep, error) {
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

func (b *Builder) BuildALRDeps(ctx context.Context, input InstallInput, depends []string) (buildDeps []*commonbuild.BuiltDep, repoDeps []string, err error) {
	if err != nil {
		return nil, nil, fmt.Errorf("failed init resolver: %w", err)
	}

	if len(depends) > 0 {
		b.out.Info(gotext.Get("Installing dependencies"))

		found, notFound, err := b.repos.FindPkgs(ctx, depends)
		if err != nil {
			return nil, nil, fmt.Errorf("failed FindPkgs: %w", err)
		}
		repoDeps = notFound

		pkgs := cliprompts.FlattenPkgs(
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

func firejailedPatternMatch(fullName, pattern string) (bool, error) {
	g, err := glob.Compile(pattern)
	if err != nil {
		return false, err
	}
	return g.Match(fullName), nil
}

func isFirejailExcluded(pkg *staplerfile.Package, cfg commonbuild.Config, out output.Output) bool {
	if pkg.FireJailed.Resolved() {
		disableFirejail := false
		disabledPattern := ""

		fullName := pkg.FormatFullName()

		for _, pattern := range cfg.FirejailExclude() {
			matched, err := firejailedPatternMatch(fullName, pattern)
			if err != nil {
				slog.Debug("failed to match pattern", "err", err)
				continue
			}
			if matched {
				disableFirejail = true
				disabledPattern = pattern
				break
			}
		}

		if disableFirejail && !cfg.HideFirejailExcludeWarning() {
			out.Warn(gotext.Get(
				"Firejail is disabled for %q package due to ignore pattern %q in config. Security isolation will not be applied. Ensure you understand the risks.",
				fullName, disabledPattern,
			))
		}

		return disableFirejail
	}

	return false
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
