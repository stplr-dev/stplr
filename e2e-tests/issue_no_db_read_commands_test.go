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

func TestE2EReadCommandsWithoutDB(t *testing.T) {
	t.Parallel()

	t.Run("stplr list works without db", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)

		r.Command("stplr", "list").
			ExpectSuccess().
			Run(t)

		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)
	}))

	t.Run("stplr search works without db", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)

		r.Command("stplr", "search", "foo-pkg").
			ExpectSuccess().
			Run(t)

		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)
	}))

	t.Run("stplr info works without db", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)

		r.Command("stplr", "info", "foo-pkg").
			ExpectFailure().
			ExpectStderrContains("Package not found: package not found").
			Run(t)

		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)
	}))

	t.Run("stplr list --upgradable works without db", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)

		r.Command("stplr", "list", "--upgradable").
			ExpectSuccess().
			Run(t)
	}))

	t.Run("stplr list does not create db", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		r.Command("stplr", "list").
			ExpectSuccess().
			Run(t)

		r.Command("bash", "-c", "test ! -f /var/cache/stplr/db").
			ExpectSuccess().
			Run(t)
	}))
}
