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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type MockCacheExecutor struct {
	mock.Mock
}

func (m *MockCacheExecutor) CheckForBuiltPackage(ctx context.Context, input build.CheckForBuiltPackageInput, pkg *staplerfile.Package) (string, bool, error) {
	args := m.Called(ctx, input, pkg)
	path, _ := args.Get(0).(string)
	ok, _ := args.Get(1).(bool)
	return path, ok, args.Error(2)
}

func TestCheckCacheStep(t *testing.T) {
	ctx := context.Background()

	type cacheResult struct {
		pkg   *staplerfile.Package
		path  string
		found bool
		err   error
	}

	tests := []struct {
		name              string
		clean             bool
		cacheResults      []cacheResult
		expectedBuiltDeps []string
		expectExit        bool
		expectErr         bool
	}{
		{
			name:  "All packages found in cache",
			clean: false,
			cacheResults: []cacheResult{
				{&staplerfile.Package{Name: "pkg1"}, "/path/to/pkg1", true, nil},
				{&staplerfile.Package{Name: "pkg2"}, "/path/to/pkg2", true, nil},
			},
			expectedBuiltDeps: []string{"/path/to/pkg1", "/path/to/pkg2"},
			expectExit:        true,
		},
		{
			name:  "Some packages found in cache",
			clean: false,
			cacheResults: []cacheResult{
				{&staplerfile.Package{Name: "pkg1"}, "/path/to/pkg1", true, nil},
				{&staplerfile.Package{Name: "pkg2"}, "", false, nil},
			},
			expectedBuiltDeps: []string{"/path/to/pkg1"},
			expectExit:        false,
		},
		{
			name:  "Cache check error",
			clean: false,
			cacheResults: []cacheResult{
				{&staplerfile.Package{Name: "pkg1"}, "", false, errors.New("cache error")},
			},
			expectErr: true,
		},
		{
			name:  "Clean build, skips cache check",
			clean: true,
			cacheResults: []cacheResult{
				{&staplerfile.Package{Name: "pkg1"}, "", false, nil},
			},
			expectedBuiltDeps: nil,
			expectExit:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCache := new(MockCacheExecutor)
			step := build.CheckCacheStep(mockCache)

			packages := [](*staplerfile.Package){}
			for _, res := range tt.cacheResults {
				packages = append(packages, res.pkg)
				if !tt.clean {
					mockCache.On("CheckForBuiltPackage", ctx, mock.Anything, res.pkg).Return(res.path, res.found, res.err)
				}
			}

			state := &build.BuildState{
				Input: &commonbuild.BuildInput{
					Opts: &types.BuildOpts{Clean: tt.clean},
				},
				Packages: packages,
			}

			err := step.Run(ctx, state)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectExit, state.ShouldExit)

			var builtPaths []string
			for _, dep := range state.BuiltDeps {
				builtPaths = append(builtPaths, dep.Path)
			}
			assert.ElementsMatch(t, tt.expectedBuiltDeps, builtPaths)
		})
	}
}
