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
	"fmt"
	"testing"

	"go.alt-gnome.ru/capytest"
	"go.alt-gnome.ru/capytest/providers/podman"
)

var ALL_SYSTEMS []string = []string{
	"ubuntu-24.04",
	"alt-sisyphus",
	"fedora-41",
	// "archlinux",
	// "alpine",
	// "opensuse-leap",
	// "redos-8",
}

var ALT_SISYPHUS = []string{
	"alt-sisyphus",
}

var AUTOREQ_AUTOPROV_SYSTEMS []string = []string{
	// "alt-sisyphus",
	"fedora-41",
}

var RPM_SYSTEMS []string = []string{
	"fedora-41",
}

var COMMON_SYSTEMS []string = []string{
	"ubuntu-24.04",
}

func execShouldNoError(t *testing.T, r capytest.Runner, cmd string, args ...string) {
	t.Helper()
	r.Command(cmd, args...).ExpectSuccess().Run(t)
}

func execShouldError(t *testing.T, r capytest.Runner, cmd string, args ...string) {
	t.Helper()
	r.Command(cmd, args...).ExpectFailure().Run(t)
}

const REPO_NAME_FOR_E2E_TESTS = "alr-repo"
const REPO_URL_FOR_E2E_TESTS = "https://altlinux.space/stapler/repo-for-tests.git"
const REPO_URL_FOR_E2E_TESTS_SINGLE_PACKAGE = "https://altlinux.space/stapler/repo-for-tests-single-package.git"

func defaultPrepare(t *testing.T, r capytest.Runner) {
	execShouldNoError(t, r,
		"sudo",
		"stplr",
		"repo",
		"add",
		REPO_NAME_FOR_E2E_TESTS,
		REPO_URL_FOR_E2E_TESTS,
	)

	execShouldNoError(t, r,
		"sudo",
		"stplr",
		"ref",
	)
}

func runMatrixSuite(t *testing.T, name string, images []string, test func(t *testing.T, r capytest.Runner)) {
	t.Helper()
	for _, image := range images {
		ts := capytest.NewTestSuite(t, podman.Provider(
			podman.WithImage(fmt.Sprintf("ghcr.io/maks1ms/stplr-e2e-test-image-%s", image)),
			podman.WithVolumes("./stplr:/tmp/stplr"),
			podman.WithPrivileged(true),
		))
		ts.BeforeEach(func(t *testing.T, r capytest.Runner) {
			execShouldNoError(t, r, "/bin/stplr-test-setup", "stplr-install")
			execShouldNoError(t, r, "/bin/stplr-test-setup", "passwordless-sudo-setup")
		})

		testName := fmt.Sprintf("%s/%s", name, image)
		if name == "" {
			testName = image
		}
		ts.Run(testName, test)
	}
}

func matrixSuite(images []string, test func(t *testing.T, r capytest.Runner)) func(t *testing.T) {
	return func(t *testing.T) {
		runMatrixSuite(
			t,
			"",
			images,
			test,
		)
	}
}
