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

package finddeps

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBracketValue(t *testing.T) {
	assert.Equal(t, "test", bracketValue("[test]"))
	assert.Equal(t, "", bracketValue("[test"))
	assert.Equal(t, "", bracketValue("test]"))
	assert.Equal(t, "", bracketValue("test"))
	assert.Equal(t, "a", bracketValue("[a]test[b]"))
}

func TestDiffSets(t *testing.T) {
	assert.Equal(t, map[string]struct{}{"a": {}}, diffSets(
		map[string]struct{}{"a": {}},
		map[string]struct{}{"b": {}},
	))

	assert.Equal(t, map[string]struct{}{"a": {}}, diffSets(
		map[string]struct{}{"a": {}},
		map[string]struct{}{},
	))

	assert.Equal(t, map[string]struct{}{}, diffSets(
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

func TestRelWithSlash(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		target   string
		expected string
	}{
		{
			name:     "1",
			base:     "/home/user",
			target:   "/home/user/docs/file.txt",
			expected: "/docs/file.txt",
		},
		{
			name:     "2",
			base:     "/home/user",
			target:   "/home/user/file.txt",
			expected: "/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := relWithSlash(tt.base, tt.target)
			require.NoError(t, err)
			assert.Equal(t, filepath.ToSlash(tt.expected), filepath.ToSlash(result))
		})
	}
}

func TestMatchesAnyPattern(t *testing.T) {
	tests := []struct {
		path     string
		patterns []string
		expected bool
	}{
		{
			path:     "/foo/bar.txt",
			patterns: []string{"**/*.txt"},
			expected: true,
		},
		{
			path:     "/foo/bar/baz.go",
			patterns: []string{"**/*.txt"},
			expected: false,
		},
		{
			path:     "/foo/bar/baz.go",
			patterns: []string{"**/*.go", "/foo/**"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			match, err := matchesAnyPattern(tt.path, tt.patterns)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, match)
		})
	}
}

func TestMakeRegexList(t *testing.T) {
	input := []string{`^foo`, `bar$`, `invalid[`}
	regexList := makeRegexList(input)

	require.Len(t, regexList, 2)
	assert.True(t, regexList[0].MatchString("foobar"))
	assert.True(t, regexList[1].MatchString("mybar"))
}

func TestMatchesRegexList(t *testing.T) {
	regexList := makeRegexList([]string{`^foo`, `bar$`})

	assert.True(t, matchesRegexList("foobar", regexList))
	assert.True(t, matchesRegexList("mybar", regexList))
	assert.False(t, matchesRegexList("baz", regexList))

	assert.False(t, matchesRegexList("anything", nil))
	assert.False(t, matchesRegexList("anything", []*regexp.Regexp{}))
}
