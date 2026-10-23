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

package setup

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nopOutput struct{}

func (nopOutput) Info(string, ...any)  {}
func (nopOutput) Warn(string, ...any)  {}
func (nopOutput) Error(string, ...any) {}

func currentUser(t *testing.T) *user.User {
	t.Helper()

	usr, err := user.Current()
	require.NoError(t, err)
	return usr
}

func newTestUseCase(t *testing.T, content string) (*useCase, *[]string) {
	t.Helper()

	file := filepath.Join(t.TempDir(), "subuid")
	require.NoError(t, os.WriteFile(file, []byte(content), 0o644))

	var calls []string
	u := &useCase{
		out:      nopOutput{},
		username: currentUser(t).Username,
		kinds:    []subIDKind{{file: file, flag: "--add-subuids"}},
		addRange: func(_ context.Context, flag, username string, start, end uint64) error {
			calls = append(calls, fmt.Sprintf("%s %d-%d %s", flag, start, end, username))
			return nil
		},
		lookPath: func(string) (string, error) { return "", nil },
	}
	return u, &calls
}

func TestRunAddsRangeAfterExisting(t *testing.T) {
	u, calls := newTestUseCase(t, "alice:100000:65536\nbob:100000:1000000000")

	require.NoError(t, u.Run(context.Background()))
	assert.Equal(t, []string{"--add-subuids 1000100000-1000165535 " + u.username}, *calls)
}

func TestRunSkipsExistingByName(t *testing.T) {
	u, calls := newTestUseCase(t, currentUser(t).Username+":300000:65536\n")

	require.NoError(t, u.Run(context.Background()))
	assert.Empty(t, *calls)
}

func TestRunSkipsExistingByUID(t *testing.T) {
	u, calls := newTestUseCase(t, currentUser(t).Uid+":300000:65536\n")

	require.NoError(t, u.Run(context.Background()))
	assert.Empty(t, *calls)
}

func TestRunFailsWhenExhausted(t *testing.T) {
	u, calls := newTestUseCase(t, "bob:100000:4294860000\n")

	require.Error(t, u.Run(context.Background()))
	assert.Empty(t, *calls)
}

func TestRunFailsWhenFileMissing(t *testing.T) {
	u, _ := newTestUseCase(t, "")
	require.NoError(t, os.Remove(u.kinds[0].file))

	require.Error(t, u.Run(context.Background()))
}
