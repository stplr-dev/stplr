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

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/pkg/types"
)

func TestConvertValueAllowedKeysHandled(t *testing.T) {
	for _, key := range config.AllowedKeys() {
		_, err := config.ConvertValue(key, "")
		if err != nil {
			assert.NotContains(t, err.Error(), "unknown config key", key)
		}
	}
}

func newTestConfig(t *testing.T) (*config.ALRConfig, string, string, string) {
	t.Helper()
	base := t.TempDir()
	systemDir := filepath.Join(base, "system")
	userDir := filepath.Join(base, "user")
	overridesDir := filepath.Join(base, "overrides")
	emptyCfg := filepath.Join(base, "stplr.toml")
	cfg := config.New(
		config.WithRepoDirs(systemDir, userDir, overridesDir),
		config.WithSystemConfigPath(emptyCfg),
	)
	return cfg, systemDir, userDir, overridesDir
}

func writeTomlFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

func TestALRConfigLoadWithRepoDirs(t *testing.T) {
	cfg, systemDir, userDir, _ := newTestConfig(t)
	writeTomlFile(t, systemDir, "sys.toml", `name = "sys"`)
	writeTomlFile(t, userDir, "usr.toml", `name = "usr"`)

	require.NoError(t, cfg.Load())
	repos := cfg.Repos()
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Name
	}
	assert.Contains(t, names, "sys")
	assert.Contains(t, names, "usr")
}

func TestALRConfigLoadOverridesApplied(t *testing.T) {
	cfg, systemDir, _, overridesDir := newTestConfig(t)
	writeTomlFile(t, systemDir, "repo.toml", `name = "repo"`)
	writeTomlFile(t, overridesDir, "repo.toml", `disabled = true`)

	require.NoError(t, cfg.Load())
	repos := cfg.Repos()
	require.Len(t, repos, 1)
	assert.True(t, repos[0].Disabled)
}

func TestALRConfigAddRepo(t *testing.T) {
	cfg, _, userDir, _ := newTestConfig(t)
	require.NoError(t, cfg.Load())

	require.NoError(t, cfg.AddRepo(types.Repo{Name: "newrepo", URL: "https://example.com"}))

	_, err := os.Stat(filepath.Join(userDir, "newrepo.toml"))
	assert.NoError(t, err)
}

func TestALRConfigRemoveRepoSystemRepo(t *testing.T) {
	cfg, systemDir, _, _ := newTestConfig(t)
	writeTomlFile(t, systemDir, "sys.toml", `name = "sys"`)

	require.NoError(t, cfg.Load())
	err := cfg.RemoveRepo("sys")
	assert.Error(t, err)
}

func TestALRConfigSetRepoOverride(t *testing.T) {
	cfg, _, _, overridesDir := newTestConfig(t)
	require.NoError(t, cfg.Load())

	boolTrue := true
	require.NoError(t, cfg.SetRepoOverride("repo", types.RepoOverride{Disabled: &boolTrue}))

	_, err := os.Stat(filepath.Join(overridesDir, "repo.toml"))
	assert.NoError(t, err)
}

func TestALRConfigUpdateRepoFromPull(t *testing.T) {
	t.Run("user repo is updated", func(t *testing.T) {
		cfg, _, userDir, _ := newTestConfig(t)
		writeTomlFile(t, userDir, "pulled.toml", `name = "pulled"`)
		require.NoError(t, cfg.Load())

		updated := types.Repo{Name: "pulled", URL: "https://example.com", Title: "Pulled Repo"}
		require.NoError(t, cfg.UpdateRepoFromPull("pulled", updated))

		_, err := os.Stat(filepath.Join(userDir, "pulled.toml"))
		assert.NoError(t, err)
	})

	t.Run("system repo is skipped", func(t *testing.T) {
		cfg, systemDir, userDir, _ := newTestConfig(t)
		writeTomlFile(t, systemDir, "sysrepo.toml", `name = "sysrepo"`)
		require.NoError(t, cfg.Load())

		updated := types.Repo{Name: "sysrepo", URL: "https://example.com", Title: "Updated"}
		require.NoError(t, cfg.UpdateRepoFromPull("sysrepo", updated))

		_, err := os.Stat(filepath.Join(userDir, "sysrepo.toml"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}
