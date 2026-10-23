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

package subid

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	input := strings.Join([]string{
		"# comment",
		"",
		"alice:100000:65536",
		"bob:0200000:08",
		"broken line",
		"x:abc:1",
		"y:1:2:3",
		"z:1:0",
		"big:1:99999999999",
		"923:300000:65536",
	}, "\n")

	ranges, err := Parse(strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, []Range{
		{Owner: "alice", Start: 100000, Size: 65536},
		{Owner: "bob", Start: 200000, Size: 8},
		{Owner: "923", Start: 300000, Size: 65536},
	}, ranges)
}

func TestOwnedBy(t *testing.T) {
	ranges := []Range{
		{Owner: "alice", Start: 100000, Size: 65536},
		{Owner: "923", Start: 300000, Size: 65536},
		{Owner: "stapler-builder", Start: 400000, Size: 65536},
	}

	assert.Equal(t, []Range{ranges[1], ranges[2]}, OwnedBy(ranges, "stapler-builder", "923"))
	assert.Empty(t, OwnedBy(ranges, "bob", "1000"))
}

func TestNextFree(t *testing.T) {
	tests := []struct {
		name   string
		ranges []Range
		start  uint64
		ok     bool
	}{
		{name: "empty", start: 100000, ok: true},
		{
			name: "after overlapping ranges",
			ranges: []Range{
				{Owner: "alice", Start: 100000, Size: 65536},
				{Owner: "root", Start: 100000, Size: 1000000000},
			},
			start: 1000100000,
			ok:    true,
		},
		{
			name:   "below minimum",
			ranges: []Range{{Owner: "alice", Start: 1000, Size: 10}},
			start:  100000,
			ok:     true,
		},
		{
			name:   "exhausted",
			ranges: []Range{{Owner: "root", Start: 100000, Size: math.MaxUint32 - 100000 - 1000}},
			ok:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, ok := NextFree(tt.ranges, 100000, 65536)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.start, start)
		})
	}
}

func TestMappings(t *testing.T) {
	maps := Mappings(999, []Range{
		{Owner: "b", Start: 200000, Size: 1000},
		{Owner: "b", Start: math.MaxUint32 - 10, Size: 100},
		{Owner: "b", Start: 300000, Size: 65536},
	})

	assert.Equal(t, []IDMap{
		{ContainerID: 0, HostID: 999, Size: 1},
		{ContainerID: 1, HostID: 200000, Size: 1000},
		{ContainerID: 1001, HostID: 300000, Size: 65536},
	}, maps)
	assert.Equal(t, "0 999 1\n1 200000 1000\n1001 300000 65536\n", Format(maps))
}
