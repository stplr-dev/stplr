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

	"go.stplr.dev/stplr/internal/scripter"
)

type readScriptStep struct {
	reader scripter.ScriptReader
	parser scripter.PackagesParser
}

func ReadScriptStep(
	reader scripter.ScriptReader,
	parser scripter.PackagesParser,
) *readScriptStep {
	return &readScriptStep{
		reader: reader,
		parser: parser,
	}
}

func (s *readScriptStep) Name() string {
	return "read script"
}

func (s *readScriptStep) Run(ctx context.Context, state *BuildState) error {
	f, err := s.reader.Read(ctx, state.Input.Script)
	if err != nil {
		return err
	}

	state.BasePackage, state.Packages, err = s.parser.ParsePackages(ctx, f, state.Input.Packages(), *state.Input.OSRelease())
	if err != nil {
		return err
	}
	state.Version = state.Packages[0].Version
	state.ScriptFile = f
	return nil
}
