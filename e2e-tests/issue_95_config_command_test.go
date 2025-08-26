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
)

func TestE2EIssue95ConfigCommand(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		property     string
		defaultValue string
		newValue     string
	}

	cases := []testCase{
		{
			name:         "autoPull",
			property:     "autoPull",
			defaultValue: "true",
			newValue:     "false",
		},

		{
			name:         "useRootCmd",
			property:     "useRootCmd",
			defaultValue: "true",
			newValue:     "false",
		},
		{
			name:         "forbidSkipInChecksums",
			property:     "forbidSkipInChecksums",
			defaultValue: "false",
			newValue:     "true",
		},
		{
			name:         "forbidBuildCommand",
			property:     "forbidBuildCommand",
			defaultValue: "false",
			newValue:     "true",
		},
		{
			name:         "rootCmd",
			property:     "rootCmd",
			defaultValue: "sudo",
			newValue:     "pkexec",
		},
		{
			name:         "pagerStyle",
			property:     "pagerStyle",
			defaultValue: "native",
			newValue:     "test",
		},
		{
			name:         "logLevel",
			property:     "logLevel",
			defaultValue: "info",
			newValue:     "ERROR",
		},
	}

	runMatrixSuite(
		t,
		"issue-95-config-command",
		COMMON_SYSTEMS,
		func(t *testing.T, r capytest.Runner) {
			defaultPrepare(t, r)

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					r.Command("stplr", "config", "get", tc.property).
						ExpectSuccess().
						ExpectStdoutRegex(fmt.Sprintf("^%s\n$", tc.defaultValue)).
						ExpectStderrEmpty().
						Run(t)

					execShouldError(t, r, "sudo", "stplr", "config", "set", tc.property)
					execShouldNoError(t, r, "sudo", "stplr", "config", "set", tc.property, tc.newValue)

					r.Command("stplr", "config", "show").
						ExpectSuccess().
						ExpectStdoutContains(fmt.Sprintf("%s: %s", tc.property, tc.newValue)).
						ExpectStderrEmpty().
						Run(t)

					r.Command("stplr", "config", "get").
						ExpectSuccess().
						ExpectStdoutContains(fmt.Sprintf("%s: %s", tc.property, tc.newValue)).
						ExpectStderrEmpty().
						Run(t)

					r.Command("stplr", "config", "get", tc.property).
						ExpectSuccess().
						ExpectStdoutRegex(fmt.Sprintf("^%s\n$", tc.newValue)).
						ExpectStderrEmpty().
						Run(t)

					r.Command("stplr", "config", "get", tc.property).
						ExpectSuccess().
						ExpectStdoutRegex(fmt.Sprintf("^%s\n$", tc.newValue)).
						ExpectStderrEmpty().
						Run(t)

					execShouldNoError(t, r, "sudo", "stplr", "config", "set", tc.property, tc.defaultValue)
				})
			}
		},
	)
}
