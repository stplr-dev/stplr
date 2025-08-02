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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffSets(t *testing.T) {
	assert.Equal(t, map[string]struct{}{"a": {}}, DiffSets(
		map[string]struct{}{"a": {}},
		map[string]struct{}{"b": {}},
	))

	assert.Equal(t, map[string]struct{}{"a": {}}, DiffSets(
		map[string]struct{}{"a": {}},
		map[string]struct{}{},
	))

	assert.Equal(t, map[string]struct{}{}, DiffSets(
		map[string]struct{}{"a": {}},
		map[string]struct{}{
			"a": {},
			"b": {},
		},
	))
}

func TestToSet(t *testing.T) {
	input := []string{"a", "b", "a"}
	expected := map[string]struct{}{
		"a": {},
		"b": {},
	}

	result := ToSet(input)
	assert.Equal(t, expected, result)
}
