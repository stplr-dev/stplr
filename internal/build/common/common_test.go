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

package common_test

import (
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/internal/build/common"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/types"
)

func TestCreateBuildEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		info        *distro.OSRelease
		dirs        types.Directories
		expectedEnv map[string]string
	}{
		{
			name: "All fields populated",
			info: &distro.OSRelease{
				Name:       "Ubuntu",
				PrettyName: "Ubuntu 22.04 LTS",
				ID:         "ubuntu",
				VersionID:  "22.04",
				Like:       []string{"debian"},
			},
			dirs: types.Directories{
				PkgDir:  "/build/pkg",
				SrcDir:  "/build/src",
				HomeDir: "/build/home",
			},
			expectedEnv: map[string]string{
				"DISTRO_NAME":        "Ubuntu",
				"DISTRO_PRETTY_NAME": "Ubuntu 22.04 LTS",
				"DISTRO_ID":          "ubuntu",
				"DISTRO_VERSION_ID":  "22.04",
				"DISTRO_ID_LIKE":     "debian",
				"pkgdir":             "/build/pkg",
				"srcdir":             "/build/src",
				"ARCH":               "amd64",
				"NCPU":               strconv.Itoa(runtime.NumCPU()),
				"HOME":               "/build/home",
			},
		},
		{
			name: "No dirs provided",
			info: &distro.OSRelease{
				Name:       "Alpine",
				PrettyName: "Alpine Linux",
				ID:         "alpine",
				VersionID:  "3.18",
				Like:       []string{"musl"},
			},
			dirs: types.Directories{},
			expectedEnv: map[string]string{
				"DISTRO_NAME":        "Alpine",
				"DISTRO_PRETTY_NAME": "Alpine Linux",
				"DISTRO_ID":          "alpine",
				"DISTRO_VERSION_ID":  "3.18",
				"DISTRO_ID_LIKE":     "musl",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := common.CreateBuildEnvVars(tt.info, tt.dirs)

			// Convert the resulting env slice to a map for easier assertion
			envMap := make(map[string]string)
			for _, kv := range envVars {
				parts := strings.SplitN(kv, "=", 2)
				if len(parts) == 2 {
					envMap[parts[0]] = parts[1]
				}
			}

			for key, expected := range tt.expectedEnv {
				actual, exists := envMap[key]
				assert.True(t, exists, "Expected key %q to exist in env", key)
				assert.Equal(t, expected, actual, "Mismatch for env var %q", key)
			}
		})
	}
}
