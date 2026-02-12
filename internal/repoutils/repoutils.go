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
	"log/slog"
	"os"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/pelletier/go-toml/v2"
	"go.elara.ws/vercmp"

	"go.stplr.dev/stplr/internal/config"
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

	// If the version doesn't have a "v" prefix, it's not a standard version.
	// It may be "unknown" or a git version, but either way, there's no way
	// to compare it to the repo version, so only compare versions with the "v".
	if strings.HasPrefix(config.Version, "v") {
		if vercmp.Compare(config.Version, repocfg.Repo.MinVersion) == -1 {
			slog.Warn(
				gotext.Get("Stapler repo's minimum Stapler version is greater than the current version. Try updating Stapler if something doesn't work."))
		}
	}

	var repo types.Repo

	repo.URL = repocfg.Repo.URL
	repo.Ref = repocfg.Repo.Ref
	repo.Mirrors = repocfg.Repo.Mirrors
	repo.ReportUrl = repocfg.Repo.ReportUrl
	repo.Title = repocfg.Repo.Title
	repo.Summary = repocfg.Repo.Summary
	repo.Description = repocfg.Repo.Description
	repo.Homepage = repocfg.Repo.Homepage
	repo.Icon = repocfg.Repo.Icon

	return &repo, nil
}
