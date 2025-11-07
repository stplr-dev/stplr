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

	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type checkCacheStep struct {
	cacheExecutor CacheExecutor
}

func CheckCacheStep(cacheExecutor CacheExecutor) *checkCacheStep {
	return &checkCacheStep{cacheExecutor: cacheExecutor}
}

func (s *checkCacheStep) Run(ctx context.Context, state *BuildState) error {
	if !state.Input.Opts.Clean {
		var remaining []*staplerfile.Package
		for _, pkg := range state.Packages {
			builtPkgPath, ok, err := s.cacheExecutor.CheckForBuiltPackage(ctx, state.Input, pkg)
			if err != nil {
				return err
			}
			if ok {
				state.BuiltDeps = append(state.BuiltDeps, &commonbuild.BuiltDep{
					Path: builtPkgPath,
				})
			} else {
				remaining = append(remaining, pkg)
			}
		}

		if len(remaining) == 0 {
			state.ShouldExit = true
			return nil
		}
	}

	return nil
}
