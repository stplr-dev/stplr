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

package db_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.stplr.dev/stplr/pkg/staplerfile"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/db"
)

type TestALRConfig struct{}

func (c *TestALRConfig) GetPaths() *config.Paths {
	return &config.Paths{
		DBPath: ":memory:",
	}
}

func prepareDb() *db.Database {
	database := db.New(&TestALRConfig{})
	database.Init(context.Background())
	return database
}

var testPkg = staplerfile.Package{
	Name:    "test",
	Version: "0.0.1",
	Release: 1,
	Epoch:   2,
	Description: staplerfile.OverridableFromMap(map[string]string{
		"en": "Test package",
		"ru": "Проверочный пакет",
	}),
	Homepage: staplerfile.OverridableFromMap(map[string]string{
		"en": "https://gitea.plemya-x.ru/xpamych/ALR",
	}),
	Maintainer: staplerfile.OverridableFromMap(map[string]string{
		"en": "Evgeniy Khramov <xpamych@yandex.ru>",
		"ru": "Евгений Храмов <xpamych@yandex.ru>",
	}),
	Architectures: []string{"arm64", "amd64"},
	Licenses:      []string{"GPL-3.0-or-later"},
	Provides:      []string{"test"},
	Conflicts:     []string{"test"},
	Replaces:      []string{"test-old"},
	Depends: staplerfile.OverridableFromMap(map[string][]string{
		"": {"sudo"},
	}),
	BuildDepends: staplerfile.OverridableFromMap(map[string][]string{
		"":     {"golang"},
		"arch": {"go"},
	}),
	Repository:        "default",
	Summary:           staplerfile.OverridableFromMap(map[string]string{}),
	Group:             staplerfile.OverridableFromMap(map[string]string{}),
	OptDepends:        staplerfile.OverridableFromMap(map[string][]string{}),
	NonFreeUrl:        staplerfile.OverridableFromMap(map[string]string{}),
	AppStreamAppID:    staplerfile.OverridableFromMap(map[string]string{}),
	AppStreamMetaInfo: staplerfile.OverridableFromMap(map[string]string{}),
}

func TestInit(t *testing.T) {
	ctx := context.Background()
	database := prepareDb()
	defer database.Close()

	ver, ok := database.GetVersion(ctx)
	if !ok {
		t.Errorf("Expected version to be present")
	} else if ver != db.CurrentVersion {
		t.Errorf("Expected version %d, got %d", db.CurrentVersion, ver)
	}
}

func TestInsertPackage(t *testing.T) {
	ctx := context.Background()
	database := prepareDb()
	defer database.Close()

	err := database.InsertPackage(ctx, testPkg)
	assert.NoError(t, err)

	pkgs, err := database.GetPkgs(ctx, "name = 'test' AND repository = 'default'")
	assert.NoError(t, err)

	assert.Len(t, pkgs, 1)
	assert.Equal(t, testPkg, pkgs[0])
}

func TestGetPkgs(t *testing.T) {
	ctx := context.Background()
	database := prepareDb()
	defer database.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"

	err := database.InsertPackage(ctx, x1)
	assert.NoError(t, err)

	err = database.InsertPackage(ctx, x2)
	assert.NoError(t, err)

	pkgs, err := database.GetPkgs(ctx, "name LIKE 'x%'")
	assert.NoError(t, err)

	for _, dbPkg := range pkgs {
		assert.True(t, strings.HasPrefix(dbPkg.Name, "x"), "Expected package name to start with 'x', got %s", dbPkg.Name)
	}
}

func TestGetPkg(t *testing.T) {
	ctx := context.Background()
	database := prepareDb()
	defer database.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"

	err := database.InsertPackage(ctx, x1)
	assert.NoError(t, err)

	err = database.InsertPackage(ctx, x2)
	assert.NoError(t, err)

	pkg, err := database.GetPkg("name LIKE 'x%'")
	assert.NoError(t, err)

	assert.Equal(t, x1, *pkg)
}

func TestDeletePkgs(t *testing.T) {
	ctx := context.Background()
	database := prepareDb()
	defer database.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"

	err := database.InsertPackage(ctx, x1)
	assert.NoError(t, err)

	err = database.InsertPackage(ctx, x2)
	assert.NoError(t, err)

	err = database.DeletePkgs(ctx, "name = 'x1'")
	assert.NoError(t, err)
}

func TestJsonArrayContains(t *testing.T) {
	ctx := context.Background()
	database := prepareDb()
	defer database.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"
	x2.Provides = append(x2.Provides, "x")

	err := database.InsertPackage(ctx, x1)
	assert.NoError(t, err)

	err = database.InsertPackage(ctx, x2)
	assert.NoError(t, err)

	pkgs, err := database.GetPkgs(ctx, "name = 'x2'")
	assert.NoError(t, err)

	if assert.Len(t, pkgs, 1) {
		assert.Equal(t, "x2", pkgs[0].Name, "Expected package name to be 'x2'")
		// Verify the provides field contains 'x'
		assert.Contains(t, pkgs[0].Provides, "x")
	}
}
