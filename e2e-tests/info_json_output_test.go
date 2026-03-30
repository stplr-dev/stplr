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

//go:build e2e

package e2etests_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"go.alt-gnome.ru/capytest"
)

func TestE2EInfoJsonOutput(t *testing.T) {
	t.Parallel()

	t.Run("info --json output works", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		var stdout bytes.Buffer
		r.Command("stplr", "info", "bar-pkg", "--json").
			ExpectStdoutMatchesSnapshot().
			ExpectSuccess().
			ExpectStderrEmpty().
			WithCaptureStdout(&stdout).
			Run(t)

		require.True(t, json.Valid(stdout.Bytes()), "output is not valid JSON: %s", stdout.String())
		var result []json.RawMessage
		err := json.Unmarshal(stdout.Bytes(), &result)
		require.NoError(t, err, "output is not a valid JSON array: %s", stdout.String())
	}))
}
