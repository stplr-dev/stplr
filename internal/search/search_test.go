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

package search_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/internal/search"
)

func TestSearhOptionsBuilder(t *testing.T) {
	type testCase struct {
		name          string
		prepare       func() *search.SearchOptions
		expectedWhere string
		expectedArgs  []any
	}

	for _, tc := range []testCase{
		{
			name: "Empty fields",
			prepare: func() *search.SearchOptions {
				return search.NewSearchOptions().
					Build()
			},
			expectedWhere: "",
			expectedArgs:  []any{},
		},
		{
			name: "All fields",
			prepare: func() *search.SearchOptions {
				return search.NewSearchOptions().
					WithName("foo").
					WithDescription("bar").
					WithRepository("buz").
					WithProvides("test").
					Build()
			},
			expectedWhere: "name LIKE ? AND description LIKE ? AND repository = ? AND json_array_contains(provides, ?)",
			expectedArgs:  []any{"%foo%", "%bar%", "buz", "test"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			whereClause, args := tc.prepare().WhereClause()
			assert.Equal(t, tc.expectedWhere, whereClause)
			assert.ElementsMatch(t, tc.expectedArgs, args)
		})
	}
}
