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
	"errors"
	"log/slog"

	"github.com/leonelquinteros/gotext"
)

type flatVarsStep struct{}

func FlatVarsStep() *flatVarsStep { return &flatVarsStep{} }

func (s *flatVarsStep) Run(ctx context.Context, state *BuildState) error {
	buildDepends := []string{}
	optDepends := []string{}
	depends := []string{}
	sources := []string{}
	checksums := []string{}
	for _, pkg := range state.Packages {
		buildDepends = append(buildDepends, pkg.BuildDepends.Resolved()...)
		optDepends = append(optDepends, pkg.OptDepends.Resolved()...)
		depends = append(depends, pkg.Depends.Resolved()...)
		sources = append(sources, pkg.Sources.Resolved()...)
		checksums = append(checksums, pkg.Checksums.Resolved()...)
	}
	if len(sources) != len(checksums) {
		slog.Error(gotext.Get("The checksums array must be the same length as sources"))
		return errors.New("exit")
	}
	sources, checksums = removeDuplicatesSources(sources, checksums)

	state.FlatVars.Sources = sources
	state.FlatVars.Checksums = checksums
	state.FlatVars.BuildDepends = removeDuplicates(buildDepends)
	state.FlatVars.OptDepends = removeDuplicates(optDepends)
	state.FlatVars.Depends = removeDuplicates(depends)

	return nil
}
