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

func TestParseMap(t *testing.T) {
	input := strings.Join([]string{
		"         0          0 4294967295",
		"",
		"broken line",
		"x y z",
		"1 2 0",
	}, "\n")

	ranges, err := ParseMap(strings.NewReader(input))
	require.NoError(t, err)
	assert.Equal(t, []Range{{Start: 0, Size: 4294967295}}, ranges)

	ranges, err = ParseMap(strings.NewReader("0 100000 65536\n1 200000 100\n"))
	require.NoError(t, err)
	assert.Equal(t, []Range{
		{Start: 0, Size: 65536},
		{Start: 1, Size: 100},
	}, ranges)
}

func TestClipTo(t *testing.T) {
	valid := []Range{{Start: 0, Size: 65536}}

	assert.Equal(t, []Range{{Owner: "b", Start: 0, Size: 1000}}, ClipTo([]Range{
		{Owner: "b", Start: 0, Size: 1000},
	}, valid))
	assert.Equal(t, []Range{{Owner: "b", Start: 65000, Size: 536}}, ClipTo([]Range{
		{Owner: "b", Start: 65000, Size: 1000},
	}, valid))
	assert.Empty(t, ClipTo([]Range{
		{Owner: "b", Start: 100000, Size: 65536},
	}, valid))
	assert.Equal(t, []Range{
		{Owner: "b", Start: 15, Size: 5},
		{Owner: "b", Start: 25, Size: 10},
	}, ClipTo([]Range{
		{Owner: "b", Start: 10, Size: 10},
		{Owner: "b", Start: 25, Size: 10},
		{Owner: "b", Start: 40, Size: 10},
	}, []Range{{Start: 15, Size: 25}}))
}

func TestExcludeID(t *testing.T) {
	ranges := []Range{{Owner: "b", Start: 100, Size: 100}}

	assert.Equal(t, ranges, ExcludeID(ranges, 50))
	assert.Equal(t, ranges, ExcludeID(ranges, 200))
	assert.Equal(t, []Range{{Owner: "b", Start: 101, Size: 99}}, ExcludeID(ranges, 100))
	assert.Equal(t, []Range{{Owner: "b", Start: 100, Size: 99}}, ExcludeID(ranges, 199))
	assert.Equal(t, []Range{
		{Owner: "b", Start: 100, Size: 50},
		{Owner: "b", Start: 151, Size: 49},
	}, ExcludeID(ranges, 150))
	assert.Empty(t, ExcludeID([]Range{{Owner: "b", Start: 150, Size: 1}}, 150))
}

func TestCoversAll(t *testing.T) {
	assert.True(t, CoversAll([]Range{{Start: 0, Size: math.MaxUint32}}))
	assert.True(t, CoversAll([]Range{{Start: 0, Size: 100}, {Start: 100, Size: math.MaxUint32 - 100}}))
	assert.False(t, CoversAll([]Range{{Start: 0, Size: math.MaxUint32 - 1}}))
	assert.False(t, CoversAll([]Range{{Start: 0, Size: 65536}}))
	assert.False(t, CoversAll(nil))
}

func TestSubordinate(t *testing.T) {
	containerIDs := []Range{{Start: 0, Size: 65536}}
	allIDs := []Range{{Start: 0, Size: math.MaxUint32}}
	configured := []Range{{Owner: "b", Start: 100000, Size: 65536}}

	tests := []struct {
		name       string
		id         uint32
		configured []Range
		valid      []Range
		synthesize bool
		want       []Range
		wantSynth  bool
	}{
		{
			name:       "configured ids usable as is",
			id:         999,
			configured: []Range{{Owner: "b", Start: 200000, Size: 1000}},
			valid:      allIDs,
			synthesize: true,
			want:       []Range{{Owner: "b", Start: 200000, Size: 1000}},
		},
		{
			name:       "configured id inside the range is excluded",
			id:         200500,
			configured: []Range{{Owner: "b", Start: 200000, Size: 1000}},
			valid:      allIDs,
			synthesize: true,
			want: []Range{
				{Owner: "b", Start: 200000, Size: 500},
				{Owner: "b", Start: 200501, Size: 499},
			},
		},
		{
			name:       "configured ids outside the container are synthesized",
			id:         1000,
			configured: configured,
			valid:      containerIDs,
			synthesize: true,
			want: []Range{
				{Start: 0, Size: 1000},
				{Start: 1001, Size: 64535},
			},
			wantSynth: true,
		},
		{
			name:       "nothing is synthesized for newuidmap and newgidmap",
			id:         1000,
			configured: configured,
			valid:      containerIDs,
			synthesize: false,
		},
		{
			name:       "nothing is synthesized in an unrestricted namespace",
			id:         1000,
			configured: nil,
			valid:      allIDs,
			synthesize: true,
		},
		{
			name:       "nothing to map when the builder is the only valid id",
			id:         7,
			configured: configured,
			valid:      []Range{{Start: 7, Size: 1}},
			synthesize: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges, synthesized := Subordinate(tt.id, tt.configured, tt.valid, tt.synthesize)
			assert.Equal(t, tt.want, ranges)
			assert.Equal(t, tt.wantSynth, synthesized)
		})
	}
}
