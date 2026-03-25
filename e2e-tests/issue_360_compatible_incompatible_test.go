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

func TestE2EIssue360CompatibleIncompatible(t *testing.T) {
	t.Parallel()

	t.Run("compatible", func(t *testing.T) {
		t.Run("positive", matrixSuite(ALT_SISYPHUS, func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)

			r.Command("stplr", "install", "test-360-compatible").
				ExpectSuccess().
				Run(t)
		}))

		t.Run("negative", matrixSuite(FEDORA_43, func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)

			r.Command("stplr", "install", "test-360-compatible").
				ExpectFailure().
				Run(t)
		}))
	})

	t.Run("incompatible", func(t *testing.T) {
		t.Run("negative", matrixSuite(ALT_SISYPHUS, func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)

			r.Command("stplr", "install", "test-360-incompatible").
				ExpectFailure().
				Run(t)
		}))

		t.Run("positive", matrixSuite(FEDORA_43, func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)

			r.Command("stplr", "install", "test-360-incompatible").
				ExpectSuccess().
				Run(t)
		}))
	})

}
