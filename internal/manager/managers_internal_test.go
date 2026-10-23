// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package manager

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/pkg/distro"
)

type fakeManager struct {
	Manager
	name   string
	exists bool
}

func (f *fakeManager) Name() string { return f.name }
func (f *fakeManager) Exists() bool { return f.exists }

func TestDetect(t *testing.T) {
	all := func(names ...string) []Manager {
		var out []Manager
		for _, n := range []string{NamePacman, NameAPTRpm, NameAPT, NameDNF, NameYUM, NameAPK, NameZypper, NameEpm} {
			out = append(out, &fakeManager{name: n, exists: slices.Contains(names, n)})
		}
		return out
	}

	tests := []struct {
		name     string
		info     *distro.OSRelease
		existing []string
		want     string
	}{
		{"alt with pacman game", &distro.OSRelease{ID: "altlinux"}, []string{NamePacman, NameAPTRpm, NameEpm}, NameAPTRpm},
		{"alt without apt-rpm", &distro.OSRelease{ID: "altlinux"}, []string{NamePacman, NameEpm}, NameEpm},
		{"manjaro via like", &distro.OSRelease{ID: "manjaro", Like: []string{"arch"}}, []string{NamePacman, NameAPT}, NamePacman},
		{"ubuntu derivative", &distro.OSRelease{ID: "linuxmint", Like: []string{"ubuntu", "debian"}}, []string{NamePacman, NameAPT}, NameAPT},
		{"fedora without dnf", &distro.OSRelease{ID: "fedora"}, []string{NamePacman, NameYUM}, NameYUM},
		{"unknown distro falls back", &distro.OSRelease{ID: "foo"}, []string{NameAPT, NameZypper}, NameAPT},
		{"no os-release falls back", nil, []string{NameZypper}, NameZypper},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detect(tt.info, all(tt.existing...))
			if assert.NotNil(t, got) {
				assert.Equal(t, tt.want, got.Name())
			}
		})
	}

	assert.Nil(t, detect(&distro.OSRelease{ID: "altlinux"}, all()))
}
