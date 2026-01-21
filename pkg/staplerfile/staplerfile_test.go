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

package staplerfile_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

func TestParseBuildVarsWithCustomLanguage(t *testing.T) {
	r := strings.NewReader(`name=test
	desc='english'
	desc_ru='russian'
	desc_ja='japanese'
	`)
	s, err := staplerfile.ReadFromIOReader(r, "Staplerfile")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	_, pkgs, err := s.ParseBuildVars(t.Context(), &distro.OSRelease{}, []string{}, staplerfile.WithCustomLanguage("ja"))
	assert.NoError(t, err)
	assert.NotNil(t, pkgs)

	assert.Len(t, pkgs, 1)
	assert.Equal(t, pkgs[0].Name, "test")
	assert.Equal(t, pkgs[0].Description.Resolved(), "japanese")
}

func TestParseBuildVarsOptions(t *testing.T) {
	r := strings.NewReader(`sfe_249_new_extractor=1
	name=test
	`)
	s, err := staplerfile.ReadFromIOReader(r, "Staplerfile")
	assert.NoError(t, err)
	assert.NotNil(t, s)

	_, pkgs, err := s.ParseBuildVars(t.Context(), &distro.OSRelease{}, []string{}, staplerfile.WithCustomLanguage("ja"))
	assert.NoError(t, err)
	assert.NotNil(t, pkgs)

	assert.Len(t, pkgs, 1)
	assert.Equal(t, pkgs[0].Options.SFE249NewExtractor, true)
}
