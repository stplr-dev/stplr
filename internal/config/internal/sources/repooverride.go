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

type RepoOverrideSource struct {
	Dir string
}

// tomlRepoOverride mirrors RepoOverride but uses pointer booleans so that
// go-toml can distinguish an absent field from an explicit false.
type tomlRepoOverride struct {
	Disabled *bool    `toml:"disabled"`
	Ref      *string  `toml:"ref"`
	URL      *string  `toml:"url"`
	Mirrors  []string `toml:"mirrors"`
}

// Load reads all *.toml files from Dir and returns a map from repo name to its override.
// The file stem is used as the repo name.
// Returns an empty map (not an error) if the directory does not exist.
func (s *RepoOverrideSource) Load() (map[string]types.RepoOverride, error) {
	entries, err := os.ReadDir(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]types.RepoOverride{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read override dir %s: %w", s.Dir, err)
	}

	result := make(map[string]types.RepoOverride)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".toml")
		path := filepath.Join(s.Dir, entry.Name())

		override, err := loadOverrideFile(path)
		if err != nil {
			return nil, fmt.Errorf("load override file %s: %w", path, err)
		}
		result[name] = override
	}
	return result, nil
}

func loadOverrideFile(path string) (types.RepoOverride, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return types.RepoOverride{}, err
	}
	var raw tomlRepoOverride
	if err := toml.Unmarshal(data, &raw); err != nil {
		return types.RepoOverride{}, err
	}
	return types.RepoOverride{
		Disabled: raw.Disabled,
		Ref:      raw.Ref,
		URL:      raw.URL,
		Mirrors:  raw.Mirrors,
	}, nil
}
