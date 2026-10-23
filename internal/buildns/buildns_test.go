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

package buildns

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPid(t *testing.T) {
	pid, err := readPid(strings.NewReader("1234\n"))
	require.NoError(t, err)
	assert.Equal(t, 1234, pid)

	_, err = readPid(strings.NewReader("oops\n"))
	require.Error(t, err)

	_, err = readPid(strings.NewReader(""))
	require.Error(t, err)
}

func TestPassAndInheritedPath(t *testing.T) {
	cmd := exec.Command("true")
	Pass(cmd, &Namespace{Path: "/proc/42/ns/user"})
	assert.Contains(t, cmd.Env, envPath+"=/proc/42/ns/user")

	t.Setenv(envPath, "/proc/42/ns/user")
	path, ok := InheritedPath()
	assert.True(t, ok)
	assert.Equal(t, "/proc/42/ns/user", path)

	t.Setenv(envPath, "")
	_, ok = InheritedPath()
	assert.False(t, ok)
}
