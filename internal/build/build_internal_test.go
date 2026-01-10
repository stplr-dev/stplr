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

package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/pkg/staplerfile"
)

func TestGroupPackages(t *testing.T) {
	tests := []struct {
		name     string
		input    []staplerfile.Package
		expected map[string]pkgItem
	}{
		{
			name:     "empty slice",
			input:    []staplerfile.Package{},
			expected: map[string]pkgItem{},
		},
		{
			name: "single package with BasePkgName",
			input: []staplerfile.Package{
				{Name: "package1", BasePkgName: "base1"},
			},
			expected: map[string]pkgItem{
				"base1": {
					pkg:      &staplerfile.Package{Name: "package1", BasePkgName: "base1"},
					packages: []string{"package1"},
				},
			},
		},
		{
			name: "single package without BasePkgName",
			input: []staplerfile.Package{
				{Name: "package1", BasePkgName: ""},
			},
			expected: map[string]pkgItem{
				"package1": {
					pkg:      &staplerfile.Package{Name: "package1", BasePkgName: ""},
					packages: []string{"package1"},
				},
			},
		},
		{
			name: "multiple packages with same BasePkgName",
			input: []staplerfile.Package{
				{Name: "package1", BasePkgName: "base1"},
				{Name: "package2", BasePkgName: "base1"},
				{Name: "package3", BasePkgName: "base1"},
			},
			expected: map[string]pkgItem{
				"base1": {
					pkg:      &staplerfile.Package{Name: "package1", BasePkgName: "base1"},
					packages: []string{"package1", "package2", "package3"},
				},
			},
		},
		{
			name: "multiple packages with different BasePkgName",
			input: []staplerfile.Package{
				{Name: "package1", BasePkgName: "base1"},
				{Name: "package2", BasePkgName: "base2"},
				{Name: "package3", BasePkgName: "base3"},
			},
			expected: map[string]pkgItem{
				"base1": {
					pkg:      &staplerfile.Package{Name: "package1", BasePkgName: "base1"},
					packages: []string{"package1"},
				},
				"base2": {
					pkg:      &staplerfile.Package{Name: "package2", BasePkgName: "base2"},
					packages: []string{"package2"},
				},
				"base3": {
					pkg:      &staplerfile.Package{Name: "package3", BasePkgName: "base3"},
					packages: []string{"package3"},
				},
			},
		},
		{
			name: "mixed packages with and without BasePkgName",
			input: []staplerfile.Package{
				{Name: "package1", BasePkgName: "base1"},
				{Name: "package2", BasePkgName: ""},
				{Name: "package3", BasePkgName: "base1"},
			},
			expected: map[string]pkgItem{
				"base1": {
					pkg:      &staplerfile.Package{Name: "package1", BasePkgName: "base1"},
					packages: []string{"package1", "package3"},
				},
				"package2": {
					pkg:      &staplerfile.Package{Name: "package2", BasePkgName: ""},
					packages: []string{"package2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := groupPackages(tt.input)

			require.Equal(t, len(tt.expected), len(result), "maps should have the same length")

			for key, expectedItem := range tt.expected {
				actualItem, exists := result[key]
				require.True(t, exists, "key %s should exist in result", key)

				assert.Equal(t, expectedItem.pkg.Name, actualItem.pkg.Name, "pkg.Name should match for key %s", key)
				assert.Equal(t, expectedItem.pkg.BasePkgName, actualItem.pkg.BasePkgName, "pkg.BasePkgName should match for key %s", key)

				assert.ElementsMatch(t, expectedItem.packages, actualItem.packages, "packages slice should match for key %s", key)
			}
		})
	}
}

func TestGroupPackagesPointerUniqueness(t *testing.T) {
	input := []staplerfile.Package{
		{Name: "package1", BasePkgName: "base1"},
		{Name: "package2", BasePkgName: "base2"},
	}

	result := groupPackages(input)

	pkg1 := result["base1"].pkg
	pkg2 := result["base2"].pkg

	assert.NotSame(t, pkg1, pkg2, "pointers should be different")
	assert.NotSame(t, pkg1, &input[0], "result pointer should not point to original slice element")
	assert.NotSame(t, pkg2, &input[1], "result pointer should not point to original slice element")
}

func TestGroupPackagesPreservesFirstPackage(t *testing.T) {
	input := []staplerfile.Package{
		{Name: "first", BasePkgName: "base1"},
		{Name: "second", BasePkgName: "base1"},
		{Name: "third", BasePkgName: "base1"},
	}

	result := groupPackages(input)

	require.Contains(t, result, "base1")
	assert.Equal(t, "first", result["base1"].pkg.Name, "should preserve first package")
	assert.Equal(t, []string{"first", "second", "third"}, result["base1"].packages)
}

func TestFirejailedPatternMatch(t *testing.T) {
	tests := []struct {
		name     string
		pkg      string
		pattern  string
		expected bool
	}{
		{
			name:     "all",
			pkg:      "stplr-repo/test",
			pattern:  "*",
			expected: true,
		},
		{
			name:     "match by part",
			pkg:      "stplr-repo/test",
			pattern:  "*test",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := firejailedPatternMatch(tt.pkg, tt.pattern)

			assert.Equal(t, tt.expected, result)
		})
	}
}
