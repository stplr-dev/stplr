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

package build_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/build"

	"go.stplr.dev/stplr/pkg/staplerfile"
)

func TestFlatVarsStepRunSuccess(t *testing.T) {
	step := build.FlatVarsStep()
	ctx := context.Background()

	pkgs := []*staplerfile.Package{
		{
			BuildDepends: staplerfile.OverridableFromMap(map[string][]string{
				"": {"make", "gcc"},
			}),
			OptDepends: staplerfile.OverridableFromMap(map[string][]string{
				"": {"graphviz"},
			}),
			Depends: staplerfile.OverridableFromMap(map[string][]string{
				"": {"libc"},
			}),
			Sources: staplerfile.OverridableFromMap(map[string][]string{
				"": {"src.tar.gz", "helper.tar.gz"},
			}),
			Checksums: staplerfile.OverridableFromMap(map[string][]string{
				"": {"abc123", "def456"},
			}),
		},
		{
			BuildDepends: staplerfile.OverridableFromMap(map[string][]string{
				"": {"make"}, // duplicate
			}),
			OptDepends: staplerfile.OverridableFromMap(map[string][]string{
				"": {"dot"}, // unique
			}),
			Depends: staplerfile.OverridableFromMap(map[string][]string{
				"": {"libc"}, // duplicate
			}),
			Sources: staplerfile.OverridableFromMap(map[string][]string{
				"": {"helper.tar.gz"}, // duplicate
			}),
			Checksums: staplerfile.OverridableFromMap(map[string][]string{
				"": {"def456"}, // duplicate
			}),
		},
	}

	for _, pkg := range pkgs {
		staplerfile.ResolvePackage(pkg, []string{""})
	}

	state := build.NewBuildState()
	state.Packages = pkgs

	err := step.Run(ctx, state)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"src.tar.gz", "helper.tar.gz"}, state.FlatVars.Sources)
	assert.ElementsMatch(t, []string{"abc123", "def456"}, state.FlatVars.Checksums)
	assert.ElementsMatch(t, []string{"make", "gcc"}, state.FlatVars.BuildDepends)
	assert.ElementsMatch(t, []string{"graphviz", "dot"}, state.FlatVars.OptDepends)
	assert.ElementsMatch(t, []string{"libc"}, state.FlatVars.Depends)
}

func TestFlatVarsStepRunChecksumMismatch(t *testing.T) {
	step := build.FlatVarsStep()
	ctx := context.Background()

	pkgs := []*staplerfile.Package{
		{
			Sources: staplerfile.OverridableFromMap(map[string][]string{
				"": {"src.tar.gz", "helper.tar.gz"},
			}),
			Checksums: staplerfile.OverridableFromMap(map[string][]string{
				"": {"abc123"},
			}),
		},
	}

	for _, pkg := range pkgs {
		staplerfile.ResolvePackage(pkg, []string{""})
	}

	state := build.NewBuildState()
	state.Packages = pkgs

	err := step.Run(ctx, state)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksums array must be the same length")
}
