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

func TestE2ESandboxIDMappings(t *testing.T) {
	t.Parallel()

	runMatrixSuite(
		t,
		"sandbox maps subordinate ids",
		ALL_SYSTEMS,
		func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)

			r.Command("stplr", "-i=false", "build", "-p", "test-id-mappings").
				ExpectSuccess().
				ExpectStderrNotContains("can't change file ownership").
				Run(t)

			r.Command("sudo", "stplr", "-i=false", "build", "-p", "test-id-mappings").
				ExpectSuccess().
				ExpectStderrNotContains("can't change file ownership").
				Run(t)
		},
	)
}
