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

func TestEmptyBuildDepends(t *testing.T) {
	_, err := reqprov.New(&distro.OSRelease{ID: "unknown"}, "deb", "")
	require.Error(t, err)
}

func TestAltLinuxBuildDepends(t *testing.T) {
	svc, err := reqprov.New(&distro.OSRelease{ID: altId}, "rpm", "")
	require.NoError(t, err)
	deps, err := svc.BuildDepends(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []string{altRpmBuild}, deps)
}

func TestFedoraBuildDependsByID(t *testing.T) {
	svc, err := reqprov.New(&distro.OSRelease{ID: fedoraId}, "rpm", "")
	require.NoError(t, err)
	deps, err := svc.BuildDepends(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []string{fedoraRpmBuild}, deps)
}

func TestFedoraBuildDependsByLike(t *testing.T) {
	svc, err := reqprov.New(&distro.OSRelease{ID: "mycustom", Like: []string{fedoraId}}, "rpm", "")
	require.NoError(t, err)
	deps, err := svc.BuildDepends(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, []string{fedoraRpmBuild}, deps)
}
