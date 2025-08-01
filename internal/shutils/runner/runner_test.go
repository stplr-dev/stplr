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

package runner_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/shutils/runner"
)

func newTestRunner(t *testing.T) (*interp.Runner, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer

	r, err := interp.New(
		interp.StdIO(nil, &stdout, &stderr),
	)
	require.NoError(t, err)

	return r, &stdout
}

func TestEnableStrictShellModeSuccess(t *testing.T) {
	ctx := context.Background()
	r, _ := newTestRunner(t)

	err := runner.EnableStrictShellMode(ctx, r)
	require.NoError(t, err)
}

func TestEnableStrictShellModeErrInvalidCommand(t *testing.T) {
	ctx := context.Background()
	r, out := newTestRunner(t)

	require.NoError(t, runner.EnableStrictShellMode(ctx, r))

	parser := syntax.NewParser()
	node, err := parser.Parse(strings.NewReader(`echo $UNDEFINED_VAR; echo "should not print"`), "")
	require.NoError(t, err)

	err = r.Run(ctx, node)
	require.Error(t, err)
	require.NotContains(t, out.String(), "should not print")
}

func TestEnableStrictShellModeErrOnPipefail(t *testing.T) {
	ctx := context.Background()
	r, out := newTestRunner(t)

	require.NoError(t, runner.EnableStrictShellMode(ctx, r))

	parser := syntax.NewParser()
	node, err := parser.Parse(strings.NewReader(`false | true; echo "should not print"`), "")
	require.NoError(t, err)

	err = r.Run(ctx, node)
	require.Error(t, err)
	require.NotContains(t, out.String(), "should not print")
}
