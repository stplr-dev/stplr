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
	"fmt"
	"testing"

	"go.alt-gnome.ru/capytest"
)

func TestE2EIssue134DisableFirejail(t *testing.T) {
	t.Parallel()

	t.Run("firejail is disabling correctly during install", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)
		execShouldNoError(t, r, "stplr", "config", "set", "firejailExclude", "*firejail*")
		execShouldNoError(t, r, "stplr", "config", "get", "firejailExclude")

		r.Command("stplr", "install", "-i=false", fmt.Sprintf("%s/firejailed-pkg", REPO_NAME_FOR_E2E_TESTS)).
			ExpectSuccess().
			ExpectStderrContains("Security isolation will not be applied. Ensure you understand the risks").
			Run(t)

		r.Command("sh", "-c", "dpkg -L firejailed-pkg+stplr-alr-repo").
			ExpectSuccess().
			ExpectStdoutMatchesSnapshot().
			Run(t)
	}))

	t.Run("firejail is disabling correctly during build", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)
		execShouldNoError(t, r, "stplr", "config", "set", "firejailExclude", "*firejail*")
		execShouldNoError(t, r, "stplr", "config", "get", "firejailExclude")

		r.Command("stplr", "build", "-i=false", "-p", fmt.Sprintf("%s/firejailed-pkg", REPO_NAME_FOR_E2E_TESTS)).
			ExpectSuccess().
			ExpectStderrContains("Security isolation will not be applied. Ensure you understand the risks").
			Run(t)
	}))

	t.Run("firejail warning not showing up during install", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)
		execShouldNoError(t, r, "stplr", "config", "set", "firejailExclude", "*")
		execShouldNoError(t, r, "stplr", "config", "get", "firejailExclude")

		execShouldNoError(t, r, "stplr", "config", "set", "hideFirejailExcludeWarning", "true")

		r.Command("stplr", "install", "-i=false", fmt.Sprintf("%s/firejailed-pkg", REPO_NAME_FOR_E2E_TESTS)).
			ExpectSuccess().
			ExpectStderrNotContains("Security isolation will not be applied. Ensure you understand the risks").
			Run(t)

		r.Command("sh", "-c", "dpkg -L firejailed-pkg+stplr-alr-repo").
			ExpectSuccess().
			ExpectStdoutMatchesSnapshot().
			Run(t)
	}))
}
