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
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/leonelquinteros/gotext"
)

type flatVarsStep struct{}

func FlatVarsStep() *flatVarsStep { return &flatVarsStep{} }

func (s *flatVarsStep) Name() string {
	return "flat vars"
}

func (s *flatVarsStep) Run(ctx context.Context, state *BuildState) error {
	buildDepends := []string{}
	optDepends := []string{}
	depends := []string{}
	sources := []string{}
	checksums := []string{}
	for _, pkg := range state.Packages {
		buildDepends = append(buildDepends, pkg.BuildDepends.Resolved()...)
		optDepends = append(optDepends, pkg.OptDepends.Resolved()...)
		depends = append(depends, pkg.Depends.Resolved()...)
		sources = append(sources, pkg.Sources.Resolved()...)
		checksums = append(checksums, pkg.Checksums.Resolved()...)
	}
	if len(sources) != len(checksums) {
		return errors.New(gotext.Get("The checksums array must be the same length as sources"))
	}
	sources, checksums = removeDuplicatesSources(sources, checksums)

	var sourceOverrides map[int]string
	var ignoreChecksums bool
	if state.Input != nil && state.Input.Opts != nil {
		sourceOverrides = state.Input.Opts.SourceOverrides
		ignoreChecksums = state.Input.Opts.IgnoreSourceChecksums
	}
	overriddenSources := make(map[int]bool, len(sourceOverrides))
	for idx, path := range sourceOverrides {
		if idx < 0 || idx >= len(sources) {
			return fmt.Errorf("%s", gotext.Get("source override index %d is out of range (package has %d sources)", idx, len(sources)))
		}
		originalSource := sources[idx]
		sources[idx] = path
		if ignoreChecksums {
			checksums[idx] = "SKIP"
		}
		overriddenSources[idx] = true

		// Preserve the original source name via ~name query parameter
		// so that the overridden file is saved with the expected filename.
		u, err := url.Parse(path)
		if err == nil {
			q := u.Query()
			if q.Get("~name") == "" {
				q.Set("~name", getSourceName(originalSource))
				u.RawQuery = q.Encode()
				sources[idx] = u.String()
			}
		}
	}

	state.FlatVars.Sources = sources
	state.FlatVars.Checksums = checksums
	state.FlatVars.OverriddenSources = overriddenSources
	state.FlatVars.BuildDepends = removeDuplicates(buildDepends)
	state.FlatVars.OptDepends = removeDuplicates(optDepends)
	state.FlatVars.Depends = removeDuplicates(depends)

	return nil
}

func getSourceName(src string) string {
	u, err := url.Parse(src)
	if err != nil {
		return src
	}
	query := u.Query()
	if name := query.Get("~name"); name != "" {
		return name
	}
	if u.Path != "" {
		return u.Path[strings.LastIndex(u.Path, "/")+1:]
	}
	return src
}
