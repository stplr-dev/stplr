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

package reqprov_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/reqprov"
)

const (
	altId          = "altlinux"
	fedoraId       = "fedora"
	altRpmBuild    = "rpm-build"
	fedoraRpmBuild = altRpmBuild
)

func TestBuildDepends(t *testing.T) {
	tests := []struct {
		name      string
		osRelease *distro.OSRelease
		pkgFormat string
		buildOpts string
		wantDeps  []string
		wantErr   bool
	}{
		{
			name:      "empty build depends - unknown distro",
			osRelease: &distro.OSRelease{ID: "unknown"},
			pkgFormat: "deb",
			buildOpts: "",
			wantDeps:  nil,
			wantErr:   true,
		},
		{
			name:      "altlinux by ID",
			osRelease: &distro.OSRelease{ID: altId},
			pkgFormat: "rpm",
			buildOpts: "",
			wantDeps:  []string{altRpmBuild},
			wantErr:   false,
		},
		{
			name:      "fedora by ID",
			osRelease: &distro.OSRelease{ID: fedoraId},
			pkgFormat: "rpm",
			buildOpts: "",
			wantDeps:  []string{fedoraRpmBuild},
			wantErr:   false,
		},
		{
			name:      "fedora by Like",
			osRelease: &distro.OSRelease{ID: "mycustom", Like: []string{fedoraId}},
			pkgFormat: "rpm",
			buildOpts: "",
			wantDeps:  []string{fedoraRpmBuild},
			wantErr:   false,
		},
		{
			name:      "altlinux by Like",
			osRelease: &distro.OSRelease{ID: "mycustom", Like: []string{altId}},
			pkgFormat: "rpm",
			buildOpts: "",
			wantDeps:  []string{altRpmBuild},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := reqprov.New(tt.osRelease, tt.pkgFormat, tt.buildOpts)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			deps, err := svc.BuildDepends(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.wantDeps, deps)
		})
	}
}
