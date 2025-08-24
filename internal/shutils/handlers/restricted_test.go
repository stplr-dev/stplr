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

package handlers_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/shutils/handlers"
)

func TestRestrictSandbox(t *testing.T) {
	tests := []struct {
		name        string
		allowedList []string
		path        string
		expected    bool
	}{
		{
			name:        "Blacklisted system cache path",
			allowedList: []string{},
			path:        constants.SystemCachePath,
			expected:    false,
		},
		{
			name:        "Blacklisted socket dir path",
			allowedList: []string{},
			path:        constants.SocketDirPath,
			expected:    false,
		},
		{
			name:        "Allowed path in allowed list",
			allowedList: []string{"/home/user"},
			path:        "/home/user/docs",
			expected:    true,
		},
		{
			name:        "Path not in allowed list or blacklisted",
			allowedList: []string{"/home/user"},
			path:        "/etc/config",
			expected:    true,
		},
		{
			name:        "Allowed path is prefix of input path",
			allowedList: []string{"/home/user/docs"},
			path:        "/home/user",
			expected:    true,
		},
		{
			name:        "Path not allowed",
			allowedList: []string{filepath.Join(constants.SystemCachePath, "foo")},
			path:        filepath.Join(constants.SystemCachePath, "bar"),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the predicate function
			predicate := handlers.RestrictSandbox(tt.allowedList...)

			// Test the predicate
			result := predicate(tt.path)
			assert.Equal(t, tt.expected, result, "Path: %s", tt.path)
		})
	}
}
