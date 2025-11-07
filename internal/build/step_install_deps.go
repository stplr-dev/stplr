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
	"strings"

	"go.stplr.dev/stplr/internal/cliprompts"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/installer"
	"go.stplr.dev/stplr/internal/manager"
)

type installDepsStep struct {
	installerExecutor installer.InstallerExecutor
	repos             PackageFinder
	builder           *Builder
}

func InstallDepsStep(
	installerExecutor installer.InstallerExecutor,
	repos PackageFinder,
	builder *Builder,
) *installDepsStep {
	return &installDepsStep{
		installerExecutor: installerExecutor,
		repos:             repos,
		builder:           builder,
	}
}

func (s *installDepsStep) Run(ctx context.Context, state *BuildState) error {
	slog.Debug("installBuildDeps")
	alrBuildDeps, installedBuildDeps, err := s.installBuildDeps(ctx, state.Input, state.FlatVars.BuildDepends)
	if err != nil {
		return err
	}

	slog.Debug("installOptDeps")
	_, err = s.installOptDeps(ctx, state.Input, state.FlatVars.OptDepends)
	if err != nil {
		return err
	}

	depNames := make(map[string]struct{})
	for _, dep := range alrBuildDeps {
		depNames[dep.Name] = struct{}{}
	}

	// We filter so as not to re-build what has already been built at the `installBuildDeps` stage.
	var filteredDepends []string
	for _, d := range state.FlatVars.Depends {
		if _, found := depNames[d]; !found {
			filteredDepends = append(filteredDepends, d)
		}
	}

	slog.Debug("BuildALRDeps")
	newBuiltDeps, repoDeps, err := s.builder.BuildALRDeps(ctx, state.Input, filteredDepends)
	if err != nil {
		return err
	}

	state.InstalledBuildDeps = installedBuildDeps
	state.RepoDeps = repoDeps
	state.BuiltDeps = append(state.BuiltDeps, newBuiltDeps...)

	return nil
}

type InstallInput interface {
	commonbuild.OsInfoProvider
	commonbuild.BuildOptsProvider
	commonbuild.PkgFormatProvider
}

func (s *installDepsStep) installBuildDeps(ctx context.Context, input InstallInput, pkgs []string) ([]*commonbuild.BuiltDep, []string, error) {
	var builtDeps []*commonbuild.BuiltDep
	var deps []string
	var err error
	if len(pkgs) > 0 {
		deps, err = s.installerExecutor.RemoveAlreadyInstalled(ctx, pkgs)
		if err != nil {
			return nil, nil, err
		}

		builtDeps, err = s.installPkgs(ctx, input, deps)
		if err != nil {
			return nil, nil, err
		}
	}
	return builtDeps, deps, nil
}

func splitPkgAndDesc(pkgs []string) (names []string, mapping map[string]string) {
	mapping = make(map[string]string)
	for _, p := range pkgs {
		parts := strings.SplitN(p, ":", 2)
		name := strings.TrimSpace(parts[0])
		desc := ""
		if len(parts) > 1 {
			desc = strings.TrimSpace(parts[1])
		}
		names = append(names, name)
		mapping[name] = desc
	}
	return
}

func (i *installDepsStep) installOptDeps(ctx context.Context, input InstallInput, pkgs []string) ([]*commonbuild.BuiltDep, error) {
	var builtDeps []*commonbuild.BuiltDep

	namesOnly, descMap := splitPkgAndDesc(pkgs)

	optDeps, err := i.installerExecutor.RemoveAlreadyInstalled(ctx, namesOnly)
	if err != nil {
		return nil, err
	}

	if len(optDeps) == 0 {
		return builtDeps, nil
	}

	var optDepsWithDesc []string
	for _, name := range optDeps {
		desc := descMap[name]
		if desc != "" {
			optDepsWithDesc = append(optDepsWithDesc, fmt.Sprintf("%s: %s", name, desc))
		} else {
			optDepsWithDesc = append(optDepsWithDesc, name)
		}
	}

	optDeps, err = cliprompts.ChooseOptDepends(
		ctx,
		optDepsWithDesc,
		"install",
		input.BuildOpts().Interactive,
	)
	if err != nil {
		return nil, err
	}

	if len(optDeps) == 0 {
		return builtDeps, nil
	}

	builtDeps, err = i.installPkgs(ctx, input, optDeps)
	if err != nil {
		return nil, err
	}

	return builtDeps, nil
}

func (s *installDepsStep) installPkgs(ctx context.Context, input InstallInput, pkgs []string) ([]*commonbuild.BuiltDep, error) {
	builtDeps, repoDeps, err := s.builder.BuildALRDeps(ctx, input, pkgs)
	if err != nil {
		return nil, err
	}

	if len(builtDeps) > 0 {
		err = s.installerExecutor.InstallLocal(ctx, GetBuiltPaths(builtDeps), &manager.Opts{
			NoConfirm: !input.BuildOpts().Interactive,
		})
		if err != nil {
			return nil, err
		}
	}

	if len(repoDeps) > 0 {
		err = s.installerExecutor.Install(ctx, repoDeps, &manager.Opts{
			NoConfirm: !input.BuildOpts().Interactive,
		})
		if err != nil {
			return nil, err
		}
	}

	return builtDeps, nil
}
