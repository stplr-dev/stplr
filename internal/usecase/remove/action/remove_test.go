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

package action

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type fakeManager struct {
	manager.Manager
	installed map[string]bool
	removed   []string
}

func (m *fakeManager) IsInstalled(pkg string) (bool, error) {
	return m.installed[pkg], nil
}

func (m *fakeManager) Remove(_ *manager.Opts, pkgs ...string) error {
	m.removed = append(m.removed, pkgs...)
	return nil
}

type fakeFinder struct {
	found map[string][]staplerfile.Package
	err   error
}

func (f fakeFinder) FindPkgs(_ context.Context, _ []string) (map[string][]staplerfile.Package, []string, error) {
	return f.found, nil, f.err
}

func TestRun(t *testing.T) {
	const fullName = "yandex-browser+stplr-aides"

	tests := []struct {
		name      string
		arg       string
		found     map[string][]staplerfile.Package
		installed map[string]bool
		want      string
	}{
		{
			name: "removes installed Stapler package by provides using its full name",
			arg:  "yandex-browser-stable",
			found: map[string][]staplerfile.Package{
				"yandex-browser-stable": {{Name: "yandex-browser", Repository: "aides"}},
			},
			installed: map[string]bool{fullName: true},
			want:      fullName,
		},
		{
			name: "keeps system package name when matching Stapler package is not installed",
			arg:  "yandex-browser-stable",
			found: map[string][]staplerfile.Package{
				"yandex-browser-stable": {{Name: "yandex-browser", Repository: "aides"}},
			},
			want: "yandex-browser-stable",
		},
		{
			name:      "converts qualified Stapler package name",
			arg:       "aides/yandex-browser",
			installed: map[string]bool{fullName: true},
			want:      fullName,
		},
		{
			name: "keeps unknown system package name",
			arg:  "vim",
			want: "vim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &fakeManager{installed: tt.installed}
			uc := New(mgr, fakeFinder{found: tt.found})

			err := uc.Run(context.Background(), Options{Pkgs: []string{tt.arg}})

			require.NoError(t, err)
			assert.Equal(t, []string{tt.want}, mgr.removed)
		})
	}
}

func TestRunReturnsFindError(t *testing.T) {
	findErr := errors.New("find failed")
	mgr := &fakeManager{}
	uc := New(mgr, fakeFinder{err: findErr})

	err := uc.Run(context.Background(), Options{Pkgs: []string{"vim"}})

	require.ErrorIs(t, err, findErr)
	assert.Empty(t, mgr.removed)
}
