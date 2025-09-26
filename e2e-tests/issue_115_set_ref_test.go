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

func TestE2EIssue115SetRef(t *testing.T) {
	t.Parallel()
	matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		execShouldNoError(t, r,
			"sudo",
			"stplr",
			"repo",
			"add",
			REPO_NAME_FOR_E2E_TESTS,
			REPO_URL_FOR_E2E_TESTS_SINGLE_PACKAGE,
		)
		execShouldNoError(t, r, "sudo", "stplr", "ref")
		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "930962da6ed1ab0a812cb42e7d05d0d56127ec68")
		execShouldNoError(t, r, "sudo", "stplr", "fix")
		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "refs/pull/1/head")
		execShouldNoError(t, r, "sudo", "stplr", "fix")
	})(t)
}
