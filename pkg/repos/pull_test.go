// This file was originally part of the project "LURE - Linux User REpository", created by Elara Musayelyan.
// It has been modified as part of "ALR - Any Linux Repository" by Евгений Храмов.
//
// ALR - Any Linux Repository
// Copyright (C) 2025 Евгений Храмов
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

package repos_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"plemya-x.ru/alr/internal/config"
	"plemya-x.ru/alr/internal/db"
	database "plemya-x.ru/alr/internal/db"
	"plemya-x.ru/alr/internal/types"
	"plemya-x.ru/alr/pkg/repos"
)

type TestEnv struct {
	Ctx context.Context
	Cfg *TestALRConfig
	Db  *db.Database
}

type TestALRConfig struct {
	CacheDir string
	RepoDir  string
	PkgsDir  string
}

func (c *TestALRConfig) GetPaths(ctx context.Context) *config.Paths {
	return &config.Paths{
		DBPath:   ":memory:",
		CacheDir: c.CacheDir,
		RepoDir:  c.RepoDir,
		PkgsDir:  c.PkgsDir,
	}
}

func (c *TestALRConfig) Repos(ctx context.Context) []types.Repo {
	return []types.Repo{}
}

func prepare(t *testing.T) *TestEnv {
	t.Helper()

	cacheDir, err := os.MkdirTemp("/tmp", "alr-pull-test.*")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	repoDir := filepath.Join(cacheDir, "repo")
	err = os.MkdirAll(repoDir, 0o755)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	pkgsDir := filepath.Join(cacheDir, "pkgs")
	err = os.MkdirAll(pkgsDir, 0o755)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	cfg := &TestALRConfig{
		CacheDir: cacheDir,
		RepoDir:  repoDir,
		PkgsDir:  pkgsDir,
	}

	ctx := context.Background()

	db := database.New(cfg)
	db.Init(ctx)

	return &TestEnv{
		Cfg: cfg,
		Db:  db,
		Ctx: ctx,
	}
}

func cleanup(t *testing.T, e *TestEnv) {
	t.Helper()

	err := os.RemoveAll(e.Cfg.CacheDir)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	e.Db.Close()
}

func TestPull(t *testing.T) {
	e := prepare(t)
	defer cleanup(t, e)

	rs := repos.New(
		e.Cfg,
		e.Db,
	)

	err := rs.Pull(e.Ctx, []types.Repo{
		{
			Name: "default",
			URL:  "https://gitea.plemya-x.ru/Plemya-x/xpamych-alr-repo.git",
		},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	result, err := e.Db.GetPkgs(e.Ctx, "true")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	var pkgAmt int
	for result.Next() {
		var dbPkg db.Package
		err = result.StructScan(&dbPkg)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}
		pkgAmt++
	}

	if pkgAmt == 0 {
		t.Errorf("Expected at least 1 matching package, but got %d", pkgAmt)
	}
}
