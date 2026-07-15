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
	"testing"

	"go.alt-gnome.ru/capytest"
)

func TestE2ESourceOverride(t *testing.T) {
	t.Parallel()

	t.Run("successful source override", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		r.Command("sh", "-c", `echo "OVERRIDDEN_CONTENT" > /tmp/override.txt`).
			ExpectSuccess().
			Run(t)

		r.Command("stplr", "build", "-p", "test-source-override", "--source", "0:/tmp/override.txt", "--ignore-overriden-source-checksum").
			ExpectSuccess().
			ExpectStderrContains("Using local override for source [0]").
			Run(t)

		r.Command("sh", "-c", "dpkg-deb -x $(ls *.deb) /tmp/extracted && cat /tmp/extracted/opt/source-override-content").
			ExpectSuccess().
			ExpectStdoutContains("OVERRIDDEN_CONTENT").
			Run(t)
	}))

	t.Run("successful source override during install", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		r.Command("sh", "-c", `echo "OVERRIDDEN_CONTENT" > /tmp/override.txt`).
			ExpectSuccess().
			Run(t)

		r.Command("stplr", "install", "-i=false", "test-source-override", "--source", "0:/tmp/override.txt", "--ignore-overriden-source-checksum").
			ExpectSuccess().
			ExpectStderrContains("Using local override for source [0]").
			Run(t)

		r.Command("cat", "/opt/source-override-content").
			ExpectSuccess().
			ExpectStdoutContains("OVERRIDDEN_CONTENT").
			Run(t)
	}))

	t.Run("out of range source override index", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		r.Command("sh", "-c", `echo "OVERRIDDEN_CONTENT" > /tmp/override.txt`).
			ExpectSuccess().
			Run(t)

		r.Command("stplr", "build", "-p", "test-source-override", "--source", "99:/tmp/override.txt").
			ExpectFailure().
			ExpectStderrContains("source override index 99 is out of range").
			Run(t)
	}))
}
