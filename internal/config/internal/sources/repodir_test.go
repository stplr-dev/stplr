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

package sources_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/config/internal/sources"
	"go.stplr.dev/stplr/pkg/types"
)

func TestRepoDirSourceNonExistentDir(t *testing.T) {
	s := sources.RepoDirSource{Dir: "/nonexistent/path", Origin: types.RepoOriginSystem}
	repos, err := s.LoadRepos()
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestRepoDirSourceSingleFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "altlinux.toml"), `
name = "altlinux"
url = "https://github.com/altlinux/stplr-repo"
ref = "main"
`)

	s := sources.RepoDirSource{Dir: dir, Origin: types.RepoOriginSystem}
	repos, err := s.LoadRepos()
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "altlinux", repos[0].Name)
	assert.Equal(t, "https://github.com/altlinux/stplr-repo", repos[0].URL)
	assert.Equal(t, "main", repos[0].Ref)
	assert.Equal(t, types.RepoOriginSystem, repos[0].Origin)
	assert.Equal(t, filepath.Join(dir, "altlinux.toml"), repos[0].FilePath)
}

func TestRepoDirSourceFallbackName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "myrepo.toml"), `url = "https://example.com"`)

	s := sources.RepoDirSource{Dir: dir, Origin: types.RepoOriginGlobal}
	repos, err := s.LoadRepos()
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "myrepo", repos[0].Name, "filename stem used as fallback name")
}

func TestRepoDirSourceSortedOrder(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "zzz.toml"), `name = "zzz"`)
	writeFile(t, filepath.Join(dir, "aaa.toml"), `name = "aaa"`)
	writeFile(t, filepath.Join(dir, "mmm.toml"), `name = "mmm"`)

	s := sources.RepoDirSource{Dir: dir, Origin: types.RepoOriginGlobal}
	repos, err := s.LoadRepos()
	require.NoError(t, err)
	require.Len(t, repos, 3)
	assert.Equal(t, "aaa", repos[0].Name)
	assert.Equal(t, "mmm", repos[1].Name)
	assert.Equal(t, "zzz", repos[2].Name)
}

func TestRepoDirSourceSkipsNonTomlFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "repo.toml"), `name = "repo"`)
	writeFile(t, filepath.Join(dir, "readme.txt"), `not a repo`)
	writeFile(t, filepath.Join(dir, "other.yaml"), `name: other`)

	s := sources.RepoDirSource{Dir: dir, Origin: types.RepoOriginGlobal}
	repos, err := s.LoadRepos()
	require.NoError(t, err)
	assert.Len(t, repos, 1)
}

func TestRepoDirSourceInvalidToml(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "bad.toml"), `this is not valid toml ][`)

	s := sources.RepoDirSource{Dir: dir, Origin: types.RepoOriginGlobal}
	_, err := s.LoadRepos()
	assert.Error(t, err)
}

func TestRepoDirSourceSnakeCaseFields(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "repo.toml"), `
name = "repo"
url = "https://example.com"
report_url = "https://bugs.example.com"
disabled = true
`)

	s := sources.RepoDirSource{Dir: dir, Origin: types.RepoOriginSystem}
	repos, err := s.LoadRepos()
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "https://bugs.example.com", repos[0].ReportUrl)
	assert.True(t, repos[0].Disabled)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
