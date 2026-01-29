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
	"fmt"
	"testing"

	"go.alt-gnome.ru/capytest"
)

func TestE2EFakerootDoesNotInvokeRootCmd(t *testing.T) {
	t.Parallel()

	t.Run("in fakeroot rootCmd not invoked", matrixSuite(COMMON_SYSTEMS, func(t *testing.T, r capytest.Runner) {
		defaultPrepare(t, r)

		rootCmd := "/tmp/rootcmd.sh"
		marker := "/tmp/rootcmd_called"

		execShouldNoError(t, r, "sh", "-c", fmt.Sprintf(`
set -e
cat > %s <<'EOF'
#!/bin/sh
echo called >> %s
exit 42
EOF
chmod +x %s
`, rootCmd, marker, rootCmd))

		execShouldNoError(t, r, "sudo", "stplr", "config", "set", "rootCmd", rootCmd)

		// fakeroot is detected via FAKEROOTKEY.
		// Setting it emulates fakeroot environment where real root checks
		// must be skipped.
		// migrate is expected to fail under fakeroot due to lack of real privileges.
		execShouldError(t, r, "sh", "-c", "FAKEROOTKEY=1 stplr migrate")

		// The important part is that rootCmd must NOT be invoked.
		execShouldNoError(t, r, "sh", "-c", fmt.Sprintf("test ! -f %s", marker))
	}))
}
