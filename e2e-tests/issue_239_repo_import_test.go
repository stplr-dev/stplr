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
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.alt-gnome.ru/capytest"
)

func TestE2EIssue239RepoImport(t *testing.T) {
	t.Parallel()

	setupRepoBackup := func(t *testing.T, r capytest.Runner) {
		execShouldNoError(t, r, "sudo", "cp", "-r", filepath.Join("/var/cache/stplr/repo", REPO_NAME_FOR_E2E_TESTS), "/tmp/stplr-repo")
	}

	captureRepoList := func(t *testing.T, r capytest.Runner) string {
		var stdout bytes.Buffer
		r.Command("stplr", "repo", "list").
			WithCaptureStdout(&stdout).
			ExpectSuccess().
			Run(t)
		return stdout.String()
	}

	removeRepo := func(t *testing.T, r capytest.Runner) {
		execShouldNoError(t, r, "sudo", "stplr", "repo", "rm", REPO_NAME_FOR_E2E_TESTS)
	}

	t.Run("import from file and verify repository is pulled", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		stdoutBefore := captureRepoList(t, r)
		setupRepoBackup(t, r)
		removeRepo(t, r)

		r.Command("sudo", "stplr", "repo", "import", REPO_NAME_FOR_E2E_TESTS, "/tmp/stplr-repo/stapler-repo.toml").
			ExpectSuccess().
			ExpectStderrContains("Repository pulled successfully").
			Run(t)

		stdoutAfter := captureRepoList(t, r)
		assert.Equal(t, stdoutBefore, stdoutAfter)

		r.Command("sudo", "stplr", "repo", "import", REPO_NAME_FOR_E2E_TESTS, "/tmp/stplr-repo/stapler-repo.toml").
			ExpectFailure().
			Run(t)
	}))

	t.Run("import with --no-pull flag skips pulling", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		stdoutBefore := captureRepoList(t, r)
		setupRepoBackup(t, r)
		removeRepo(t, r)

		r.Command("sudo", "stplr", "repo", "import", "--no-pull", REPO_NAME_FOR_E2E_TESTS, "/tmp/stplr-repo/stapler-repo.toml").
			ExpectSuccess().
			ExpectStderrNotContains("Repository pulled successfully").
			Run(t)

		stdoutAfter := captureRepoList(t, r)
		assert.Equal(t, stdoutBefore, stdoutAfter)

		r.Command("sudo", "stplr", "repo", "import", "--no-pull", REPO_NAME_FOR_E2E_TESTS, "/tmp/stplr-repo/stapler-repo.toml").
			ExpectFailure().
			Run(t)
	}))

	t.Run("import with --ignore-existing flag allows reimport", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)
		setupRepoBackup(t, r)

		r.Command("sudo", "stplr", "repo", "import", "--ignore-existing", REPO_NAME_FOR_E2E_TESTS, "/tmp/stplr-repo/stapler-repo.toml").
			ExpectSuccess().
			Run(t)

		r.Command("sudo", "stplr", "repo", "import", "--ignore-existing", REPO_NAME_FOR_E2E_TESTS, "/tmp/stplr-repo/stapler-repo.toml").
			ExpectSuccess().
			Run(t)
	}))

	t.Run("import from stdin using pipe", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		stdoutBefore := captureRepoList(t, r)
		setupRepoBackup(t, r)
		removeRepo(t, r)

		r.Command("sudo", "sh", "-c", fmt.Sprintf("cat /tmp/stplr-repo/stapler-repo.toml | stplr repo import %s -", REPO_NAME_FOR_E2E_TESTS)).
			ExpectSuccess().
			ExpectStderrContains("Repository pulled successfully").
			Run(t)

		stdoutAfter := captureRepoList(t, r)
		assert.Equal(t, stdoutBefore, stdoutAfter)
	}))
}
