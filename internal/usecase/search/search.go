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

package search

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/search"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/overrides"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type Searcher interface {
	Search(ctx context.Context, opts *search.SearchOptions) ([]staplerfile.Package, error)
}

type useCase struct {
	searcher Searcher
	info     *distro.OSRelease

	stdout io.Writer
}

type Options struct {
	Name        string
	Description string
	Repository  string
	Provides    string
	Format      string
	All         bool
}

func New(searcher Searcher, info *distro.OSRelease) *useCase {
	return &useCase{
		searcher: searcher,
		info:     info,

		stdout: os.Stdout,
	}
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	var resolver *staplerfile.Resolver

	if !opts.All {
		resolver = staplerfile.NewResolver(u.info)
		err := resolver.Init()
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error initializing resolver"))
		}
	}

	packages, err := u.searcher.Search(
		ctx,
		search.NewSearchOptions().
			WithName(opts.Name).
			WithDescription(opts.Description).
			WithRepository(opts.Repository).
			WithProvides(opts.Provides).
			Build(),
	)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("Error while executing search"))
	}

	return u.outputResults(packages, resolver, opts.Format, opts.All)
}

func (u *useCase) outputResults(packages []staplerfile.Package, resolver *staplerfile.Resolver, format string, all bool) error {
	var tmpl *template.Template
	var err error
	if format != "" {
		tmpl, err = template.New("format").Parse(format)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error parsing format template"))
		}
	}

	for _, pkg := range packages {
		if resolver != nil {
			resolver.Resolve(&pkg)
		}
		if tmpl != nil {
			err = tmpl.Execute(os.Stdout, &pkg)
			if err != nil {
				return errors.WrapIntoI18nError(err, gotext.Get("Error executing template"))
			}
			fmt.Fprintln(u.stdout)
		} else {
			fmt.Fprintln(u.stdout, pkg.Name)
		}
	}

	return nil
}
