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

package runner

import (
	"context"
	"strings"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func EnableStrictShellMode(ctx context.Context, r *interp.Runner) error {
	// It should be done via interp.RunnerOption,
	// but due to the issue below, it cannot be done.
	// - https://github.com/mvdan/sh/issues/962
	script, err := syntax.NewParser().Parse(strings.NewReader("set -euo pipefail"), "")
	if err != nil {
		return err
	}
	if err := r.Run(ctx, script); err != nil {
		return err
	}
	return nil
}
