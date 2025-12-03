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
	"log/slog"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/scripter"
)

type prepareStep struct {
	scriptExecutor   scripter.ScriptExecutor
	sourceDownloader SourceDownloaderExecutor
	out              output.Output
}

func PrepareStep(
	scriptExecutor scripter.ScriptExecutor,
	sourceDownloader SourceDownloaderExecutor,
) *prepareStep {
	return &prepareStep{
		scriptExecutor:   scriptExecutor,
		sourceDownloader: sourceDownloader,
		out:              output.NewConsoleOutput(),
	}
}

func (s *prepareStep) Name() string {
	return "prepare"
}

func (b *prepareStep) Run(ctx context.Context, state *BuildState) error {
	slog.Debug("PrepareDirs")
	err := b.scriptExecutor.PrepareDirs(ctx, state.Input, state.BasePackage)
	if err != nil {
		return err
	}

	b.out.Info(gotext.Get("Downloading sources"))

	err = b.sourceDownloader.DownloadSources(
		ctx,
		state.Input,
		state.BasePackage,
		SourcesInput{
			Sources:   state.FlatVars.Sources,
			Checksums: state.FlatVars.Checksums,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
