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

package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.stplr.dev/stplr/internal/constants"
)

func setCommonCmdEnv(cmd *exec.Cmd) {
	cmd.Env = []string{
		fmt.Sprintf("HOME=%s", constants.SystemCachePath),
		fmt.Sprintf("LOGNAME=%s", constants.BuilderUser),
		fmt.Sprintf("USER=%s", constants.BuilderUser),
		"PATH=/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/sbin:/usr/local/bin",
	}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "LANG=") ||
			strings.HasPrefix(env, "LANGUAGE=") ||
			strings.HasPrefix(env, "LC_") ||
			strings.HasPrefix(env, "STPLR_LOG_LEVEL=") {
			cmd.Env = append(cmd.Env, env)
		}
	}
}

func passCommonEnv(cmd *exec.Cmd) {
	cmd.Env = []string{
		"PATH=/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/sbin:/usr/local/bin",
	}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "HOME=") ||
			strings.HasPrefix(env, "LOGNAME=") ||
			strings.HasPrefix(env, "USER=") ||
			strings.HasPrefix(env, "LANG=") ||
			strings.HasPrefix(env, "LANGUAGE=") ||
			strings.HasPrefix(env, "LC_") ||
			strings.HasPrefix(env, "STPLR_LOG_LEVEL=") {
			cmd.Env = append(cmd.Env, env)
		}
	}
}
