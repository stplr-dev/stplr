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

package cpu

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCompatibleWith(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		list     []string
		expected bool
	}{
		{"Exact match", "amd64", []string{"amd64", "armv7"}, true},
		{"Universal target", "all", []string{"amd64"}, true},
		{"Universal list", "armv7", []string{"all"}, true},
		{"ARM invalid version", "invalid", []string{"armv7"}, false},
		{"No match", "amd64", []string{"armv6", "armv7"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCompatibleWith(tt.target, tt.list)
			if got != tt.expected {
				t.Errorf("IsCompatibleWith(%q, %v) = %v; want %v", tt.target, tt.list, got, tt.expected)
			}
		})
	}
}

func TestIsCompatibleARM(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		arch     string
		expected bool
	}{
		{
			name:     "Target arm7 is compatible with arch arm6",
			target:   "arm7",
			arch:     "arm6",
			expected: true,
		},
		{
			name:     "Target arm7 is compatible with arch arm7",
			target:   "arm7",
			arch:     "arm7",
			expected: true,
		},
		{
			name:     "Target arm6 is not compatible with arch arm7",
			target:   "arm6",
			arch:     "arm7",
			expected: false,
		},
		{
			name:     "Non-ARM target",
			target:   "amd64",
			arch:     "arm6",
			expected: false,
		},
		{
			name:     "Non-ARM arch",
			target:   "arm6",
			arch:     "amd64",
			expected: false,
		},
		{
			name:     "Default ARM version with no number (target arm)",
			target:   "arm",
			arch:     "arm5",
			expected: true,
		},
		{
			name:     "Default ARM version with no number (arch arm)",
			target:   "arm7",
			arch:     "arm",
			expected: true,
		},
		{
			name:     "Both target and arch are 'arm' (default version 5)",
			target:   "arm",
			arch:     "arm",
			expected: true,
		},
		{
			name:     "Invalid ARM version in target",
			target:   "armx",
			arch:     "arm5",
			expected: false,
		},
		{
			name:     "Invalid ARM version in arch",
			target:   "arm7",
			arch:     "armx",
			expected: false,
		},
		{
			name:     "Unsupported prefix like armv7",
			target:   "armv7",
			arch:     "arm6",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isCompatibleARM(tt.target, tt.arch))
		})
	}
}

func TestCompatibleArches(t *testing.T) {
	tests := []struct {
		name     string
		arch     string
		expected []string
		err      error
	}{
		{
			name:     "ARM version 7",
			arch:     "arm7",
			expected: []string{"arm7", "arm6", "arm5"},
			err:      nil,
		},
		{
			name:     "ARM version 6",
			arch:     "arm6",
			expected: []string{"arm6", "arm5"},
			err:      nil,
		},
		{
			name:     "ARM version 5",
			arch:     "arm5",
			expected: []string{"arm5"},
			err:      nil,
		},
		{
			name:     "Non-ARM architecture",
			arch:     "x86",
			expected: []string{"x86"},
			err:      nil,
		},
		{
			name:     "Invalid ARM version",
			arch:     "arm_invalid",
			expected: nil,
			err:      &strconv.NumError{Func: "Atoi", Num: "_invalid", Err: strconv.ErrSyntax},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompatibleArches(tt.arch)
			assert.Equal(t, tt.expected, got)
			assert.Equal(t, tt.err, err)
		})
	}
}
