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

package distro

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOSReleaseFromFile(t *testing.T) {
	tests := []struct {
		fixture  string
		expected OSRelease
	}{
		{
			fixture: "debian-13",
			expected: OSRelease{
				ID:        "debian",
				VersionID: "13",
				ReleaseID: "trixie",
			},
		},
		{
			fixture: "debian-12",
			expected: OSRelease{
				ID:        "debian",
				VersionID: "12",
				ReleaseID: "bookworm",
			},
		},
		{
			fixture: "fedora-41",
			expected: OSRelease{
				ID:        "fedora",
				VersionID: "41",
				ReleaseID: "41",
			},
		},
		{
			fixture: "fedora-42",
			expected: OSRelease{
				ID:        "fedora",
				VersionID: "42",
				ReleaseID: "42",
			},
		},
		{
			fixture: "registry.altlinux.org-p11-alt-latest",
			expected: OSRelease{
				ID:        "altlinux",
				VersionID: "11",
				ReleaseID: "p11",
			},
		},
		{
			fixture: "registry.altlinux.org-sisyphus-alt-latest",
			expected: OSRelease{
				ID:        "altlinux",
				VersionID: "20250612",
				ReleaseID: "sisyphus",
			},
		},
		{
			fixture: "unknown-stapler-distro",
			expected: OSRelease{
				ID:        "unknown-stapler-distro",
				VersionID: "11.1",
				ReleaseID: "11",
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			f, err := os.Open(filepath.Join("tests-fixtures", tt.fixture))
			assert.NoError(t, err)
			defer f.Close()

			actual, err := parseOSReleaseFromFile(ctx, f)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected.ID, actual.ID, "ID mismatch")
			assert.Equal(t, tt.expected.ReleaseID, actual.ReleaseID, "ReleaseID mismatch")
		})
	}
}
