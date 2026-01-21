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
	"net/url"
	"os"
	"path/filepath"

	"github.com/jeandeaual/go-locale"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/build/common"
	"go.stplr.dev/stplr/internal/shutils/decoder"
	"go.stplr.dev/stplr/internal/shutils/handlers"
	"go.stplr.dev/stplr/internal/shutils/helpers"
	"go.stplr.dev/stplr/internal/shutils/runner"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/dl"
	"go.stplr.dev/stplr/pkg/types"
)

type ScriptFile struct {
	file *syntax.File
	path string
}

type parseOptions struct {
	language string
}

type parseOption func(*parseOptions)

func WithCustomLanguage(lang string) parseOption {
	return func(os *parseOptions) {
		os.language = lang
	}
}

func (s *ScriptFile) ParseBuildVars(ctx context.Context, info *distro.OSRelease, packages []string, opts ...parseOption) (string, []*Package, error) {
	options := &parseOptions{}
	for _, opt := range opts {
		opt(options)
	}

	r, err := s.createRunner(info)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create runner: %w", err)
	}

	if err := runner.EnableStrictShellMode(ctx, r); err != nil {
		return "", nil, fmt.Errorf("failed to enable strict shell mode: %w", err)
	}

	if err := runScript(ctx, r, s.file); err != nil {
		return "", nil, fmt.Errorf("failed to run script: %w", err)
	}

	dec, err := newDecoder(info, r, options)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get decoder: %w", err)
	}

	pkgNames, err := ParseNames(dec)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse names: %w", err)
	}

	if len(pkgNames.Names) == 0 {
		return "", nil, errors.New("package name is missing")
	}

	targetPackages := packages
	if len(targetPackages) == 0 {
		targetPackages = pkgNames.Names
	}

	scriptOpts, err := ParseScriptOptions(dec)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse names: %w", err)
	}

	pkgs, err := s.createPackagesForBuildVars(ctx, dec, info, pkgNames, targetPackages)
	if err != nil {
		return "", nil, fmt.Errorf("failed to createPackagesForBuildVars: %w", err)
	}

	for _, p := range pkgs {
		p.Options = scriptOpts
	}

	baseName := pkgNames.BasePkgName
	if len(pkgNames.Names) == 1 {
		baseName = pkgNames.Names[0]
	}

	return baseName, pkgs, nil
}

func (s *ScriptFile) createRunner(info *distro.OSRelease) (*interp.Runner, error) {
	scriptDir := filepath.Dir(s.path)
	env := common.CreateBuildEnvVars(info, types.Directories{})

	restr := handlers.WithFilter(
		handlers.RestrictSandbox(scriptDir),
	)

	return interp.New(
		interp.Env(expand.ListEnviron(env...)),
		interp.StdIO(os.Stdin, os.Stderr, os.Stderr),
		interp.ExecHandler(helpers.Restricted.ExecHandler(handlers.NopExec)),
		interp.ReadDirHandler2(handlers.RestrictedReadDir(restr)),
		interp.StatHandler(handlers.RestrictedStat(restr)),
		interp.OpenHandler(handlers.RestrictedOpen(restr)),
		interp.Dir(scriptDir),
	)
}

func (s *ScriptFile) createPackagesForBuildVars(
	ctx context.Context,
	dec *decoder.Decoder,
	info *distro.OSRelease,
	pkgNames *PackageNames,
	targetPackages []string,
) ([]*Package, error) {
	var varsOfPackages []*Package

	if len(pkgNames.Names) == 1 {
		var pkg Package
		pkg.Name = pkgNames.Names[0]
		pkg.Options = &ScriptOptions{}
		if err := dec.DecodeVars(&pkg); err != nil {
			return nil, fmt.Errorf("failed to decode vars: %w", err)
		}
		varsOfPackages = append(varsOfPackages, &pkg)
		return varsOfPackages, nil
	}

	for _, pkgName := range targetPackages {
		pkg, err := s.createPackageFromMeta(ctx, dec, info, pkgName, pkgNames.BasePkgName)
		if err != nil {
			return nil, fmt.Errorf("failed to createPackageFromMeta: %w", err)
		}
		varsOfPackages = append(varsOfPackages, pkg)
	}

	return varsOfPackages, nil
}

