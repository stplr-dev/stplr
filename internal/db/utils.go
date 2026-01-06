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

package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/exp/slices"
	"modernc.org/sqlite"

	"go.stplr.dev/stplr/pkg/staplerfile"
)

func init() {
	sqlite.MustRegisterScalarFunction("json_array_contains", 2, jsonArrayContains)
	sqlite.MustRegisterScalarFunction("regexp_match", 2, regexpMatchCached)
	sqlite.MustRegisterScalarFunction("overridable_resolve", 2, overridableResolve)
}

// jsonArrayContains is an SQLite function that checks if a JSON array
// in the database contains a given value
func jsonArrayContains(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
	value, ok := args[0].(string)
	if !ok {
		return nil, errors.New("both arguments to json_array_contains must be strings")
	}

	item, ok := args[1].(string)
	if !ok {
		return nil, errors.New("both arguments to json_array_contains must be strings")
	}

	var array []string
	err := json.Unmarshal([]byte(value), &array)
	if err != nil {
		return nil, err
	}

	return slices.Contains(array, item), nil
}

var reCache = struct {
	sync.RWMutex
	m map[string]*regexp.Regexp
}{
	m: make(map[string]*regexp.Regexp),
}

func regexpMatchCached(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
	if args[0] == nil || args[1] == nil {
		return false, nil
	}

	str, ok := args[0].(string)
	if !ok {
		return nil, errors.New("both arguments to regexp_match must be strings")
	}

	pattern, ok := args[1].(string)
	if !ok {
		return nil, errors.New("both arguments to regexp_match must be strings")
	}

	reCache.RLock()
	re, exists := reCache.m[pattern]
	reCache.RUnlock()

	if !exists {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}

		reCache.Lock()
		reCache.m[pattern] = re
		reCache.Unlock()
	}

	return re.MatchString(str), nil
}

func overridableResolve(ctx *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
	value, ok := args[0].(string)
	if !ok {
		return nil, errors.New("argument to overridable_resolve must be a string")
	}

	overridesStr, ok := args[1].(string)
	if !ok {
		return nil, errors.New("argument to overridable_resolve must be a string")
	}

	var j map[string]interface{}
	err := json.Unmarshal([]byte(value), &j)
	if err != nil {
		return nil, err
	}

	v := staplerfile.OverridableFromMap(j)
	v.Resolve(strings.Split(overridesStr, "|"))
	resolved := v.Resolved()

	b, err := json.Marshal(resolved)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}
