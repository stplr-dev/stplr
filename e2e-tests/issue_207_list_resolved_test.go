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

//go:build e2e

package e2etests_test

import (
	"testing"

	"go.alt-gnome.ru/capytest"
)

func TestE2EIssue207ListResolved(t *testing.T) {
	t.Parallel()

	t.Run("list command (installed) show resolved fields", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		// Prepare
		execShouldNoError(t, r, "sudo", "stplr", "in", "bar-pkg")

		r.Command("stplr", "list", "-I", "-f", "{{ .Package.Name }} {{ .Package.Description.Resolved }}").
			ExpectSuccess().
			ExpectStdoutMatchesSnapshot().
			Run(t)
	}))

	t.Run("list command (updatable) show resolved fields", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		// Prepare
		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "9889ee9375bdf9a6410c86dc00f225ca7eb78bac")
		execShouldNoError(t, r, "sudo", "stplr", "in", "bar-pkg")
		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "2a0187ea1118a922124f7567aa4565034dc77517")
		execShouldNoError(t, r, "sudo", "stplr", "ref")

		// Actual test
		r.Command("stplr", "list", "-U", "-f", "{{ .Package.Name }} {{ .Package.Description.Resolved }}").
			ExpectSuccess().
			ExpectStdoutMatchesSnapshot().
			Run(t)
	}))
}