func (s *ScriptFile) createPackageFromMeta(
	ctx context.Context,
	dec *decoder.Decoder,
	info *distro.OSRelease,
	pkgName, basePkgName string,
) (*Package, error) {
	funcName := fmt.Sprintf("meta_%s", pkgName)
	meta, ok := dec.GetFuncWithSubshell(funcName)

	var pkg Package
	if ok {
		metaRunner, err := meta(ctx)
		if err != nil {
			return nil, err
		}
		metaDecoder := decoder.New(info, metaRunner)
		if err := metaDecoder.DecodeVars(&pkg); err != nil {
			return nil, err
		}
	} else {
		if err := dec.DecodeVars(&pkg); err != nil {
			return nil, fmt.Errorf("failed to decode vars: %w", err)
		}
	}
	pkg.Name = pkgName
	pkg.BasePkgName = basePkgName
	pkg.Options = &ScriptOptions{}
	return &pkg, nil
}

func runScript(ctx context.Context, runner *interp.Runner, fl *syntax.File) error {
	runner.Reset()
	return runner.Run(ctx, fl)
}

func newDecoder(info *distro.OSRelease, runner *interp.Runner, options *parseOptions) (*decoder.Decoder, error) {
	d := decoder.New(info, runner)

	var systemLang string
	var err error

	if options.language == "" {
		systemLang, err = locale.GetLanguage()
		if err != nil {
			return nil, fmt.Errorf("cant get systemlang: %w", err)
		}
		if systemLang == "" || systemLang == "C" {
			systemLang = "en"
		}
	} else {
		systemLang = options.language
	}

	d.OverridesOpts = d.OverridesOpts.WithLanguages([]string{systemLang})
	return d, nil
}

func (a *ScriptFile) Path() string {
	return a.path
}

func (a *ScriptFile) File() *syntax.File {
	return a.file
}

func (a *ScriptFile) ExternalFiles(ctx context.Context, info *distro.OSRelease) ([]string, error) {
	_, pkgs, err := a.ParseBuildVars(ctx, info, nil)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, pkg := range pkgs {
		out = append(out,
			getScriptFiles(pkg)...,
		)
		out = append(out,
			getFireJailProfilePaths(pkg)...,
		)
		out = append(out,
			getNonFreeMsgFilePaths(pkg)...,
		)

		sourceFiles, err := getLocalSourceFiles(pkg)
		if err != nil {
			return nil, err
		}
		out = append(out, sourceFiles...)
	}

	return out, nil
}

func getScriptFiles(pkg *Package) []string {
	var files []string
	for _, scripts := range pkg.Scripts.All() {
		files = append(files, scripts.Files()...)
	}
	return files
}

func getFireJailProfilePaths(pkg *Package) []string {
	var files []string
	for _, profiles := range pkg.FireJailProfiles.All() {
		for _, profilePath := range profiles {
			files = append(files, profilePath)
		}
	}
	return files
}

func getNonFreeMsgFilePaths(pkg *Package) []string {
	var files []string
	for _, msgfile := range pkg.NonFreeMsgFile.All() {
		files = append(files, msgfile)
	}
	return files
}

func getLocalSourceFiles(pkg *Package) ([]string, error) {
	var files []string
	for _, sources := range pkg.Sources.All() {
		for _, src := range sources {
			u, err := url.Parse(src)
			if err != nil {
				return nil, err
			}
			if dl.IsLocalUrl(u) {
				files = append(files, u.Path)
			}
		}
	}
	return files, nil
}
