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

func TestE2EIssue298DisableRepo(t *testing.T) {
	t.Parallel()

	t.Run("successful disable and enable repo", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		execShouldNoError(t, r, "sh", "-c", "test $(stplr list | wc -l) -gt 0")
		execShouldNoError(t, r, "stplr", "repo", "set-disabled", REPO_NAME_FOR_E2E_TESTS, "true")
		execShouldNoError(t, r, "sh", "-c", "test $(stplr list | wc -l) -eq 0 || exit 1")
		execShouldNoError(t, r, "stplr", "repo", "set-disabled", REPO_NAME_FOR_E2E_TESTS, "false")
		execShouldNoError(t, r, "sh", "-c", "test $(stplr list | wc -l) -gt 0")
	}))
}
