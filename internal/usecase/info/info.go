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

package info

import (
	"context"
	"fmt"
	"io"
	"os"

	stdErrors "errors"

	"github.com/goccy/go-yaml"
	"github.com/jeandeaual/go-locale"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/cliprompts"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/overrides"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

var ErrPackageNotFound = stdErrors.New("package not found")

type useCase struct {
	rs   PackageFinder
	info *distro.OSRelease

	stdout io.Writer
}

type PackageFinder interface {
	FindPkgs(ctx context.Context, pkgs []string) (map[string][]staplerfile.Package, []string, error)
}

func New(rs PackageFinder, info *distro.OSRelease) *useCase {
	return &useCase{
		rs:     rs,
		info:   info,
		stdout: os.Stdout,
	}
}

type Options struct {
	All         bool
	Interactive bool
	Pkgs        []string
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	found, _, err := u.rs.FindPkgs(ctx, opts.Pkgs)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error finding packages"))
	}

	if len(found) == 0 {
		return errors.WrapIntoI18nError(ErrPackageNotFound, gotext.Get("Package not found"))
	}

	pkgs := cliprompts.FlattenPkgs(ctx, found, "show", opts.Interactive)

	systemLang, err := locale.GetLanguage()
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Can't detect system language"))
	}
	if systemLang == "" {
		systemLang = "en"
	}

	names, err := overrides.Resolve(
		u.info,
		overrides.DefaultOpts.
			WithLanguages([]string{systemLang}),
	)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error resolving overrides"))
	}

	for _, pkg := range pkgs {
		staplerfile.ResolvePackage(&pkg, names)
		view := staplerfile.NewPackageView(pkg)
		view.Resolved = !opts.All
		err = yaml.NewEncoder(u.stdout, yaml.UseJSONMarshaler(), yaml.OmitEmpty()).Encode(view)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error encoding script variables"))
		}
		fmt.Fprintln(u.stdout, "---")
	}

	return nil
}
