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
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"text/template"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/installer"
	"go.stplr.dev/stplr/internal/scripter"
)

type Builder struct {
	scriptResolver       ScriptResolverExecutor
	scriptExecutor       scripter.ScriptExecutor
	cacheExecutor        CacheExecutor
	scriptViewerExecutor ScriptViewerExecutor
	installerExecutor    installer.InstallerExecutor
	sourceExecutor       SourceDownloaderExecutor
	repos                PackageFinder
	nonfreeViewer        NonFreeViewerExecutor
	checksExecutor       ChecksExecutor
	out                  output.Output
}

func NewBuilder(
	scriptResolver ScriptResolverExecutor,
	scriptExecutor scripter.ScriptExecutor,
	cacheExecutor CacheExecutor,
	installerExecutor installer.InstallerExecutor,
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

func runSteps(ctx context.Context, state *BuildState, steps []BuildStep) ([]*commonbuild.BuiltDep, error) {
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

type BuildContextError struct {
	Repo      string
	Package   string
	ReportUrl string

	Err error
}

func (e *BuildContextError) Error() string {
	return fmt.Sprintf("build error for %s/%s: %s", e.Repo, e.Package, e.Err.Error())
}

func (e *BuildContextError) Unwrap() error {
	return e.Err
}

func (b *Builder) BuildPackage(ctx context.Context, input *commonbuild.BuildInput) ([]*commonbuild.BuiltDep, error) {
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

	res, stepsErr := runSteps(ctx, state, steps)
	if stepsErr != nil {
		if input.BasePkgName != "" {
			repo, err := b.repos.GetRepo(input.Repository())
			if err != nil {
				slog.Error("")
				return nil, fmt.Errorf("failed to get repo: %w", err)
			}

			if repo.ReportUrl == "" {
				return nil, stepsErr
			}

			tmpl, err := template.New("report-url").Parse(repo.ReportUrl)
			if err != nil {
				return nil, fmt.Errorf("invalid report url template: %w", err)
			}

			data := map[string]interface{}{
				"Repo":            repo.Name,
				"BasePackageName": input.BasePkgName,
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return nil, fmt.Errorf("failed to execute report url template: %w", err)
			}

			reportURL := buf.String()

			return nil, &BuildContextError{
				Repo:      repo.Name,
				Package:   input.BasePkgName,
				Err:       stepsErr,
				ReportUrl: reportURL,
			}
		}
		return nil, stepsErr
	}
	return res, nil
}
