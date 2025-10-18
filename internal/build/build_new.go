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

	"go.stplr.dev/stplr/internal/app/output"
)

type Builder struct {
	scriptResolver       ScriptResolverExecutor
	scriptExecutor       ScriptExecutor
	cacheExecutor        CacheExecutor
	scriptViewerExecutor ScriptViewerExecutor
	installerExecutor    InstallerExecutor
	sourceExecutor       SourceDownloaderExecutor
	repos                PackageFinder
	nonfreeViewer        NonFreeViewerExecutor
	checksExecutor       ChecksExecutor
	out                  output.Output
}

func NewBuilder(
	scriptResolver ScriptResolverExecutor,
	scriptExecutor ScriptExecutor,
	cacheExecutor CacheExecutor,
	installerExecutor InstallerExecutor,
	sourceExecutor SourceDownloaderExecutor,
	checksExecutor ChecksExecutor,
	nonfreeViewer NonFreeViewerExecutor,
	repos PackageFinder,
	scriptViewerExecutor ScriptViewerExecutor,
) *Builder {
	return &Builder{
		scriptResolver:       scriptResolver,
		scriptExecutor:       scriptExecutor,
		cacheExecutor:        cacheExecutor,
		installerExecutor:    installerExecutor,
		sourceExecutor:       sourceExecutor,
		nonfreeViewer:        nonfreeViewer,
		checksExecutor:       checksExecutor,
		repos:                repos,
		scriptViewerExecutor: scriptViewerExecutor,
		out:                  output.NewConsoleOutput(),
	}
}

type BuildStep interface {
	Run(ctx context.Context, state *BuildState) error
}

func runSteps(ctx context.Context, state *BuildState, steps []BuildStep) ([]*BuiltDep, error) {
	for _, step := range steps {
		if err := step.Run(ctx, state); err != nil {
			return nil, fmt.Errorf("step (%T) failed: %w", step, err)
		}

		if state.ShouldExit {
			return state.BuiltDeps, nil
		}
	}
	return state.BuiltDeps, nil
}

func (b *Builder) BuildPackage(ctx context.Context, input *BuildInput) ([]*BuiltDep, error) {
	state := NewBuildState()
	state.Input = input

	steps := []BuildStep{
		ReadScriptStep(
			b.scriptExecutor,
			b.scriptExecutor,
		),
		CheckCacheStep(
			b.cacheExecutor,
		),
		ScriptViewStep(
			b.scriptViewerExecutor,
		),
		NonfreeViewStep(
			b.nonfreeViewer,
		),
		// Check arch and check is package already installed
		ChecksStep(
			b.checksExecutor,
		),
		// Flat packages dependencies in one slice
		FlatVarsStep(),
		// Add reqprov deps
		ModifyBuilddepsStep(),
		InstallDepsStep(
			b.installerExecutor,
			b.repos,
			b,
		),
		//
		PrepareStep(
			b.scriptExecutor,
			b.sourceExecutor,
		),
		BuildPackagesStep(
			b.scriptExecutor,
		),
		PostStep(
			b.installerExecutor,
		),
	}

	return runSteps(ctx, state, steps)
}
