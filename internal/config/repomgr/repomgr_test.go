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

package repomgr_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/config/repomgr"
	"go.stplr.dev/stplr/pkg/types"
)

func newTestRegistry(t *testing.T) (*repomgr.RepoRegistry, string, string, string) {
	t.Helper()
	base := t.TempDir()
	systemDir := filepath.Join(base, "system")
	userDir := filepath.Join(base, "user")
	overridesDir := filepath.Join(base, "overrides")
	rr := repomgr.New(systemDir, userDir, overridesDir)
	return rr, systemDir, userDir, overridesDir
}

func writeToml(t *testing.T, dir, filename, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644))
}

func TestLoadAllSystemOnly(t *testing.T) {
	rr, systemDir, _, _ := newTestRegistry(t)
	writeToml(t, systemDir, "altlinux.toml", `name = "altlinux"`)

	repos, err := rr.LoadAll(nil)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "altlinux", repos[0].Name)
	assert.Equal(t, types.RepoOriginSystem, repos[0].Origin)
}

func TestLoadAllUserOverridesSystem(t *testing.T) {
	rr, systemDir, userDir, _ := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"
url = "https://system.example.com"`)
	writeToml(t, userDir, "repo.toml", `name = "repo"
url = "https://user.example.com"`)

	repos, err := rr.LoadAll(nil)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "https://user.example.com", repos[0].URL)
	assert.Equal(t, types.RepoOriginGlobal, repos[0].Origin)
}

func TestLoadAllInlineBetweenSystemAndUser(t *testing.T) {
	rr, systemDir, userDir, _ := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"
url = "https://system.example.com"`)
	writeToml(t, userDir, "repo.toml", `name = "repo"
url = "https://user.example.com"`)

	inline := []types.Repo{{Name: "repo", URL: "https://inline.example.com"}}
	repos, err := rr.LoadAll(inline)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	// User dir has the highest priority
	assert.Equal(t, "https://user.example.com", repos[0].URL)
}

func TestLoadAllInlineAddsNewRepo(t *testing.T) {
	rr, systemDir, _, _ := newTestRegistry(t)
	writeToml(t, systemDir, "system-repo.toml", `name = "system-repo"`)

	inline := []types.Repo{{Name: "inline-repo", URL: "https://inline.example.com"}}
	repos, err := rr.LoadAll(inline)
	require.NoError(t, err)
	assert.Len(t, repos, 2)
}

func TestLoadAllOverrideApplied(t *testing.T) {
	rr, systemDir, _, overridesDir := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"`)
	writeToml(t, overridesDir, "repo.toml", `disabled = true`)

	repos, err := rr.LoadAll(nil)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.True(t, repos[0].Disabled)
}

func TestLoadAllOverridePreservedAfterUserReplace(t *testing.T) {
	rr, systemDir, userDir, overridesDir := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"
url = "https://system.example.com"`)
	writeToml(t, userDir, "repo.toml", `name = "repo"
url = "https://user.example.com"`)
	writeToml(t, overridesDir, "repo.toml", `disabled = true`)

	repos, err := rr.LoadAll(nil)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "https://user.example.com", repos[0].URL)
	assert.True(t, repos[0].Disabled)
}

func TestLoadAllEmptyDirs(t *testing.T) {
	rr, _, _, _ := newTestRegistry(t)
	repos, err := rr.LoadAll(nil)
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestIsSystemRepoTrueWhenNoUserFile(t *testing.T) {
	rr, systemDir, _, _ := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"`)

	_, err := rr.LoadAll(nil)
	require.NoError(t, err)
	assert.True(t, rr.IsSystemRepo("repo"))
}

func TestIsSystemRepoFalseWhenUserFileExists(t *testing.T) {
	rr, systemDir, userDir, _ := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"`)
	writeToml(t, userDir, "repo.toml", `name = "repo"`)

	_, err := rr.LoadAll(nil)
	require.NoError(t, err)
	assert.False(t, rr.IsSystemRepo("repo"))
}

func TestRemoveUserRepoSuccess(t *testing.T) {
	rr, _, userDir, _ := newTestRegistry(t)
	writeToml(t, userDir, "myrepo.toml", `name = "myrepo"`)

	_, err := rr.LoadAll(nil)
	require.NoError(t, err)

	require.NoError(t, rr.RemoveUserRepo("myrepo"))
	_, statErr := os.Stat(filepath.Join(userDir, "myrepo.toml"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestRemoveUserRepoSystemRepo(t *testing.T) {
	rr, systemDir, _, _ := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"`)

	_, err := rr.LoadAll(nil)
	require.NoError(t, err)

	err = rr.RemoveUserRepo("repo")
	assert.ErrorIs(t, err, repomgr.ErrSystemRepo)
}

func TestRemoveUserRepoNotFound(t *testing.T) {
	rr, _, _, _ := newTestRegistry(t)
	_, err := rr.LoadAll(nil)
	require.NoError(t, err)

	err = rr.RemoveUserRepo("nonexistent")
	assert.ErrorIs(t, err, repomgr.ErrRepoNotFound)
}

func TestWriteOverrideNewFile(t *testing.T) {
	rr, _, _, overridesDir := newTestRegistry(t)

	boolTrue := true
	err := rr.WriteOverride("repo", types.RepoOverride{Disabled: &boolTrue})
	require.NoError(t, err)

	data, readErr := os.ReadFile(filepath.Join(overridesDir, "repo.toml"))
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "disabled = true")
}

func TestWriteOverrideMergesWithExisting(t *testing.T) {
	rr, _, _, overridesDir := newTestRegistry(t)

	boolTrue := true
	require.NoError(t, rr.WriteOverride("repo", types.RepoOverride{Disabled: &boolTrue}))

	ref := "v2.0"
	require.NoError(t, rr.WriteOverride("repo", types.RepoOverride{Ref: &ref}))

	data, err := os.ReadFile(filepath.Join(overridesDir, "repo.toml"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "disabled = true")
	assert.Contains(t, content, "v2.0")
}

func TestUpdateFromPullSkipsSystemRepo(t *testing.T) {
	rr, systemDir, userDir, _ := newTestRegistry(t)
	writeToml(t, systemDir, "repo.toml", `name = "repo"`)

	_, err := rr.LoadAll(nil)
	require.NoError(t, err)

	updated := types.Repo{Name: "repo", URL: "https://example.com", Title: "My Repo"}
	require.NoError(t, rr.UpdateFromPull("repo", updated))

	_, err = os.Stat(filepath.Join(userDir, "repo.toml"))
	assert.True(t, os.IsNotExist(err), "system repo must not create a user file on pull")
}

func TestUpdateFromPullUpdatesUserRepo(t *testing.T) {
	rr, _, userDir, _ := newTestRegistry(t)
	writeToml(t, userDir, "repo.toml", `name = "repo"
url = "https://old.example.com"`)

	_, err := rr.LoadAll(nil)
	require.NoError(t, err)

	updated := types.Repo{Name: "repo", URL: "https://new.example.com", Title: "Updated"}
	require.NoError(t, rr.UpdateFromPull("repo", updated))

	data, err := os.ReadFile(filepath.Join(userDir, "repo.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "https://new.example.com")
}
