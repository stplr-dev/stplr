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

// Fails if the builder has no range in the file, has more than one, or if
// its range starts before the end of any other range.
const checkBuilderSubIDsScript = `
for f in /etc/subuid /etc/subgid; do
	awk -F: '
		$1 == "stapler-builder" { n++; start = $2; next }
		NF == 3 && $2 + $3 > maxEnd { maxEnd = $2 + $3 }
		END { exit !(n == 1 && start >= maxEnd) }
	' "$f" || { echo "bad ranges in $f"; cat "$f"; exit 1; }
done
`

func TestE2ESetupCommand(t *testing.T) {
	t.Parallel()

	t.Run("setup is idempotent", matrixSuite(ALL_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		r.Command("bash", "-c", checkBuilderSubIDsScript).
			ExpectSuccess().
			Run(t)

		r.Command("sudo", "stplr", "setup").
			ExpectSuccess().
			ExpectStderrNotContains("Added range").
			Run(t)

		r.Command("bash", "-c", checkBuilderSubIDsScript).
			ExpectSuccess().
			Run(t)
	}))

	t.Run("setup allocates non-overlapping subids", matrixSuite(ALL_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		execShouldNoError(t, r, "sudo", "sed", "-i", "/^stapler-builder:/d", "/etc/subuid", "/etc/subgid")

		r.Command("sudo", "stplr", "setup").
			ExpectSuccess().
			ExpectStderrContains("Added range").
			Run(t)

		r.Command("bash", "-c", checkBuilderSubIDsScript).
			ExpectSuccess().
			Run(t)
	}))

	t.Run("setup fails without subuid file", matrixSuite(FEDORA_43, func(t *testing.T, r capytest.Runner) {
		execShouldNoError(t, r, "sudo", "rm", "/etc/subuid")

		r.Command("sudo", "stplr", "setup").
			ExpectFailure().
			ExpectStderrContains("/etc/subuid").
			Run(t)
	}))
}
