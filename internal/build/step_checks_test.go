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
	"go.uber.org/mock/gomock"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/mocks"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

func TestChecksStepRun(t *testing.T) {
	ctx := context.Background()

	type result struct {
		name  string
		value bool
	}

	type testCase struct {
		name        string
		runResults  []result
		expectedErr bool
	}

	tests := []testCase{
		{
			name: "all checks pass",
			runResults: []result{
				{"pkg1", true},
				{"pkg2", true},
			},
			expectedErr: false,
		},
		{
			name: "one check fails",
			runResults: []result{
				{"pkg1", true},
				{"pkg2", false},
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockExecutor := mocks.NewMockChecksExecutor(ctrl)

			var pkgs []*staplerfile.Package
			for _, r := range tc.runResults {
				pkg := &staplerfile.Package{Name: r.name}
				pkgs = append(pkgs, pkg)
			}

			for _, r := range tc.runResults {
				mockExecutor.EXPECT().
					RunChecks(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p *staplerfile.Package, _ *build.BuildInput) (bool, error) {
						if p.Name == r.name {
							return r.value, nil
						}
						return true, nil
					})

				if !r.value {
					break
				}
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
		})
	}
}
