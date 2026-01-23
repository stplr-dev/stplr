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

package repoutils

import (
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"go.stplr.dev/stplr/pkg/types"
)

func RepoFromConfigFile(file string) (*types.Repo, error) {
	fl, err := os.Open(file)
	if err != nil {
		// for compatibility only
		return nil, nil
	}
	defer fl.Close()

	return repoFromReader(fl)
}

func RepoFromConfigString(cfg string) (*types.Repo, error) {
	r := strings.NewReader(cfg)
	return repoFromReader(r)
}

func repoFromReader(r interface {
	Read([]byte) (int, error)
},
) (*types.Repo, error) {
	var repocfg types.RepoConfig
	if err := toml.NewDecoder(r).Decode(&repocfg); err != nil {
		return nil, err
	}

	var repo types.Repo

	if repocfg.Repo.URL != "" {
		repo.URL = repocfg.Repo.URL
	}
	if repocfg.Repo.Ref != "" {
		repo.Ref = repocfg.Repo.Ref
	}
	if len(repocfg.Repo.Mirrors) > 0 {
		repo.Mirrors = repocfg.Repo.Mirrors
	}
	repo.ReportUrl = repocfg.Repo.ReportUrl

	return &repo, nil
}
