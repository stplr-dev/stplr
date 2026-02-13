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

package types

// Config represents the ALR configuration file
type Config struct {
	RootCmd          string   `json:"rootCmd" koanf:"rootCmd"`
	UseRootCmd       bool     `json:"useRootCmd" koanf:"useRootCmd"`
	PagerStyle       string   `json:"pagerStyle" koanf:"pagerStyle"`
	IgnorePkgUpdates []string `json:"ignorePkgUpdates" koanf:"ignorePkgUpdates"`
	Repos            []Repo   `json:"repo" koanf:"repo"`
	AutoPull         bool     `json:"autoPull" koanf:"autoPull"`
	LogLevel         string   `json:"logLevel" koanf:"logLevel"`

	FirejailExclude            []string `json:"firejailExclude" koanf:"firejailExclude"`
	HideFirejailExcludeWarning bool     `json:"hideFirejailExcludeWarning" koanf:"hideFirejailExcludeWarning"`

	ForbidSkipInChecksums bool `json:"forbidSkipInChecksums" koanf:"forbidSkipInChecksums"`
	ForbidBuildCommand    bool `json:"forbidBuildCommand" koanf:"forbidBuildCommand"`
}

// Repo represents a Stapler repo within a configuration file
type Repo struct {
	Name      string   `json:"name" koanf:"name"`
	URL       string   `json:"url" koanf:"url"`
	Ref       string   `json:"ref" koanf:"ref"`
	Mirrors   []string `json:"mirrors" koanf:"mirrors"`
	ReportUrl string   `json:"report_url" koanf:"report_url"`

	Title       string `json:"title" koanf:"title"`
	Summary     string `json:"summary" koanf:"summary"`
	Description string `json:"description" koanf:"description"`
	Homepage    string `json:"homepage" koanf:"homepage"`
	Icon        string `json:"icon" koanf:"icon"`

	Disabled bool `json:"disabled" koanf:"disabled"`
}

func (r *Repo) MergeFrom(other *Repo) {
	updateIfNotEmpty := func(dst *string, src string) {
		if src != "" {
			*dst = src
		}
	}

	updateIfNotEmpty(&r.Name, other.Name)
	updateIfNotEmpty(&r.URL, other.URL)
	updateIfNotEmpty(&r.Ref, other.Ref)
	updateIfNotEmpty(&r.ReportUrl, other.ReportUrl)
	updateIfNotEmpty(&r.Title, other.Title)
	updateIfNotEmpty(&r.Summary, other.Summary)
	updateIfNotEmpty(&r.Description, other.Description)
	updateIfNotEmpty(&r.Homepage, other.Homepage)
	updateIfNotEmpty(&r.Icon, other.Icon)

	if len(other.Mirrors) > 0 {
		r.Mirrors = other.Mirrors
	}
}
