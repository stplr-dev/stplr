// ALR - Any Linux Repository
// Copyright (C) 2025 The ALR Authors
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

func TestE2EIssue75InstallWithDeps(t *testing.T) {
	runMatrixSuite(
		t,
		"issue-75-ref-specify",
		COMMON_SYSTEMS,
		func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)
			execShouldNoError(t, r, "sudo", "alr", "repo", "set-ref", "alr-repo", "bd26236cd7")
			execShouldNoError(t, r, "sh", "-c", "test $(alr list | wc -l) -eq 2 || exit 1")
		},
	)
}
