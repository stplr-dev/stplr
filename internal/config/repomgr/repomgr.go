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

package repomgr

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"go.stplr.dev/stplr/internal/config/internal/sources"
	"go.stplr.dev/stplr/internal/config/savers"
	"go.stplr.dev/stplr/pkg/types"
)

var (
	ErrSystemRepo   = errors.New("repo is provided by a system package; remove the providing package to delete it")
	ErrRepoNotFound = errors.New("repo not found")
)

// RepoRegistry merges repos from multiple directory sources and manages writes.
// The loading priority (later wins over earlier for the same name) is:
//  1. systemDir  (/usr/lib/stplr/repos.d)
//  2. inline repos from stplr.toml (passed to LoadAll)
//  3. userDir    (/etc/stplr/repos.d)
//
// Overrides from overridesDir are applied last, on top of the merged base.
type RepoRegistry struct {
	systemDir    string
	userDir      string
	overridesDir string
	writer       savers.RepoDirWriterExecutor

	// Populated by the last LoadAll call.
	userFiles   map[string]bool
	systemFiles map[string]bool
}

func New(systemDir, userDir, overridesDir string) *RepoRegistry {
	return &RepoRegistry{
		systemDir:    systemDir,
		userDir:      userDir,
		overridesDir: overridesDir,
	}
}

func (rr *RepoRegistry) SetWriter(w savers.RepoDirWriterExecutor) {
	rr.writer = w
}

// LoadAll builds the merged repo list. inlineRepos comes from the [[repo]] array
// in stplr.toml and slots between system and user directories in priority.
func (rr *RepoRegistry) LoadAll(inlineRepos []types.Repo) ([]types.RepoWithMeta, error) {
	systemSrc := sources.RepoDirSource{Dir: rr.systemDir, Origin: types.RepoOriginSystem}
	userSrc := sources.RepoDirSource{Dir: rr.userDir, Origin: types.RepoOriginUser}
	overrideSrc := sources.RepoOverrideSource{Dir: rr.overridesDir}

	systemRepos, err := systemSrc.LoadRepos()
	if err != nil {
		return nil, fmt.Errorf("load system repos: %w", err)
	}
	userRepos, err := userSrc.LoadRepos()
	if err != nil {
		return nil, fmt.Errorf("load user repos: %w", err)
	}
	overrides, err := overrideSrc.Load()
	if err != nil {
		return nil, fmt.Errorf("load repo overrides: %w", err)
	}

	// ordered map: insertion order preserved via separate slice
	order := make([]string, 0)
	byName := make(map[string]types.RepoWithMeta)

	insert := func(m types.RepoWithMeta) {
		if _, exists := byName[m.Name]; !exists {
			order = append(order, m.Name)
		}
		byName[m.Name] = m
	}

	for _, r := range systemRepos {
		insert(r)
	}
	for _, r := range inlineRepos {
		insert(types.RepoWithMeta{Repo: r, Origin: types.RepoOriginInline})
	}
	for _, r := range userRepos {
		insert(r)
	}

	userFiles := make(map[string]bool, len(userRepos))
	for _, r := range userRepos {
		userFiles[r.Name] = true
	}
	systemFiles := make(map[string]bool, len(systemRepos))
	for _, r := range systemRepos {
		systemFiles[r.Name] = true
	}
	rr.userFiles = userFiles
	rr.systemFiles = systemFiles

	result := make([]types.RepoWithMeta, 0, len(order))
	for _, name := range order {
		entry := byName[name]
		if o, ok := overrides[name]; ok {
			entry.Repo = types.ApplyOverride(entry.Repo, o)
		}
		result = append(result, entry)
	}
	return result, nil
}

// IsSystemRepo returns true if the repo originates from systemDir and has no
// user file in userDir. Must be called after LoadAll.
func (rr *RepoRegistry) IsSystemRepo(name string) bool {
	return !rr.userFiles[name]
}

// WriteUserRepo creates or overwrites userDir/<name>.toml.
func (rr *RepoRegistry) WriteUserRepo(repo types.Repo) error {
	data, err := toml.Marshal(repo)
	if err != nil {
		return fmt.Errorf("marshal repo: %w", err)
	}
	if rr.writer != nil {
		return rr.writer.WriteUserRepo(context.Background(), repo.Name, data)
	}
	if err := os.MkdirAll(rr.userDir, 0o755); err != nil {
		return fmt.Errorf("create user repos dir: %w", err)
	}
	return os.WriteFile(filepath.Join(rr.userDir, repo.Name+".toml"), data, 0o644) //gosec:disable G306 -- repo config files in /etc must be world-readable
}

// RemoveUserRepo deletes userDir/<name>.toml.
// Returns ErrSystemRepo if the repo exists only in systemDir.
// Returns ErrRepoNotFound if the repo is not known at all.
func (rr *RepoRegistry) RemoveUserRepo(name string) error {
	// Pre-check in-memory state (populated by LoadAll) before attempting the file
	// operation. This avoids relying on os.ErrNotExist surviving the RPC boundary
	// when using a writer plugin (error type info is lost through net/rpc serialization).
	if !rr.userFiles[name] {
		if rr.systemFiles[name] {
			return ErrSystemRepo
		}
		return ErrRepoNotFound
	}
	if rr.writer != nil {
		return rr.writer.RemoveUserRepo(context.Background(), name)
	}
	return os.Remove(filepath.Join(rr.userDir, name+".toml"))
}

// WriteOverride writes or updates overridesDir/<name>.toml.
// Fields set in o replace any existing values in the file; other existing fields are preserved.
func (rr *RepoRegistry) WriteOverride(name string, o types.RepoOverride) error {
	// Load existing override to preserve fields not included in this call.
	overrideSrc := sources.RepoOverrideSource{Dir: rr.overridesDir}
	existing, err := overrideSrc.Load()
	if err != nil {
		return err
	}

	merged := types.MergeOverrides(existing[name], o)
	data, err := marshalOverride(merged)
	if err != nil {
		return err
	}

	if rr.writer != nil {
		return rr.writer.WriteOverride(context.Background(), name, data)
	}
	if err := os.MkdirAll(rr.overridesDir, 0o755); err != nil {
		return fmt.Errorf("create overrides dir: %w", err)
	}
	return os.WriteFile(filepath.Join(rr.overridesDir, name+".toml"), data, 0o644) //gosec:disable G306 -- repo config files in /etc must be world-readable
}

// RemoveOverride deletes overridesDir/<name>.toml, removing all overrides for the repo.
// Returns ErrRepoNotFound if the repo is not known. Succeeds silently if no override file exists.
func (rr *RepoRegistry) RemoveOverride(name string) error {
	if !rr.userFiles[name] && !rr.systemFiles[name] {
		return ErrRepoNotFound
	}
	if rr.writer != nil {
		return rr.writer.RemoveOverride(context.Background(), name)
	}
	err := os.Remove(filepath.Join(rr.overridesDir, name+".toml"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// UpdateFromPull saves an updated repo to userDir/<name>.toml after a pull.
// System repos are skipped.
func (rr *RepoRegistry) UpdateFromPull(name string, updated types.Repo) error {
	if rr.IsSystemRepo(name) {
		return nil
	}
	return rr.WriteUserRepo(updated)
}

func marshalOverride(o types.RepoOverride) ([]byte, error) {
	data, err := toml.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("marshal override: %w", err)
	}
	return data, nil
}
