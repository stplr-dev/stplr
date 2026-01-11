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

package commonbuild

import (
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/types"
)

type BuiltDep struct {
	Name string
	Path string
}

type BuildOptsProvider interface {
	BuildOpts() *types.BuildOpts
}

type OsInfoProvider interface {
	OSRelease() *distro.OSRelease
}

type PkgFormatProvider interface {
	PkgFormat() string
}

type RepositoryProvider interface {
	Repository() string
}

type Config interface {
	GetPaths() *config.Paths
	PagerStyle() string
	FirejailExclude() []string
	HideFirejailExcludeWarning() bool
}

type FunctionsOutput struct {
	Contents *[]string
}

type OSReleaser interface {
	OSRelease() *distro.OSRelease
}

type PkgFormatter interface {
	PkgFormat() string
}

type RepositoryGetter interface {
	Repository() string
}
