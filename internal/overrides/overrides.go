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

// Resolve generates a slice of possible override names in the order that they should be checked
func Resolve(info *distro.OSRelease, opts *Opts) ([]string, error) {
	if opts == nil {
		opts = DefaultOpts
	}

	if !opts.Overrides {
		return []string{opts.Name}, nil
	}

	langs, err := parseLangs(opts.Languages, opts.LanguageTags)
	if err != nil {
		return nil, err
	}

	architectures, err := cpu.CompatibleArches(cpu.Arch())
	if err != nil {
		return nil, err
	}

	distros := []string{info.ID}
	if opts.LikeDistros {
		distros = append(distros, info.Like...)
	}

	var out []string
	for _, lang := range langs {
		for _, distro := range distros {
			for _, arch := range architectures {
				out = append(out, opts.Name+"_"+arch+"_"+distro+"_"+lang)
			}

			out = append(out, opts.Name+"_"+distro+"_"+lang)
		}

		for _, arch := range architectures {
			out = append(out, opts.Name+"_"+arch+"_"+lang)
		}

		out = append(out, opts.Name+"_"+lang)
	}

	for _, distro := range distros {
		for _, arch := range architectures {
			out = append(out, opts.Name+"_"+arch+"_"+distro)
		}

		out = append(out, opts.Name+"_"+distro)
	}

	for _, arch := range architectures {
		out = append(out, opts.Name+"_"+arch)
	}

	out = append(out, opts.Name)

	for index, item := range out {
		out[index] = strings.TrimPrefix(item, "_")
	}

	return out, nil
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
	if info.ID == "altlinux" {
		return fmt.Sprintf("alt%d", release)
	}

	if info.ID == "fedora" || slices.Contains(info.Like, "fedora") {
		re := regexp.MustCompile(`platform:(\S+)`)
		match := re.FindStringSubmatch(info.PlatformID)
		if len(match) > 1 {
			return fmt.Sprintf("%d.%s", release, match[1])
		}
	}

	return fmt.Sprintf("%d", release)
}

func ParseReleasePlatformSpecific(s string, info *distro.OSRelease) (int, error) {
	if info.ID == "altlinux" {
		if strings.HasPrefix(s, "alt") {
			return strconv.Atoi(s[3:])
		}
	}

	if info.ID == "fedora" || slices.Contains(info.Like, "fedora") {
		parts := strings.SplitN(s, ".", 2)
		return strconv.Atoi(parts[0])
	}

	return strconv.Atoi(s)
}
