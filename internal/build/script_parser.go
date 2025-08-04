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

package build

import (
	"context"

	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type LocalScriptParser struct{}

func (p *LocalScriptParser) ParsePackages(
	ctx context.Context,
	f *staplerfile.ScriptFile,
	packages []string, info distro.OSRelease,
) (string, []*staplerfile.Package, error) {
	return f.ParseBuildVars(ctx, &info, packages)
}
