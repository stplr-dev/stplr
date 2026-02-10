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

func TestE2EIssue266ImplementIgnorePkgUpdates(t *testing.T) {
	t.Parallel()

	t.Run("ignorePkgUpdates Works", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "9889ee9375bdf9a6410c86dc00f225ca7eb78bac")
		execShouldNoError(t, r, "sudo", "stplr", "ref")
		execShouldNoError(t, r, "sudo", "stplr", "in", "bar-pkg")
		execShouldNoError(t, r, "sh", "-c", "test $(stplr list -U | wc -l) -eq 0 || exit 1")
		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "2a0187ea1118a922124f7567aa4565034dc77517")
		execShouldNoError(t, r, "sh", "-c", "test $(stplr list -U | wc -l) -eq 1 || exit 1")
		execShouldNoError(t, r, "sudo", "stplr", "config", "set", "ignorePkgUpdates", "*")
		execShouldNoError(t, r, "sh", "-c", "test $(stplr list -U | wc -l) -eq 0 || exit 1")

		r.Command("sudo", "stplr", "-i=false", "upgrade").
			ExpectStderrContains("There is nothing to do.").
			ExpectSuccess().
			Run(t)
	}))
}
