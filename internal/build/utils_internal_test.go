// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
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

package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicatesSources(t *testing.T) {
	type testCase struct {
		Name         string
		Sources      []string
		Checksums    []string
		NewSources   []string
		NewChecksums []string
	}

	for _, tc := range []testCase{{
		Name:         "prefer non-skip values",
		Sources:      []string{"a", "b", "c", "a"},
		Checksums:    []string{"skip", "skip", "skip", "1"},
		NewSources:   []string{"a", "b", "c"},
		NewChecksums: []string{"1", "skip", "skip"},
	}} {
		t.Run(tc.Name, func(t *testing.T) {
			s, c := removeDuplicatesSources(tc.Sources, tc.Checksums)
			assert.Equal(t, s, tc.NewSources)
			assert.Equal(t, c, tc.NewChecksums)
		})
	}
}
