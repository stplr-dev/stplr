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
	"github.com/stretchr/testify/mock"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type MockChecksExecutor struct {
	mock.Mock
}

func (m *MockChecksExecutor) RunChecks(ctx context.Context, pkg *staplerfile.Package, input *build.BuildInput) (bool, error) {
	args := m.Called(ctx, pkg, input)
	return args.Bool(0), args.Error(1)
}

func TestChecksStepRun(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		name        string
		runResults  map[string]bool
		expectedErr bool
	}

	tests := []testCase{
		{
			name: "all checks pass",
			runResults: map[string]bool{
				"pkg1": true,
				"pkg2": true,
			},
			expectedErr: false,
		},
		{
			name: "one check fails",
			runResults: map[string]bool{
				"pkg1": true,
				"pkg2": false,
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockExecutor := new(MockChecksExecutor)

			var pkgs []*staplerfile.Package
			for name, result := range tc.runResults {
				pkg := &staplerfile.Package{Name: name}
				pkgs = append(pkgs, pkg)
				mockExecutor.On("RunChecks", ctx, pkg, mock.Anything).Return(result, nil)
			}

			input := &build.BuildInput{
				Opts: &types.BuildOpts{Interactive: true},
			}

			state := build.NewBuildState()
			state.BasePackage = "base"
			state.Packages = pkgs
			state.Input = input

			step := build.ChecksStep(mockExecutor)
			err := step.Run(ctx, state)

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}
