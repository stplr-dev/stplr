// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
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

package repos

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/db"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type TestALRConfig struct{}

func (c *TestALRConfig) GetPaths() *config.Paths {
	return &config.Paths{
		DBPath: ":memory:",
	}
}

func (c *TestALRConfig) Repos() []types.Repo {
	return []types.Repo{
		{
			Name: "test",
			URL:  "https://test",
		},
	}
}

func createReadCloserFromString(input string) io.ReadCloser {
	reader := strings.NewReader(input)
	return struct {
		io.Reader
		io.Closer
	}{
		Reader: reader,
		Closer: io.NopCloser(reader),
	}
}

func TestUpdatePkg(t *testing.T) {
	type testCase struct {
		name   string
		file   string
		verify func(context.Context, *db.Database)
	}

	repo := types.Repo{
		Name: "test",
		URL:  "https://test",
	}

	for _, tc := range []testCase{
		{
			name: "single package",
			file: `name=foo
version='0.0.1'
release=1
desc="main desc"
deps=('sudo')
build_deps=('golang')
`,
			verify: func(ctx context.Context, database *db.Database) {
				result, err := database.GetPkgs(ctx, "1 = 1")
				assert.NoError(t, err)
				pkgCount := 0
				for _, pkg := range result {
					assert.Equal(t, "foo", pkg.Name)
					assert.Equal(t, alrsh.OverridableFromMap(map[string]string{"": "main desc"}), pkg.Description)
					assert.Equal(t, alrsh.OverridableFromMap(map[string][]string{"": {"sudo"}}), pkg.Depends)
					pkgCount++
				}
				assert.Equal(t, 1, pkgCount)
			},
		},
		{
			name: "multiple package",
			file: `basepkg_name=foo
name=(
	bar
	buz
)
version='0.0.1'
release=1
desc="main desc"
deps=('sudo')
build_deps=('golang')
		
meta_bar() {
	desc="foo desc"
}
			
meta_buz() {
	deps+=('doas')
}
`,
			verify: func(ctx context.Context, database *db.Database) {
				result, err := database.GetPkgs(ctx, "1 = 1")
				assert.NoError(t, err)

				pkgCount := 0
				for _, pkg := range result {
					if err != nil {
						t.Errorf("Expected no error, got %s", err)
					}
					if pkg.Name == "bar" {
						assert.Equal(t, alrsh.OverridableFromMap(map[string]string{"": "foo desc"}), pkg.Description)
						assert.Equal(t, alrsh.OverridableFromMap(map[string][]string{"": {"sudo"}}), pkg.Depends)
					}

					if pkg.Name == "buz" {
						assert.Equal(t, alrsh.OverridableFromMap(map[string]string{"": "main desc"}), pkg.Description)
						assert.Equal(t, alrsh.OverridableFromMap(map[string][]string{"": {"sudo", "doas"}}), pkg.Depends)
					}
					pkgCount++
				}
				assert.Equal(t, 2, pkgCount)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &TestALRConfig{}
			ctx := context.Background()

			database := db.New(&TestALRConfig{})
			database.Init(ctx)

			rs := New(cfg, database)

			path, err := os.MkdirTemp("", "test-update-pkg")
			assert.NoError(t, err)
			defer os.RemoveAll(path)

			runner, err := rs.processRepoChangesRunner(path, path)
			assert.NoError(t, err)

			err = rs.updatePkg(ctx, repo, runner, createReadCloserFromString(
				tc.file,
			))
			assert.NoError(t, err)

			tc.verify(ctx, database)
		})
	}
}
