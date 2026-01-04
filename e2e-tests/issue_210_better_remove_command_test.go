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

func TestE2EIssue210BetterRemoveCommand(t *testing.T) {
	t.Parallel()

	const pkg = "bar-pkg"

	t.Run(fmt.Sprintf("stplr remove %s/%s works", REPO_NAME_FOR_E2E_TESTS, pkg), matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		// Prepare
		execShouldNoError(t, r, "sudo", "stplr", "in", pkg)
		execShouldNoError(t, r, "sudo", "stplr", "remove", fmt.Sprintf("%s/%s", REPO_NAME_FOR_E2E_TESTS, pkg))
	}))
}
