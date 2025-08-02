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
	"path/filepath"
	"regexp"
)

func ToSet[T comparable](xs []T) map[T]struct{} {
	set := make(map[T]struct{}, len(xs))
	for _, x := range xs {
		set[x] = struct{}{}
	}
	return set
}

func AnyMatch[T any](items []T, predicate func(T) bool) bool {
	for _, item := range items {
		if predicate(item) {
			return true
		}
	}
	return false
}

func diffSets(a, b map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(a))
	for k := range a {
		if _, ok := b[k]; !ok {
			out[k] = struct{}{}
		}
	}
	return out
}

func looksLikeLib(path string) bool {
	ok, _ := filepath.Match("lib*.so*", filepath.Base(path))
	return ok
}

func makeRegexList(filter []string) []*regexp.Regexp {
	var regexFilters []*regexp.Regexp
	for _, f := range filter {
		r, err := regexp.Compile(f)
		if err == nil {
			regexFilters = append(regexFilters, r)
		}
	}
	return regexFilters
}

func matchesRegexList(dep string, regexFilters []*regexp.Regexp) bool {
	return AnyMatch(regexFilters, func(r *regexp.Regexp) bool {
		return r.MatchString(dep)
	})
}
