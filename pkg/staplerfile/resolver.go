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

package staplerfile

import (
	"github.com/jeandeaual/go-locale"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/pkg/distro"
)

type Resolver struct {
	info *distro.OSRelease

	names []string
}

func NewResolver(info *distro.OSRelease) *Resolver {
	return &Resolver{info: info}
}

func (r *Resolver) Init() error {
	systemLang, err := locale.GetLanguage()
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Can't detect system language"))
	}
	if systemLang == "" {
		systemLang = "en"
	}

	r.names, err = overrides.Resolve(
		r.info,
		overrides.DefaultOpts.
			WithLanguages([]string{systemLang}),
	)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error resolving overrides"))
	}

	return nil
}

func (r *Resolver) Resolve(pkg *Package) {
	ResolvePackage(pkg, r.names)
}
