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

// DO NOT EDIT MANUALLY. This file is generated.
package staplerfile

type packageResolved struct {
	Repository        string            `json:"repository"`
	Name              string            `json:"name"`
	BasePkgName       string            `json:"basepkg_name"`
	Version           string            `json:"version"`
	Release           int               `json:"release"`
	Epoch             uint              `json:"epoch"`
	Architectures     []string          `json:"architectures"`
	Licenses          []string          `json:"license"`
	Provides          []string          `json:"provides"`
	Conflicts         []string          `json:"conflicts"`
	Replaces          []string          `json:"replaces"`
	NonFree           bool              `json:"nonfree"`
	NonFreeUrl        string            `json:"nonfree_url"`
	NonFreeMsg        string            `json:"nonfree_msg"`
	NonFreeMsgFile    string            `json:"nonfree_msgfile"`
	Summary           string            `json:"summary"`
	Description       string            `json:"description"`
	Group             string            `json:"group"`
	Homepage          string            `json:"homepage"`
	Maintainer        string            `json:"maintainer"`
	Depends           []string          `json:"deps"`
	BuildDepends      []string          `json:"build_deps"`
	OptDepends        []string          `json:"opt_deps,omitempty"`
	Sources           []string          `json:"sources"`
	Checksums         []string          `json:"checksums,omitempty"`
	Backup            []string          `json:"backup"`
	Scripts           Scripts           `json:"scripts,omitempty"`
	AutoReqProvMethod string            `json:"auto_req_method"`
	AutoReq           []string          `json:"auto_req"`
	AutoReqSkipList   []string          `json:"auto_req_skiplist,omitempty"`
	AutoReqFilter     []string          `json:"auto_req_filter,omitempty"`
	AutoProv          []string          `json:"auto_prov"`
	AutoProvSkipList  []string          `json:"auto_prov_skiplist,omitempty"`
	AutoProvFilter    []string          `json:"auto_prov_filter,omitempty"`
	FireJailed        bool              `json:"firejailed"`
	FireJailProfiles  map[string]string `json:"firejail_profiles,omitempty"`
	DisableNetwork    bool              `json:"disable_network"`
}

func PackageToResolved(src *Package) packageResolved {
	return packageResolved{
		Repository:        src.Repository,
		Name:              src.Name,
		BasePkgName:       src.BasePkgName,
		Version:           src.Version,
		Release:           src.Release,
		Epoch:             src.Epoch,
		Architectures:     src.Architectures,
		Licenses:          src.Licenses,
		Provides:          src.Provides,
		Conflicts:         src.Conflicts,
		Replaces:          src.Replaces,
		NonFree:           src.NonFree,
		NonFreeUrl:        src.NonFreeUrl.Resolved(),
		NonFreeMsg:        src.NonFreeMsg.Resolved(),
		NonFreeMsgFile:    src.NonFreeMsgFile.Resolved(),
		Summary:           src.Summary.Resolved(),
		Description:       src.Description.Resolved(),
		Group:             src.Group.Resolved(),
		Homepage:          src.Homepage.Resolved(),
		Maintainer:        src.Maintainer.Resolved(),
		Depends:           src.Depends.Resolved(),
		BuildDepends:      src.BuildDepends.Resolved(),
		OptDepends:        src.OptDepends.Resolved(),
		Sources:           src.Sources.Resolved(),
		Checksums:         src.Checksums.Resolved(),
		Backup:            src.Backup.Resolved(),
		Scripts:           src.Scripts.Resolved(),
		AutoReqProvMethod: src.AutoReqProvMethod.Resolved(),
		AutoReq:           src.AutoReq.Resolved(),
		AutoReqSkipList:   src.AutoReqSkipList.Resolved(),
		AutoReqFilter:     src.AutoReqFilter.Resolved(),
		AutoProv:          src.AutoProv.Resolved(),
		AutoProvSkipList:  src.AutoProvSkipList.Resolved(),
		AutoProvFilter:    src.AutoProvFilter.Resolved(),
		FireJailed:        src.FireJailed.Resolved(),
		FireJailProfiles:  src.FireJailProfiles.Resolved(),
		DisableNetwork:    src.DisableNetwork.Resolved(),
	}
}

func ResolvePackage(pkg *Package, overrides []string) {
	pkg.NonFreeUrl.Resolve(overrides)
	pkg.NonFreeMsg.Resolve(overrides)
	pkg.NonFreeMsgFile.Resolve(overrides)
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
	pkg.AutoReqProvMethod.Resolve(overrides)
	pkg.AutoReq.Resolve(overrides)
	pkg.AutoReqSkipList.Resolve(overrides)
	pkg.AutoReqFilter.Resolve(overrides)
	pkg.AutoProv.Resolve(overrides)
	pkg.AutoProvSkipList.Resolve(overrides)
	pkg.AutoProvFilter.Resolve(overrides)
	pkg.FireJailed.Resolve(overrides)
	pkg.FireJailProfiles.Resolve(overrides)
	pkg.DisableNetwork.Resolve(overrides)
}
