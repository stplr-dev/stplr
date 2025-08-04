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

	"go.stplr.dev/stplr/pkg/staplerfile"
)

type ScriptViewerExecutor interface {
	ViewScript(ctx context.Context, input *BuildInput, sf *staplerfile.ScriptFile, basePkg string) error
}

type scriptViewStep struct {
	scriptViewer ScriptViewerExecutor
}

func ScriptViewStep(scriptViewer ScriptViewerExecutor) *scriptViewStep {
	return &scriptViewStep{scriptViewer: scriptViewer}
}

func (s *scriptViewStep) Run(ctx context.Context, state *BuildState) error {
	return s.scriptViewer.ViewScript(
		ctx,
		state.Input,
		state.ScriptFile,
		state.BasePackage,
	)
}
