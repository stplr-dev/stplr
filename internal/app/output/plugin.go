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

package output

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
)

type pluginOutput struct {
	logger hclog.Logger
}

func NewPluginOutput() *pluginOutput {
	return &pluginOutput{
		logger: hclog.New(&hclog.LoggerOptions{
			Output:      os.Stderr,
			Level:       hclog.Trace,
			JSONFormat:  true,
			DisableTime: true,
		}),
	}
}

func (out *pluginOutput) print(level hclog.Level, msg string, args ...any) {
	out.logger.Log(
		level,
		fmt.Sprintf(msg, args...),
		"@_type", "user",
	)
}

func (out *pluginOutput) Info(msg string, args ...any) {
	out.print(hclog.Info, msg, args...)
}

func (out *pluginOutput) Warn(msg string, args ...any) {
	out.print(hclog.Warn, msg, args...)
}

func (out *pluginOutput) Error(msg string, args ...any) {
	out.print(hclog.Error, msg, args...)
}
