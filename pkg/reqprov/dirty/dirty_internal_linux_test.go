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

package dirty

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsELF(t *testing.T) {
	assert.True(t, isELF("test_data/binary"))
	assert.False(t, isELF("test_data/binary.c"))
}

func TestProcessElf(t *testing.T) {
	provSet := make(map[string]struct{})
	needSet := make(map[string]struct{})

	require.NoError(t, processELF("test_data/binary", provSet, needSet))
	require.Equal(t, map[string]struct{}{
		"binary()(64bit)": {},
	}, provSet)
	require.Equal(t, map[string]struct{}{
		"libreadline.so.8()(64bit)": {},
		"libc.so.6()(64bit)":        {},
	}, needSet)

	require.NoError(t, processELF("test_data/libtest.so", provSet, needSet))
	require.Equal(t, map[string]struct{}{
		"libtest.so()(64bit)": {},
		"binary()(64bit)":     {},
	}, provSet)
	require.Equal(t, map[string]struct{}{
		"libreadline.so.8()(64bit)": {},
		"libc.so.6()(64bit)":        {},
		"libm.so.6()(64bit)":        {},
	}, needSet)
}
