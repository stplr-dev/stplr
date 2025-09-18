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
	"io/fs"
	"os"
	"path/filepath"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/osutils"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type stepInputState struct {
	opts Options
	deps *appbuilder.AppDeps
	wd   string

	origUid int
	origGid int

	packages   []string
	scriptArgs *build.BuildPackageFromScriptArgs
	dbArgs     *build.BuildPackageFromDbArgs
}

type stepRuntimeState struct {
	installer build.InstallerExecutor
	scripter  build.ScriptExecutor
	copier    build.ScriptCopier
	cleanups  []func()
}

type stepOutputState struct {
	out []*build.BuiltDep
}

type stepState struct {
	input   stepInputState
	runtime stepRuntimeState
	output  stepOutputState
}

func (s *stepState) Cleanup() {
	for i := len(s.runtime.cleanups) - 1; i >= 0; i-- {
		s.runtime.cleanups[i]()
	}
}

type step interface {
	Execute(ctx context.Context, state *stepState) error
}

type checkStep struct{}

func (s *checkStep) Execute(ctx context.Context, state *stepState) error {
	if state.input.deps.Cfg.ForbidBuildCommand() {
		return errors.NewI18nError(gotext.Get("Your settings do not allow build command"))
	}

	if err := utils.EnsureIsPrivilegedGroupMemberOrRoot(); err != nil {
		return err
	}

	return nil
}

type prepareScriptArgs struct{}

func (s *prepareScriptArgs) Execute(ctx context.Context, state *stepState) error {
	input := &state.input

	script, err := filepath.Abs(input.opts.Script)
	if err != nil {
		return fmt.Errorf("cannot get absolute script path: %w", err)
	}

	subpackage := input.opts.Subpackage

	if subpackage != "" {
		input.packages = []string{subpackage}
	}

	input.scriptArgs = &build.BuildPackageFromScriptArgs{
		Script:    script,
		Packages:  input.packages,
		BuildArgs: buildArgsFromOptions(input.opts, input.deps),
	}

	return nil
}

type PackageFinder interface {
	FindPkgs(ctx context.Context, names []string) (map[string][]staplerfile.Package, []string, error)
}

type prepareDbArgs struct {
	finder PackageFinder
}

func (s *prepareDbArgs) Execute(ctx context.Context, state *stepState) error {
	input := &state.input

	pkgs, _, err := s.finder.FindPkgs(ctx, []string{input.opts.Package})
	if err != nil {
		return fmt.Errorf("failed to find pkgs: %w", err)
	}

	pkg := cliutils.FlattenPkgs(ctx, pkgs, "build", input.opts.Interactive)
	if pkg[0].BasePkgName != "" {
		input.packages = append(input.packages, pkg[0].Name)
	}

	input.dbArgs = &build.BuildPackageFromDbArgs{
		Package:   &pkg[0],
		Packages:  input.packages,
		BuildArgs: buildArgsFromOptions(input.opts, input.deps),
	}

	return nil
}

type setupCopier struct{}

func (s *setupCopier) Execute(ctx context.Context, state *stepState) error {
	runtime := &state.runtime
	copier, cleanup, err := build.GetSafeScriptCopier()
	if err != nil {
		return fmt.Errorf("failed to init copier: %w", err)
	}
	runtime.copier = copier
	runtime.cleanups = append(runtime.cleanups, cleanup)
	return nil
}

type copyScript struct {
	fsys fs.FS
}

func (s *copyScript) Execute(ctx context.Context, state *stepState) error {
	input := &state.input
	runtime := &state.runtime

	file, err := staplerfile.ReadFromFS(s.fsys, input.scriptArgs.Script)
	if err != nil {
		return err
	}

	input.scriptArgs.Script, err = runtime.copier.Copy(ctx, file, input.deps.Info)
	if err != nil {
		return err
	}

	runtime.cleanups = append(runtime.cleanups, func() {
		err = os.RemoveAll(filepath.Dir(input.scriptArgs.Script))
		if err != nil {
			panic(err)
		}
	})

	return nil
}

type modifyCfgPaths struct{}

func (s *modifyCfgPaths) Execute(ctx context.Context, state *stepState) error {
	return config.PatchToUserDirs(state.input.deps.Cfg)
}

type buildStep struct{}

func (s *buildStep) Execute(ctx context.Context, state *stepState) error {
	input := &state.input
	runtime := &state.runtime

	builder, err := build.NewMainBuilder(
		input.deps.Cfg,
		input.deps.Manager,
		input.deps.Repos,
		runtime.scripter,
		runtime.installer,
	)
	if err != nil {
		return err
	}

	var pkgs []*build.BuiltDep
	switch {
	case input.scriptArgs != nil:
		pkgs, err = builder.BuildPackageFromScript(
			ctx,
			input.scriptArgs,
		)
	case input.dbArgs != nil:
		pkgs, err = builder.BuildPackageFromDb(
			ctx,
			input.dbArgs,
		)
	}
	if err != nil {
		return err
	}

	state.output.out = pkgs

	return nil
}

type prepareInstallerAndScripterStep struct{}

func (s *prepareInstallerAndScripterStep) Execute(ctx context.Context, state *stepState) error {
	res, cleanup, err := build.PrepareInstallerAndScripter()
	if err != nil {
		return err
	}
	state.runtime.installer = res.Installer
	state.runtime.scripter = res.Scripter
	state.runtime.cleanups = append(state.runtime.cleanups, cleanup)

	return nil
}

type copyOutExecutor func(ctx context.Context, pkg *build.BuiltDep, state *stepState) error

type copyOutStep struct {
	copy copyOutExecutor
}

func copyOutViaCopier(ctx context.Context, pkg *build.BuiltDep, state *stepState) error {
	name := filepath.Base(pkg.Path)

	return state.runtime.copier.CopyOut(
		ctx,
		pkg.Path,
		filepath.Join(state.input.wd, name),
		state.input.origUid,
		state.input.origGid,
	)
}

func copyOutViaOsutils(ctx context.Context, pkg *build.BuiltDep, state *stepState) error {
	name := filepath.Base(pkg.Path)
	return osutils.CopyFile(pkg.Path, filepath.Join(state.input.wd, name))
}

func (s *copyOutStep) Execute(ctx context.Context, state *stepState) error {
	for _, pkg := range state.output.out {
		if err := s.copy(ctx, pkg, state); err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error moving the package"))
		}
	}
	return nil
}

func buildArgsFromOptions(opts Options, deps *appbuilder.AppDeps) build.BuildArgs {
	return build.BuildArgs{
		Opts: &types.BuildOpts{
			Clean:       opts.Clean,
			Interactive: opts.Interactive,
		},
		PkgFormat_: build.GetPkgFormat(deps.Manager),
		Info:       deps.Info,
	}
}
