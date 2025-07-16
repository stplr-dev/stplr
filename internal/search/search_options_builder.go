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

package search

import (
	"fmt"
	"strings"
)

type SearchOptions struct {
	conditions []string
	args       []any
}

func (o *SearchOptions) WhereClause() (string, []any) {
	if len(o.conditions) == 0 {
		return "", nil
	}
	return strings.Join(o.conditions, " AND "), o.args
}

type SearchOptionsBuilder struct {
	options SearchOptions
}

func NewSearchOptions() *SearchOptionsBuilder {
	return &SearchOptionsBuilder{}
}

func (b *SearchOptionsBuilder) withGeneralLike(key, value string) *SearchOptionsBuilder {
	if value != "" {
		b.options.conditions = append(b.options.conditions, fmt.Sprintf("%s LIKE ?", key))
		b.options.args = append(b.options.args, "%"+value+"%")
	}
	return b
}

func (b *SearchOptionsBuilder) withGeneralEqual(key string, value any) *SearchOptionsBuilder {
	if value != "" {
		b.options.conditions = append(b.options.conditions, fmt.Sprintf("%s = ?", key))
		b.options.args = append(b.options.args, value)
	}
	return b
}

func (b *SearchOptionsBuilder) withGeneralJsonArrayContains(key string, value any) *SearchOptionsBuilder {
	if value != "" {
		b.options.conditions = append(b.options.conditions, fmt.Sprintf("json_array_contains(%s, ?)", key))
		b.options.args = append(b.options.args, value)
	}
	return b
}

func (b *SearchOptionsBuilder) WithName(name string) *SearchOptionsBuilder {
	return b.withGeneralLike("name", name)
}

func (b *SearchOptionsBuilder) WithDescription(description string) *SearchOptionsBuilder {
	return b.withGeneralLike("description", description)
}

func (b *SearchOptionsBuilder) WithRepository(repository string) *SearchOptionsBuilder {
	return b.withGeneralEqual("repository", repository)
}

func (b *SearchOptionsBuilder) WithProvides(provides string) *SearchOptionsBuilder {
	return b.withGeneralJsonArrayContains("provides", provides)
}

func (b *SearchOptionsBuilder) Build() *SearchOptions {
	return &b.options
}
