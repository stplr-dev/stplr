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

	"go.stplr.dev/stplr/internal/shutils/decoder"
	"go.stplr.dev/stplr/pkg/reqprov"
)

type modifyBuilddepsStep struct{}

func ModifyBuilddepsStep() *modifyBuilddepsStep {
	return &modifyBuilddepsStep{}
}

func (s *modifyBuilddepsStep) Name() string {
	return "modify build deps"
}

func (s *modifyBuilddepsStep) Run(ctx context.Context, state *BuildState) error {
	for _, vars := range state.Packages {
		if len(vars.AutoReq.Resolved()) == 1 && decoder.IsTruthy(vars.AutoReq.Resolved()[0]) {
			f, err := reqprov.New(
				state.Input.OSRelease(),
				state.Input.PkgFormat(),
				vars.AutoReqProvMethod.Resolved(),
			)
			if err != nil {
				return fmt.Errorf("failed to init reqprov: %w", err)
			}

			newBuildDeps, err := f.BuildDepends(ctx)
			if err != nil {
				return fmt.Errorf("failed to get build deps from provreq %w", err)
			}
			state.FlatVars.BuildDepends = append(state.FlatVars.BuildDepends, newBuildDeps...)
		}
	}

	state.FlatVars.BuildDepends = removeDuplicates(state.FlatVars.BuildDepends)

	return nil
}
