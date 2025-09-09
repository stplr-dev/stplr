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

func TestE2EAlrAddRepo(t *testing.T) {
	t.Parallel()

	runMatrixSuite(
		t,
		"add-repo-remove-repo",
		COMMON_SYSTEMS,
		func(t *testing.T, r capytest.Runner) {
			execShouldNoError(t, r, "sudo", "stplr", "repo", "add", "alr-repo", REPO_URL_FOR_E2E_TESTS)
			execShouldNoError(t, r, "bash", "-c", "cat /etc/stplr/stplr.toml")
			execShouldNoError(t, r, "sudo", "stplr", "repo", "rm", "alr-repo")

			r.Command("bash", "-c", "cat /etc/stplr/stplr.toml").
				ExpectStdoutContains("repo = []").
				Run(t)
		},
	)
}
