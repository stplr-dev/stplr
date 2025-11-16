// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
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

package scripter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"go.stplr.dev/stplr/internal/cpu"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

func TestGetBasePkgInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const testPackage = "test-package"

	tests := []struct {
		name          string
		pkg           *staplerfile.Package
		osRelease     *distro.OSRelease
		repository    string
		buildOpts     *types.BuildOpts
		expectedName  string
		expectedArch  string
		expectedEpoch string
	}{
		{
			name: "with suffix",
			pkg: &staplerfile.Package{
				Name:    testPackage,
				Version: "1.0.0",
				Release: 1,
				Epoch:   2,
			},
			osRelease: &distro.OSRelease{
				ID: "ubuntu",
			},
			repository: "main",
			buildOpts: &types.BuildOpts{
				NoSuffix: false,
			},
			expectedName:  "test-package+stplr-main",
			expectedArch:  cpu.Arch(),
			expectedEpoch: "2",
		},
		{
			name: "without suffix",
			pkg: &staplerfile.Package{
				Name:    testPackage,
				Version: "1.0.0",
				Release: 1,
				Epoch:   3,
			},
			osRelease: &distro.OSRelease{
				ID: "ubuntu",
			},
			repository: "main",
			buildOpts: &types.BuildOpts{
				NoSuffix: true,
			},
			expectedName:  "test-package",
			expectedArch:  cpu.Arch(),
			expectedEpoch: "3",
		},
		{
			name: "altlinux release format",
			pkg: &staplerfile.Package{
				Name:    testPackage,
				Version: "1.0.0",
				Release: 5,
				Epoch:   1,
			},
			osRelease: &distro.OSRelease{
				ID: "altlinux",
			},
			repository: "main",
			buildOpts: &types.BuildOpts{
				NoSuffix: false,
			},
			expectedName:  "test-package+stplr-main",
			expectedArch:  cpu.Arch(),
			expectedEpoch: "1",
		},
		{
			name: "fedora release format",
			pkg: &staplerfile.Package{
				Name:    testPackage,
				Version: "1.0.0",
				Release: 2,
				Epoch:   0,
			},
			osRelease: &distro.OSRelease{
				ID:         "fedora",
				PlatformID: "platform:f36",
			},
			repository: "updates",
			buildOpts: &types.BuildOpts{
				NoSuffix: false,
			},
			expectedName:  "test-package+stplr-updates",
			expectedArch:  cpu.Arch(),
			expectedEpoch: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInput := NewMockgetBasePkgInfoInput(ctrl)

			mockInput.EXPECT().OSRelease().Return(tt.osRelease).AnyTimes()
			mockInput.EXPECT().Repository().Return(tt.repository).AnyTimes()
			mockInput.EXPECT().BuildOpts().Return(tt.buildOpts).AnyTimes()

			// Call the function with mocked dependencies
			result := GetBasePkgInfo(tt.pkg, mockInput)

			require.NotNil(t, result, "result should not be nil")
			assert.Equal(t, tt.expectedName, result.Name, "package name should match")
			assert.Equal(t, tt.expectedArch, result.Arch, "architecture should match")
			assert.Equal(t, tt.pkg.Version, result.Version, "version should match")
			assert.Equal(t, tt.expectedEpoch, result.Epoch, "epoch should match")
			assert.NotEmpty(t, result.Release, "release should not be empty")
		})
	}
}

func TestGetBasePkgInfoNilInputs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInput := NewMockgetBasePkgInfoInput(ctrl)
	mockInput.EXPECT().OSRelease().Return(&distro.OSRelease{ID: "ubuntu"}).AnyTimes()
	mockInput.EXPECT().Repository().Return("main").AnyTimes()
	mockInput.EXPECT().BuildOpts().Return(&types.BuildOpts{NoSuffix: false}).AnyTimes()

	// Test with nil package - this would panic in real code, but let's make sure our test handles it gracefully
	assert.Panics(t, func() {
		GetBasePkgInfo(nil, mockInput)
	}, "should panic with nil package")
}
