// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2025 The Stapler Authors
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

package migrate_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/internal/app/deps"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/service/repos"
	"go.stplr.dev/stplr/internal/usecase/migrate"
)

// errStopAfterMigration causes Run to exit after migrateInlineRepos
// without trying to call RereadAll on a nil *repos.Repos.
var errStopAfterMigration = errors.New("stop")

type noopResetter struct{}

func (n *noopResetter) Reset(_ context.Context) error { return nil }

type fileWriter struct{ path string }

func (w *fileWriter) Write(b []byte) (int, error) {
	return len(b), os.WriteFile(w.path, b, 0o644)
}

func stopGetter() (*repos.Repos, deps.Cleanup, error) {
	return nil, func() {
		// The function is empty because it is a test function.
	}, errStopAfterMigration
}

func newMigrateTestConfig(t *testing.T) (*config.ALRConfig, string, string) {
	t.Helper()
	base := t.TempDir()
	cfgFile := filepath.Join(base, "stplr.toml")
	userDir := filepath.Join(base, "user")
	overridesDir := filepath.Join(base, "overrides")
	systemDir := filepath.Join(base, "system")

	cfg := config.New(
		config.WithSystemConfigPath(cfgFile),
		config.WithSystemConfigWriter(&fileWriter{path: cfgFile}),
		config.WithRepoDirs(systemDir, userDir, overridesDir),
	)
	return cfg, cfgFile, userDir
}

func runMigrate(t *testing.T, cfg *config.ALRConfig) {
	t.Helper()
	u := migrate.New(cfg, &noopResetter{}, &noopResetter{}, stopGetter)
	err := u.Run(context.Background())
	// errStopAfterMigration is expected; any other error is a test failure.
	if err != nil && !errors.Is(err, errStopAfterMigration) {
		t.Fatalf("unexpected error from migrate: %v", err)
	}
}

func TestMigrateInlineReposMigratedToFiles(t *testing.T) {
	cfg, cfgFile, userDir := newMigrateTestConfig(t)

	require.NoError(t, os.WriteFile(cfgFile, []byte(`
[[repo]]
name = "myrepo"
url = "https://example.com"
`), 0o644))

	require.NoError(t, cfg.Load())
	require.Len(t, cfg.InlineRepos(), 1)

	runMigrate(t, cfg)

	_, err := os.Stat(filepath.Join(userDir, "myrepo.toml"))
	assert.NoError(t, err, "myrepo.toml should exist in userDir after migration")
}

func TestMigrateIdempotentOnSecondRun(t *testing.T) {
	cfg, cfgFile, userDir := newMigrateTestConfig(t)

	require.NoError(t, os.WriteFile(cfgFile, []byte(`
[[repo]]
name = "myrepo"
url = "https://example.com"
`), 0o644))

	require.NoError(t, cfg.Load())
	runMigrate(t, cfg)

	require.NoError(t, cfg.Load())
	runMigrate(t, cfg)

	entries, err := os.ReadDir(userDir)
	require.NoError(t, err)
	tomlFiles := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".toml" {
			tomlFiles++
		}
	}
	assert.Equal(t, 1, tomlFiles, "exactly one repo file should exist after two runs")
}

func TestMigrateInlineReposCleared(t *testing.T) {
	cfg, cfgFile, _ := newMigrateTestConfig(t)

	require.NoError(t, os.WriteFile(cfgFile, []byte(`
[[repo]]
name = "myrepo"
url = "https://example.com"
`), 0o644))

	require.NoError(t, cfg.Load())
	runMigrate(t, cfg)

	// After migration the stplr.toml should no longer contain inline repos.
	require.NoError(t, cfg.Load())
	assert.Empty(t, cfg.InlineRepos(), "inline repos should be cleared from stplr.toml after migration")
}
