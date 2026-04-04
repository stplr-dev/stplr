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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/config/internal/sources"
)

func TestRepoOverrideSourceNonExistentDir(t *testing.T) {
	s := sources.RepoOverrideSource{Dir: "/nonexistent/path"}
	result, err := s.Load()
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRepoOverrideSourceDisabledTrue(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "myrepo.toml"), `disabled = true`)

	s := sources.RepoOverrideSource{Dir: dir}
	result, err := s.Load()
	require.NoError(t, err)
	require.Contains(t, result, "myrepo")
	require.NotNil(t, result["myrepo"].Disabled)
	assert.True(t, *result["myrepo"].Disabled)
}

func TestRepoOverrideSourceDisabledFalse(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "myrepo.toml"), `disabled = false`)

	s := sources.RepoOverrideSource{Dir: dir}
	result, err := s.Load()
	require.NoError(t, err)
	require.Contains(t, result, "myrepo")
	require.NotNil(t, result["myrepo"].Disabled, "explicit false must produce non-nil pointer")
	assert.False(t, *result["myrepo"].Disabled)
}

func TestRepoOverrideSourcePartialFields(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "repo.toml"), `ref = "v2.0"`)

	s := sources.RepoOverrideSource{Dir: dir}
	result, err := s.Load()
	require.NoError(t, err)
	require.Contains(t, result, "repo")
	o := result["repo"]
	assert.Nil(t, o.Disabled, "absent disabled field must be nil")
	assert.Nil(t, o.URL, "absent url field must be nil")
	require.NotNil(t, o.Ref)
	assert.Equal(t, "v2.0", *o.Ref)
}

func TestRepoOverrideSourceMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "repo-a.toml"), `disabled = true`)
	writeFile(t, filepath.Join(dir, "repo-b.toml"), `url = "https://mirror.example.com"`)

	s := sources.RepoOverrideSource{Dir: dir}
	result, err := s.Load()
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "repo-a")
	assert.Contains(t, result, "repo-b")
}

func TestRepoOverrideSourceMirrorsSlice(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "repo.toml"), `mirrors = ["https://m1.com", "https://m2.com"]`)

	s := sources.RepoOverrideSource{Dir: dir}
	result, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"https://m1.com", "https://m2.com"}, result["repo"].Mirrors)
}

func TestRepoOverrideSourceEmptyFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "repo.toml"), ``)

	s := sources.RepoOverrideSource{Dir: dir}
	result, err := s.Load()
	require.NoError(t, err)
	require.Contains(t, result, "repo")
	o := result["repo"]
	assert.Nil(t, o.Disabled)
	assert.Nil(t, o.Ref)
	assert.Nil(t, o.URL)
	assert.Nil(t, o.Mirrors)
}
