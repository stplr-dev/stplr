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

package repos_test

import (
	"reflect"
	"testing"

	"go.stplr.dev/stplr/internal/repos"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

func TestFindPkgs(t *testing.T) {
	e := prepare(t)
	defer cleanup(t, e)

	rs := repos.New(
		e.Cfg,
		e.Db,
	)

	err := rs.Pull(e.Ctx, []types.Repo{
		{
			Name: "default",
			URL:  "https://codeberg.org/stapler/repo-for-tests.git",
		},
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	found, notFound, err := rs.FindPkgs(
		e.Ctx,
		[]string{"nonexistentpackage1", "nonexistentpackage2"},
	)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if !reflect.DeepEqual(notFound, []string{"nonexistentpackage1", "nonexistentpackage2"}) {
		t.Errorf("Expected 'nonexistentpackage{1,2} not to be found")
	}

	if len(found) != 0 {
		t.Errorf("Expected 0 package found, got %d", len(found))
	}

	/*
		alrPkgs, ok := found["sta"]
		if !ok {
			t.Fatalf("Expected 'alr' packages to be found")
		}

		if len(alrPkgs) < 2 {
			t.Errorf("Expected two 'alr' packages to be found")
		}

		for i, pkg := range alrPkgs {
			if !strings.HasPrefix(pkg.Name, "sta") {
				t.Errorf("Expected package name of all found packages to start with 'alr', got %s on element %d", pkg.Name, i)
			}
		}
	*/
}

func TestFindPkgsEmpty(t *testing.T) {
	e := prepare(t)
	defer cleanup(t, e)

	rs := repos.New(
		e.Cfg,
		e.Db,
	)

	err := e.Db.InsertPackage(e.Ctx, alrsh.Package{
		Name:       "test1",
		Repository: "default",
		Version:    "0.0.1",
		Release:    1,
		Provides:   []string{""},
		Description: alrsh.OverridableFromMap(map[string]string{
			"en": "Test package 1",
			"ru": "Проверочный пакет 1",
		}),
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	err = e.Db.InsertPackage(e.Ctx, alrsh.Package{
		Name:       "test2",
		Repository: "default",
		Version:    "0.0.1",
		Release:    1,
		Provides:   []string{"test"},
		Description: alrsh.OverridableFromMap(map[string]string{
			"en": "Test package 2",
			"ru": "Проверочный пакет 2",
		}),
	})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	found, notFound, err := rs.FindPkgs(e.Ctx, []string{"test", ""})
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if len(notFound) != 0 {
		t.Errorf("Expected all packages to be found")
	}

	if len(found) != 1 {
		t.Errorf("Expected 1 package found, got %d", len(found))
	}

	testPkgs, ok := found["test"]
	if !ok {
		t.Fatalf("Expected 'test' packages to be found")
	}

	if len(testPkgs) != 1 {
		t.Errorf("Expected one 'test' package to be found, got %d", len(testPkgs))
	}

	if testPkgs[0].Name != "test2" {
		t.Errorf("Expected 'test2' package, got '%s'", testPkgs[0].Name)
	}
}
