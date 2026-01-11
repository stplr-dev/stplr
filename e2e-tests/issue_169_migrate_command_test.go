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

	"go.stplr.dev/stplr/internal/db"
)

func TestE2EIssue169MigrateCommand(t *testing.T) {
	t.Parallel()

	t.Run("call migrate multiple works", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		r.Command("sudo", "stplr", "migrate").
			ExpectSuccess().
			Run(t)

		r.Command("sudo", "stplr", "migrate").
			ExpectSuccess().
			Run(t)
	}))

	t.Run("migrate updates db", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		// Install sqlite3 for validating db changes
		r.Command("sudo", "apt", "update").
			ExpectSuccess().
			Run(t)
		r.Command("sudo", "apt", "install", "-y", "sqlite3").
			ExpectSuccess().
			Run(t)

		// Dirty but simple disable network
		r.Command("sudo", "sh", "-c", "echo 'nameserver 127.0.0.1' > /etc/resolv.conf").
			ExpectSuccess().
			Run(t)
		r.Command("getent", "hosts", "google.com").
			ExpectExitCode(2).
			Run(t)

		dbPath := "/var/cache/stplr/db"

		r.Command("sudo", "sqlite3", dbPath,
			"UPDATE version SET version = 0;").
			ExpectSuccess().
			Run(t)
		r.Command("sqlite3", dbPath, "SELECT version FROM version;").
			ExpectSuccess().
			ExpectStdoutContains("0").
			Run(t)

		r.Command("sudo", "stplr", "migrate").
			ExpectSuccess().
			Run(t)

		r.Command("sqlite3", dbPath, "SELECT version FROM version;").
			ExpectSuccess().
			ExpectStdoutContains(fmt.Sprintf("%d", db.CurrentVersion)).
			Run(t)
	}))
}
