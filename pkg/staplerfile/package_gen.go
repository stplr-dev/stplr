// SPDX-License-Identifier: GPL-3.0-or-later
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

// DO NOT EDIT MANUALLY. This file is generated.
package staplerfile

type packageResolved struct {
	Repository       string            `json:"repository"`
	Name             string            `json:"name"`
	BasePkgName      string            `json:"basepkg_name"`
	Version          string            `json:"version"`
	Release          int               `json:"release"`
	Epoch            uint              `json:"epoch"`
	Architectures    []string          `json:"architectures"`
	Licenses         []string          `json:"license"`
	Provides         []string          `json:"provides"`
	Conflicts        []string          `json:"conflicts"`
	Replaces         []string          `json:"replaces"`
	Summary          string            `json:"summary"`
	Description      string            `json:"description"`
	Group            string            `json:"group"`
	Homepage         string            `json:"homepage"`
	Maintainer       string            `json:"maintainer"`
	Depends          []string          `json:"deps"`
	BuildDepends     []string          `json:"build_deps"`
	OptDepends       []string          `json:"opt_deps,omitempty"`
	Sources          []string          `json:"sources"`
	Checksums        []string          `json:"checksums,omitempty"`
	Backup           []string          `json:"backup"`
	Scripts          Scripts           `json:"scripts,omitempty"`
	AutoReq          []string          `json:"auto_req"`
	AutoProv         []string          `json:"auto_prov"`
	AutoReqSkipList  []string          `json:"auto_req_skiplist,omitempty"`
	AutoProvSkipList []string          `json:"auto_prov_skiplist,omitempty"`
	FireJailed       bool              `json:"firejailed"`
	FireJailProfiles map[string]string `json:"firejail_profiles,omitempty"`
}

func PackageToResolved(src *Package) packageResolved {
	return packageResolved{
		Repository:       src.Repository,
		Name:             src.Name,
		BasePkgName:      src.BasePkgName,
		Version:          src.Version,
		Release:          src.Release,
		Epoch:            src.Epoch,
		Architectures:    src.Architectures,
		Licenses:         src.Licenses,
		Provides:         src.Provides,
		Conflicts:        src.Conflicts,
		Replaces:         src.Replaces,
		Summary:          src.Summary.Resolved(),
		Description:      src.Description.Resolved(),
		Group:            src.Group.Resolved(),
		Homepage:         src.Homepage.Resolved(),
		Maintainer:       src.Maintainer.Resolved(),
		Depends:          src.Depends.Resolved(),
		BuildDepends:     src.BuildDepends.Resolved(),
		OptDepends:       src.OptDepends.Resolved(),
		Sources:          src.Sources.Resolved(),
		Checksums:        src.Checksums.Resolved(),
		Backup:           src.Backup.Resolved(),
		Scripts:          src.Scripts.Resolved(),
		AutoReq:          src.AutoReq.Resolved(),
		AutoProv:         src.AutoProv.Resolved(),
		AutoReqSkipList:  src.AutoReqSkipList.Resolved(),
		AutoProvSkipList: src.AutoProvSkipList.Resolved(),
		FireJailed:       src.FireJailed.Resolved(),
		FireJailProfiles: src.FireJailProfiles.Resolved(),
	}
}

func ResolvePackage(pkg *Package, overrides []string) {
	pkg.Summary.Resolve(overrides)
	pkg.Description.Resolve(overrides)
	pkg.Group.Resolve(overrides)
	pkg.Homepage.Resolve(overrides)
	pkg.Maintainer.Resolve(overrides)
	pkg.Depends.Resolve(overrides)
	pkg.BuildDepends.Resolve(overrides)
	pkg.OptDepends.Resolve(overrides)
	pkg.Sources.Resolve(overrides)
	pkg.Checksums.Resolve(overrides)
	pkg.Backup.Resolve(overrides)
	pkg.Scripts.Resolve(overrides)
	pkg.AutoReq.Resolve(overrides)
	pkg.AutoProv.Resolve(overrides)
	pkg.AutoReqSkipList.Resolve(overrides)
	pkg.AutoProvSkipList.Resolve(overrides)
	pkg.FireJailed.Resolve(overrides)
	pkg.FireJailProfiles.Resolve(overrides)
}
