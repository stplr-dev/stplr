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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/pkg/types"
)

func ptr[T any](v T) *T { return &v }

func TestApplyOverrideDisabledOnly(t *testing.T) {
	base := types.Repo{Name: "test", URL: "https://example.com", Disabled: false}
	result := types.ApplyOverride(base, types.RepoOverride{Disabled: ptr(true)})

	assert.True(t, result.Disabled)
	assert.Equal(t, base.URL, result.URL)
	assert.Equal(t, base.Name, result.Name)
}

func TestApplyOverrideMultipleFields(t *testing.T) {
	base := types.Repo{Name: "test", URL: "https://old.com", Ref: "main"}
	result := types.ApplyOverride(base, types.RepoOverride{
		URL: ptr("https://new.com"),
		Ref: ptr("v2"),
	})

	assert.Equal(t, "https://new.com", result.URL)
	assert.Equal(t, "v2", result.Ref)
	assert.Equal(t, "test", result.Name)
}

func TestApplyOverrideNilFields(t *testing.T) {
	base := types.Repo{Name: "test", Disabled: true, Ref: "main"}
	result := types.ApplyOverride(base, types.RepoOverride{})

	assert.True(t, result.Disabled)
	assert.Equal(t, "main", result.Ref)
}

func TestApplyOverrideEmptySliceVsNil(t *testing.T) {
	base := types.Repo{
		Mirrors: []string{"https://mirror1.com", "https://mirror2.com"},
	}

	result := types.ApplyOverride(base, types.RepoOverride{})
	assert.Equal(t, base.Mirrors, result.Mirrors, "nil Mirrors in override should not clear base Mirrors")

	result2 := types.ApplyOverride(base, types.RepoOverride{Mirrors: []string{}})
	assert.Empty(t, result2.Mirrors, "empty Mirrors in override should clear base Mirrors")
}

func TestApplyOverrideDoesNotMutateBase(t *testing.T) {
	base := types.Repo{Name: "test", Disabled: false}
	_ = types.ApplyOverride(base, types.RepoOverride{Disabled: ptr(true)})

	assert.False(t, base.Disabled, "ApplyOverride must not mutate the base repo")
}
