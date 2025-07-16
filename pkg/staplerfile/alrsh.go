// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
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

package staplerfile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/cpu"
	"go.stplr.dev/stplr/internal/shutils/decoder"
	"go.stplr.dev/stplr/internal/shutils/handlers"
	"go.stplr.dev/stplr/internal/shutils/helpers"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/types"
)

type ScriptFile struct {
	file *syntax.File
	path string
}

func createBuildEnvVars(info *distro.OSRelease, dirs types.Directories) []string {
	env := os.Environ()

	env = append(
		env,
		"DISTRO_NAME="+info.Name,
		"DISTRO_PRETTY_NAME="+info.PrettyName,
		"DISTRO_ID="+info.ID,
		"DISTRO_VERSION_ID="+info.VersionID,
		"DISTRO_ID_LIKE="+strings.Join(info.Like, " "),
		"ARCH="+cpu.Arch(),
		"NCPU="+strconv.Itoa(runtime.NumCPU()),
	)

	if dirs.ScriptDir != "" {
		env = append(env, "scriptdir="+dirs.ScriptDir)
	}

	if dirs.PkgDir != "" {
		env = append(env, "pkgdir="+dirs.PkgDir)
	}

	if dirs.SrcDir != "" {
		env = append(env, "srcdir="+dirs.SrcDir)
	}

	return env
}

func (s *ScriptFile) ParseBuildVars(ctx context.Context, info *distro.OSRelease, packages []string) (string, []*Package, error) {
	runner, err := s.createRunner(info)
	if err != nil {
		return "", nil, err
	}

	if err := runScript(ctx, runner, s.file); err != nil {
		return "", nil, err
	}

	dec := newDecoder(info, runner)

	pkgNames, err := ParseNames(dec)
	if err != nil {
		return "", nil, err
	}

	if len(pkgNames.Names) == 0 {
		return "", nil, errors.New("package name is missing")
	}

	targetPackages := packages
	if len(targetPackages) == 0 {
		targetPackages = pkgNames.Names
	}

	varsOfPackages, err := s.createPackagesForBuildVars(ctx, dec, pkgNames, targetPackages)
	if err != nil {
		return "", nil, err
	}

	baseName := pkgNames.BasePkgName
	if len(pkgNames.Names) == 1 {
		baseName = pkgNames.Names[0]
	}

	return baseName, varsOfPackages, nil
}

func (s *ScriptFile) createRunner(info *distro.OSRelease) (*interp.Runner, error) {
	scriptDir := filepath.Dir(s.path)
	env := createBuildEnvVars(info, types.Directories{ScriptDir: scriptDir})

	return interp.New(
		interp.Env(expand.ListEnviron(env...)),
		interp.StdIO(os.Stdin, os.Stderr, os.Stderr),
		interp.ExecHandler(helpers.Restricted.ExecHandler(handlers.NopExec)),
		interp.ReadDirHandler2(handlers.RestrictedReadDir(scriptDir)),
		interp.StatHandler(handlers.RestrictedStat(scriptDir)),
		interp.OpenHandler(handlers.RestrictedOpen(scriptDir)),
		interp.Dir(scriptDir),
	)
}

func (s *ScriptFile) createPackagesForBuildVars(
	ctx context.Context,
	dec *decoder.Decoder,
	pkgNames *PackageNames,
	targetPackages []string,
) ([]*Package, error) {
	var varsOfPackages []*Package

	if len(pkgNames.Names) == 1 {
		var pkg Package
		pkg.Name = pkgNames.Names[0]
		if err := dec.DecodeVars(&pkg); err != nil {
			return nil, err
		}
		varsOfPackages = append(varsOfPackages, &pkg)
		return varsOfPackages, nil
	}

	for _, pkgName := range targetPackages {
		pkg, err := s.createPackageFromMeta(ctx, dec, pkgName, pkgNames.BasePkgName)
		if err != nil {
			return nil, err
		}
		varsOfPackages = append(varsOfPackages, pkg)
	}

	return varsOfPackages, nil
}

func (s *ScriptFile) createPackageFromMeta(
	ctx context.Context,
	dec *decoder.Decoder,
	pkgName, basePkgName string,
) (*Package, error) {
	funcName := fmt.Sprintf("meta_%s", pkgName)
	meta, ok := dec.GetFuncWithSubshell(funcName)
	if !ok {
		return nil, fmt.Errorf("func %s is missing", funcName)
	}

	metaRunner, err := meta(ctx)
	if err != nil {
		return nil, err
	}

	metaDecoder := decoder.New(&distro.OSRelease{}, metaRunner)

	var vars Package
	if err := metaDecoder.DecodeVars(&vars); err != nil {
		return nil, err
	}

	vars.Name = pkgName
	vars.BasePkgName = basePkgName

	return &vars, nil
}

func runScript(ctx context.Context, runner *interp.Runner, fl *syntax.File) error {
	runner.Reset()
	return runner.Run(ctx, fl)
}

func newDecoder(info *distro.OSRelease, runner *interp.Runner) *decoder.Decoder {
	d := decoder.New(info, runner)
	// d.Overrides = false
	// d.LikeDistros = false
	return d
}

func (a *ScriptFile) Path() string {
	return a.path
}

func (a *ScriptFile) File() *syntax.File {
	return a.file
}
