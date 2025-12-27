// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package overrides

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/text/language"

	"go.stplr.dev/stplr/internal/cpu"
	"go.stplr.dev/stplr/pkg/distro"
)

type Opts struct {
	Name         string
	Overrides    bool
	LikeDistros  bool
	Languages    []string
	LanguageTags []language.Tag
}

var DefaultOpts = &Opts{
	Overrides:   true,
	LikeDistros: true,
	Languages:   []string{"en"},
}

func genCombinations(variantsList ...[]string) []string {
	var results []string
	var dfs func(int, []string)
	dfs = func(idx int, current []string) {
		if idx == 0 {
			slices.Reverse(current)
			results = append(results, strings.Join(current, "_"))
			return
		}
		for _, v := range variantsList[idx-1] {
			dfs(idx-1, append(current, v))
		}
		dfs(idx-1, current)
	}
	dfs(len(variantsList), []string{})
	return results
}

// Resolve generates a slice of possible override names in the order that they should be checked
func Resolve(info *distro.OSRelease, opts *Opts) ([]string, error) {
	// Validate inputs
	if info == nil {
		return nil, fmt.Errorf("OSRelease info cannot be nil")
	}
	if opts == nil {
		opts = DefaultOpts
	}

	// If overrides are disabled, return only the base name
	if !opts.Overrides {
		return []string{opts.Name}, nil
	}

	// Parse languages
	langs, err := parseLangs(opts.Languages, opts.LanguageTags)
	if err != nil {
		return nil, fmt.Errorf("failed to parse languages: %w", err)
	}

	// Get compatible architectures
	arches, err := cpu.CompatibleArches(cpu.Arch())
	if err != nil {
		return nil, fmt.Errorf("failed to get compatible architectures: %w", err)
	}

	// Collect distributions
	distros := []string{info.ID}
	if opts.LikeDistros {
		distros = append(distros, info.Like...)
	}

	if info.ReleaseID != "" {
		origDistros := distros
		distros = []string{}
		for _, d := range origDistros {
			if d != "" {
				distros = append(distros, strings.Join([]string{d, info.ReleaseID}, "_"))
			}
		}
		distros = append(distros, origDistros...)
	}

	// comb := combGenerator{reverse: true}

	var result []string
	// for _, combination := range comb.Generate(arches, distros, langs) {
	for _, combination := range genCombinations(arches, distros, langs) {
		parts := []string{}
		if opts.Name != "" {
			parts = append(parts, opts.Name)
		}
		if combination != "" {
			parts = append(parts, combination)
		}
		result = append(result, strings.Join(parts, "_"))
	}

	return result, nil
}

func (o *Opts) WithName(name string) *Opts {
	out := &Opts{}
	*out = *o

	out.Name = name
	return out
}

func (o *Opts) WithOverrides(v bool) *Opts {
	out := &Opts{}
	*out = *o

	out.Overrides = v
	return out
}

func (o *Opts) WithLikeDistros(v bool) *Opts {
	out := &Opts{}
	*out = *o

	out.LikeDistros = v
	return out
}

func (o *Opts) WithLanguages(langs []string) *Opts {
	out := &Opts{}
	*out = *o

	out.Languages = langs
	return out
}

func parseLangs(langs []string, tags []language.Tag) ([]string, error) {
	out := make([]string, len(tags)+len(langs))
	for i, tag := range tags {
		base, _ := tag.Base()
		out[i] = base.String()
	}
	for i, lang := range langs {
		tag, err := language.Parse(lang)
		if err != nil {
			return nil, err
		}
		base, _ := tag.Base()
		out[len(tags)+i] = base.String()
	}
	slices.Sort(out)
	out = slices.Compact(out)
	return out, nil
}

func ReleasePlatformSpecific(release int, info *distro.OSRelease) string {
	if distro.IsIdEqualOrLike(info, "altlinux") {
		return fmt.Sprintf("alt%d", release)
	}

	if distro.IsIdEqualOrLike(info, "fedora") {
		re := regexp.MustCompile(`platform:(\S+)`)
		match := re.FindStringSubmatch(info.PlatformID)
		if len(match) > 1 {
			return fmt.Sprintf("%d.%s", release, match[1])
		}
	}

	return fmt.Sprintf("%d", release)
}

func ParseReleasePlatformSpecific(s string, info *distro.OSRelease) (int, error) {
	if distro.IsIdEqualOrLike(info, "altlinux") {
		if strings.HasPrefix(s, "alt") {
			return strconv.Atoi(s[3:])
		}
	}

	if distro.IsIdEqualOrLike(info, "fedora") {
		parts := strings.SplitN(s, ".", 2)
		return strconv.Atoi(parts[0])
	}

	return strconv.Atoi(s)
}
