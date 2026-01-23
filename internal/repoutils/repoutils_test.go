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

package repoutils_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/repoutils"
)

func TestRepoFromConfigFile(t *testing.T) {
	minimalConfig := `[repo]
url = "https://example.com/repo"
`
	minimalFile, err := os.CreateTemp("", "minimal-repo-*.toml")
	require.NoError(t, err)
	defer os.Remove(minimalFile.Name())
	_, err = minimalFile.WriteString(minimalConfig)
	require.NoError(t, err)
	minimalFile.Close()

	fullConfig := `[repo]
url = "https://example.com/repo"
ref = "main"
mirrors = ["https://mirror1.com/repo", "https://mirror2.com/repo"]
report_url = "https://example.com/report"
`
	fullFile, err := os.CreateTemp("", "full-repo-*.toml")
	require.NoError(t, err)
	defer os.Remove(fullFile.Name())
	_, err = fullFile.WriteString(fullConfig)
	require.NoError(t, err)
	fullFile.Close()

	invalidConfig := `[repo
url = "https://example.com/repo"
`
	invalidFile, err := os.CreateTemp("", "invalid-repo-*.toml")
	require.NoError(t, err)
	defer os.Remove(invalidFile.Name())
	_, err = invalidFile.WriteString(invalidConfig)
	require.NoError(t, err)
	invalidFile.Close()

	t.Run("Minimal config", func(t *testing.T) {
		repo, err := repoutils.RepoFromConfigFile(minimalFile.Name())
		require.NoError(t, err)
		require.NotNil(t, repo)
		require.Equal(t, "https://example.com/repo", repo.URL)
		require.Empty(t, repo.Ref)
		require.Empty(t, repo.Mirrors)
		require.Empty(t, repo.ReportUrl)
	})

	t.Run("Full config", func(t *testing.T) {
		repo, err := repoutils.RepoFromConfigFile(fullFile.Name())
		require.NoError(t, err)
		require.NotNil(t, repo)
		require.Equal(t, "https://example.com/repo", repo.URL)
		require.Equal(t, "main", repo.Ref)
		require.Equal(t, []string{"https://mirror1.com/repo", "https://mirror2.com/repo"}, repo.Mirrors)
		require.Equal(t, "https://example.com/report", repo.ReportUrl)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		repo, err := repoutils.RepoFromConfigFile("non-existent-file.toml")
		require.NoError(t, err)
		require.Nil(t, repo)
	})

	t.Run("Invalid TOML", func(t *testing.T) {
		repo, err := repoutils.RepoFromConfigFile(invalidFile.Name())
		require.Error(t, err)
		require.Nil(t, repo)
	})
}

func TestRepoFromConfigString(t *testing.T) {
	t.Run("Minimal config", func(t *testing.T) {
		cfg := `[repo]
url = "https://example.com/repo"
`
		repo, err := repoutils.RepoFromConfigString(cfg)
		require.NoError(t, err)
		require.NotNil(t, repo)
		require.Equal(t, "https://example.com/repo", repo.URL)
		require.Empty(t, repo.Ref)
		require.Empty(t, repo.Mirrors)
		require.Empty(t, repo.ReportUrl)
	})

	t.Run("Full config", func(t *testing.T) {
		cfg := `[repo]
url = "https://example.com/repo"
ref = "main"
mirrors = ["https://mirror1.com/repo", "https://mirror2.com/repo"]
report_url = "https://example.com/report"
`
		repo, err := repoutils.RepoFromConfigString(cfg)
		require.NoError(t, err)
		require.NotNil(t, repo)
		require.Equal(t, "https://example.com/repo", repo.URL)
		require.Equal(t, "main", repo.Ref)
		require.Equal(t, []string{"https://mirror1.com/repo", "https://mirror2.com/repo"}, repo.Mirrors)
		require.Equal(t, "https://example.com/report", repo.ReportUrl)
	})

	t.Run("Invalid TOML", func(t *testing.T) {
		cfg := `[repo
url = "https://example.com/repo"
`
		repo, err := repoutils.RepoFromConfigString(cfg)
		require.Error(t, err)
		require.Nil(t, repo)
	})
}
