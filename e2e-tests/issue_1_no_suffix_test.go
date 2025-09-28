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
	"fmt"
	"testing"

	"go.alt-gnome.ru/capytest"
)

func TestE2EIssue1NoSuffix(t *testing.T) {
	t.Parallel()

	t.Run("with_no_suffix_flag", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		t.Parallel()
		defaultPrepare(t, r)

		execShouldNoError(t, r, "stplr", "build", "-p", "foo-pkg", "--no-suffix")

		r.Command("sh", "-c", "ls *.deb").
			ExpectSuccess().
			ExpectStdoutRegex(`(?m)^foo-pkg_[0-9.]+-[0-9]+_all\.deb$`).
			Run(t)
	}))

	t.Run("without_no_suffix_flag", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		t.Parallel()
		defaultPrepare(t, r)

		execShouldNoError(t, r, "stplr", "build", "-p", "foo-pkg")

		r.Command("sh", "-c", "ls *.deb").
			ExpectSuccess().
			ExpectStdoutRegex(fmt.Sprintf(`(?m)^foo-pkg(?:\+stplr-%s)_[0-9.]+-[0-9]+_all\.deb$`, REPO_NAME_FOR_E2E_TESTS)).
			Run(t)
	}))
}
