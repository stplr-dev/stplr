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

func TestE2EForbidSettings(t *testing.T) {
	t.Parallel()

	t.Run("forbidSkipInChecksums works", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		const testPackage = "test-forbid-checksums"

		disable := func(t *testing.T, r capytest.Runner) {
			execShouldNoError(t, r, "sudo", "stplr", "config", "set", "forbidSkipInChecksums", "false")
		}
		enable := func(t *testing.T, r capytest.Runner) {
			execShouldNoError(t, r, "sudo", "stplr", "config", "set", "forbidSkipInChecksums", "true")
		}

		defaultPrepare(t, r)
		t.Run("build", func(t *testing.T) {
			disable(t, r)
			r.Command("stplr", "build", "-p", testPackage).
				ExpectSuccess().
				Run(t)

			enable(t, r)
			r.Command("stplr", "build", "-p", testPackage).
				ExpectStderrContains("Your settings do not allow SKIP in checksums").
				ExpectFailure().
				Run(t)
		})

		t.Run("install", func(t *testing.T) {
			disable(t, r)
			r.Command("stplr", "-i=false", "install", testPackage).
				ExpectSuccess().
				Run(t)

			// Cleanup cache
			r.Command("stplr", "fix").
				ExpectSuccess().
				Run(t)

			enable(t, r)
			r.Command("stplr", "-i=false", "install", testPackage).
				ExpectStderrContains("Your settings do not allow SKIP in checksums").
				ExpectFailure().
				Run(t)
		})
	}))

	t.Run("forbidBuildCommand works", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		const testPackage = "test-forbid-checksums"

		defaultPrepare(t, r)

		execShouldNoError(t, r, "sudo", "stplr", "config", "set", "forbidBuildCommand", "false")
		r.Command("stplr", "build", "-p", testPackage).
			ExpectSuccess().
			Run(t)

		execShouldNoError(t, r, "sudo", "stplr", "config", "set", "forbidBuildCommand", "true")
		r.Command("stplr", "build", "-p", testPackage).
			ExpectStderrContains("Your settings do not allow build command").
			ExpectFailure().
			Run(t)
	}))
}
