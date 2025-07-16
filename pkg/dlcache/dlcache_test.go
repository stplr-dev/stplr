// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package dlcache_test

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/pkg/dlcache"
)

type TestALRConfig struct {
	CacheDir string
}

func (c *TestALRConfig) GetPaths() *config.Paths {
	return &config.Paths{
		CacheDir: c.CacheDir,
	}
}

func prepare(t *testing.T) *TestALRConfig {
	t.Helper()

	dir, err := os.MkdirTemp("/tmp", "alr-dlcache-test.*")
	if err != nil {
		panic(err)
	}

	return &TestALRConfig{
		CacheDir: dir,
	}
}

func cleanup(t *testing.T, cfg *TestALRConfig) {
	t.Helper()
	os.Remove(cfg.CacheDir)
}

func TestNew(t *testing.T) {
	cfg := prepare(t)
	defer cleanup(t, cfg)

	dc := dlcache.New(cfg.GetPaths().CacheDir)

	ctx := context.Background()

	const id = "https://example.com"
	dir, err := dc.New(ctx, id)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	exp := filepath.Join(dc.BasePath(ctx), sha1sum(id))
	if dir != exp {
		t.Errorf("Expected %s, got %s", exp, dir)
	}

	fi, err := os.Stat(dir)
	if err != nil {
		t.Errorf("stat: expected no error, got %s", err)
	}

	if !fi.IsDir() {
		t.Errorf("Expected cache item to be a directory")
	}

	dir2, ok := dc.Get(ctx, id)
	if !ok {
		t.Errorf("Expected Get() to return valid value")
	}
	if dir2 != dir {
		t.Errorf("Expected %s from Get(), got %s", dir, dir2)
	}
}

func sha1sum(id string) string {
	h := sha1.New()
	_, _ = io.WriteString(h, id)
	return hex.EncodeToString(h.Sum(nil))
}
