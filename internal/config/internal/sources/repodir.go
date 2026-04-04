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

package sources

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"go.stplr.dev/stplr/pkg/types"
)

type RepoDirSource struct {
	Dir    string
	Origin types.RepoOrigin
}

// LoadRepos reads all *.toml files from Dir, sorted by filename.
// Returns an empty slice (not an error) if the directory does not exist.
func (s *RepoDirSource) LoadRepos() ([]types.RepoWithMeta, error) {
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read repo dir %s: %w", s.Dir, err)
	}

	var repos []types.RepoWithMeta
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		path := filepath.Join(s.Dir, entry.Name())
		repo, err := loadRepoFile(path)
		if err != nil {
			return nil, fmt.Errorf("load repo file %s: %w", path, err)
		}

		if repo.Name == "" {
			repo.Name = strings.TrimSuffix(entry.Name(), ".toml")
		}

		repos = append(repos, types.RepoWithMeta{
			Repo:     repo,
			Origin:   s.Origin,
			FilePath: path,
		})
	}
	return repos, nil
}

func loadRepoFile(path string) (types.Repo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return types.Repo{}, err
	}
	var repo types.Repo
	if err := toml.Unmarshal(data, &repo); err != nil {
		return types.Repo{}, err
	}
	return repo, nil
}
