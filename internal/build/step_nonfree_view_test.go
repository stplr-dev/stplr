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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type MockNonFreeViewerExecutor struct {
	mock.Mock
}

func (m *MockNonFreeViewerExecutor) ViewNonfree(ctx context.Context, pkg *staplerfile.Package, scriptPath string, interactive bool) error {
	args := m.Called(ctx, pkg, scriptPath, interactive)
	return args.Error(0)
}

func TestNonfreeViewStepRun(t *testing.T) {
	ctx := context.Background()

	mockExecutor := new(MockNonFreeViewerExecutor)

	pkg1 := &staplerfile.Package{Name: "pkg1"}
	pkg2 := &staplerfile.Package{Name: "pkg2"}

	mockExecutor.On("ViewNonfree", ctx, pkg1, "", true).Return(nil)
	mockExecutor.On("ViewNonfree", ctx, pkg2, "", true).Return(nil)

	state := build.NewBuildState()
	state.Packages = []*staplerfile.Package{
		pkg1,
		pkg2,
	}
	state.Input = &build.BuildInput{
		Opts: &types.BuildOpts{
			Interactive: true,
		},
	}
	state.ScriptFile, _ = staplerfile.ReadFromIOReader(strings.NewReader(""), "")

	step := build.NonfreeViewStep(mockExecutor)
	err := step.Run(ctx, state)

	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
}
