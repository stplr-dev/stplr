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

func TestE2ENewRepoStructure(t *testing.T) {
	t.Parallel()

	t.Run("new repo structure override files", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		// No override file should exist before any set-* command.
		r.Command("bash", "-c", "test ! -f /etc/stplr/repo-overrides.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)

		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-disabled", REPO_NAME_FOR_E2E_TESTS, "true")

		// Override must be written to repo-overrides.d/, not to stplr.toml.
		r.Command("bash", "-c", "test -f /etc/stplr/repo-overrides.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)
		r.Command("bash", "-c", "! grep -q disabled /etc/stplr/stplr.toml").
			ExpectSuccess().
			Run(t)

		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-disabled", REPO_NAME_FOR_E2E_TESTS, "false")

		// Override file must still exist.
		r.Command("bash", "-c", "test -f /etc/stplr/repo-overrides.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)

		execShouldNoError(t, r, "sudo", "stplr", "ref")
	}))

	t.Run("new repo structure clear overrides", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		execShouldNoError(t, r, "sudo", "stplr", "repo", "set-disabled", REPO_NAME_FOR_E2E_TESTS, "true")
		execShouldNoError(t, r, "sh", "-c", "test $(stplr list | wc -l) -eq 0")

		r.Command("bash", "-c", "test -f /etc/stplr/repo-overrides.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)

		execShouldNoError(t, r, "sudo", "stplr", "repo", "clear-overrides", REPO_NAME_FOR_E2E_TESTS)

		// Override file must be gone after clear-overrides.
		r.Command("bash", "-c", "test ! -f /etc/stplr/repo-overrides.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)

		// Repo must be enabled again.
		execShouldNoError(t, r, "sh", "-c", "test $(stplr list | wc -l) -gt 0")
	}))

	t.Run("user repo updates on pull", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		r.Command("sudo", "bash", "-c", "sed -i 's/summary = .*/summary = \"TEST-SUMMARY\"/' /etc/stplr/repos.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)

		execShouldNoError(t, r, "sudo", "stplr", "ref")

		r.Command("bash", "-c", "grep -q 'Stapler repo for tests' /etc/stplr/repos.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)
	}))

	t.Run("system repo restrictions", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		// Simulate a system-provided repo.
		execShouldNoError(t, r, "sudo", "mkdir", "-p", "/usr/lib/stplr/repos.d")
		execShouldNoError(t, r, "sudo", "mv", "/etc/stplr/repos.d/alr-repo.toml", "/usr/lib/stplr/repos.d/alr-repo.toml")

		// System repos cannot be removed.
		execShouldError(t, r, "sudo", "stplr", "repo", "rm", REPO_NAME_FOR_E2E_TESTS)

		// Cannot add a repo with the same name or the same URL as a system repo.
		execShouldError(t, r, "sudo", "stplr", "repo", "add", REPO_NAME_FOR_E2E_TESTS, REPO_URL_FOR_E2E_TESTS)
		execShouldError(t, r, "sudo", "stplr", "repo", "add", "another-name", REPO_URL_FOR_E2E_TESTS)

		// System repo must not be overwritten by pull.
		r.Command("sudo", "bash", "-c", "sed -i 's/summary = .*/summary = \"TEST-SUMMARY\"/' /usr/lib/stplr/repos.d/alr-repo.toml").
			ExpectSuccess().
			Run(t)
		execShouldNoError(t, r, "sudo", "stplr", "ref")
		r.Command("bash", "-c", `grep -q 'TEST-SUMMARY' /usr/lib/stplr/repos.d/alr-repo.toml`).
			ExpectSuccess().
			Run(t)
	}))
}
