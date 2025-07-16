// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
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

func TestE2EIssue74Upgradable(t *testing.T) {
	runMatrixSuite(
		t,
		"issue-74-upgradable",
		COMMON_SYSTEMS,
		func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)
			execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "9889ee9375bdf9a6410c86dc00f225ca7eb78bac")
			execShouldNoError(t, r, "stplr", "ref")
			execShouldNoError(t, r, "sudo", "stplr", "in", "bar-pkg")
			execShouldNoError(t, r, "sh", "-c", "test $(stplr list -U | wc -l) -eq 0 || exit 1")
			execShouldNoError(t, r, "sudo", "stplr", "repo", "set-ref", REPO_NAME_FOR_E2E_TESTS, "2a0187ea1118a922124f7567aa4565034dc77517")
			execShouldNoError(t, r, "sudo", "stplr", "ref")
			execShouldNoError(t, r, "sh", "-c", "test $(stplr list -U | wc -l) -eq 1 || exit 1")
		},
	)
}
