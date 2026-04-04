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

package types

// RepoOrigin tracks where a repo was loaded from.
type RepoOrigin int

const (
	RepoOriginSystem RepoOrigin = iota // /usr/lib/stplr/repos.d
	RepoOriginUser                     // /etc/stplr/repos.d
	RepoOriginInline                   // [[repo]] in stplr.toml
)

// RepoWithMeta wraps a Repo with information about its source.
type RepoWithMeta struct {
	Repo
	Origin   RepoOrigin
	FilePath string // path to the source file; empty for inline repos
}

func ptrCopy[T any](p *T) *T {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// RepoOverride holds only the fields the user explicitly wants to override.
// Pointer fields: nil means "don't touch".
// Slice fields: nil means "don't touch", []string{} means "clear".
type RepoOverride struct {
	Disabled             *bool    `toml:"disabled"`
	Ref                  *string  `toml:"ref"`
	URL                  *string  `toml:"url"`
	Mirrors              []string `toml:"mirrors"`
	RequireSignedCommits *bool    `toml:"require_signed_commits"`
}

// ApplyOverride returns a copy of base with non-nil override fields applied.
func ApplyOverride(base Repo, o RepoOverride) Repo {
	if o.Disabled != nil {
		base.Disabled = *o.Disabled
	}
	if o.Ref != nil {
		base.Ref = *o.Ref
	}
	if o.URL != nil {
		base.URL = *o.URL
	}
	if o.Mirrors != nil {
		base.Mirrors = o.Mirrors
	}
	if o.RequireSignedCommits != nil {
		base.RequireSignedCommits = *o.RequireSignedCommits
	}
	return base
}

func MergeOverrides(a, b RepoOverride) RepoOverride {
	if b.Disabled != nil {
		a.Disabled = ptrCopy(b.Disabled)
	}
	if b.Ref != nil {
		a.Ref = ptrCopy(b.Ref)
	}
	if b.URL != nil {
		a.URL = ptrCopy(b.URL)
	}
	if b.Mirrors != nil {
		a.Mirrors = append([]string(nil), b.Mirrors...)
	}
	if b.RequireSignedCommits != nil {
		a.RequireSignedCommits = ptrCopy(b.RequireSignedCommits)
	}
	return a
}
