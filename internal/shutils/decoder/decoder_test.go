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

package decoder_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/shutils/decoder"
	"go.stplr.dev/stplr/pkg/distro"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
)

type BuildVars struct {
	Name          string                           `sh:"name,required"`
	Version       string                           `sh:"version,required"`
	Release       int                              `sh:"release,required"`
	Epoch         uint                             `sh:"epoch"`
	Description   alrsh.OverridableField[string]   `sh:"desc"`
	Homepage      string                           `sh:"homepage"`
	Maintainer    string                           `sh:"maintainer"`
	Architectures []string                         `sh:"architectures"`
	Licenses      []string                         `sh:"license"`
	Provides      []string                         `sh:"provides"`
	Conflicts     []string                         `sh:"conflicts"`
	Depends       []string                         `sh:"deps"`
	BuildDepends  alrsh.OverridableField[[]string] `sh:"build_deps"`
	Replaces      alrsh.OverridableField[[]string] `sh:"replaces"`
}

const testScript = `
	name='test'
	version='0.0.1'
	release=1
	epoch=2
	desc="Test package"
	homepage='https://gitea.plemya-x.ru/xpamych/ALR'
	maintainer='Евгений Храмов <xpamych@yandex.ru>'
	architectures=('arm64' 'amd64')
	license=('GPL-3.0-or-later')
	provides=('test')
	conflicts=('test')
	replaces=('test-old')
	replaces_test_os=('test-legacy')

	deps=('sudo')

	build_deps=('golang')
	build_deps_arch=('go')

	test() {
		echo "Test"
	}

	package() {
		install-binary test
	}
`

var osRelease = &distro.OSRelease{
	ID:   "test_os",
	Like: []string{"arch"},
}

func TestDecodeVars(t *testing.T) {
	ctx := context.Background()

	fl, err := syntax.NewParser().Parse(strings.NewReader(testScript), "Staplerfile")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	runner, err := interp.New()
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	err = runner.Run(ctx, fl)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	dec := decoder.New(osRelease, runner)

	var bv BuildVars
	err = dec.DecodeVars(&bv)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	expected := BuildVars{
		Name:    "test",
		Version: "0.0.1",
		Release: 1,
		Epoch:   2,
		Description: alrsh.OverridableFromMap(map[string]string{
			"": "Test package",
		}),
		Homepage:      "https://gitea.plemya-x.ru/xpamych/ALR",
		Maintainer:    "Евгений Храмов <xpamych@yandex.ru>",
		Architectures: []string{"arm64", "amd64"},
		Licenses:      []string{"GPL-3.0-or-later"},
		Provides:      []string{"test"},
		Conflicts:     []string{"test"},
		Replaces: alrsh.OverridableFromMap(map[string][]string{
			"":        {"test-old"},
			"test_os": {"test-legacy"},
		}),
		Depends: []string{"sudo"},
		BuildDepends: alrsh.OverridableFromMap(map[string][]string{
			"":     {"golang"},
			"arch": {"go"},
		}),
	}

	expected.Description.SetResolved("Test package")
	expected.Replaces.SetResolved([]string{"test-legacy"})
	expected.BuildDepends.SetResolved([]string{"go"})

	if !reflect.DeepEqual(bv, expected) {
		t.Errorf("Expected %v, got %v", expected, bv)
	}
}

func TestDecodeVarsMissing(t *testing.T) {
	ctx := context.Background()

	const testScript = `
		name='test'
		epoch=2
		desc="Test package"
		homepage='https://gitea.plemya-x.ru/xpamych/ALR'
		maintainer='Евгений Храмов <xpamych@yandex.ru>'
		architectures=('arm64' 'amd64')
		license=('GPL-3.0-or-later')
		provides=('test')
		conflicts=('test')
		replaces=('test-old')
		replaces_test_os=('test-legacy')

		deps=('sudo')

		build_deps=('golang')
		build_deps_arch=('go')

		test() {
			echo "Test"
		}

		package() {
			install-binary test
		}
	`

	fl, err := syntax.NewParser().Parse(strings.NewReader(testScript), "Staplerfile")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	runner, err := interp.New()
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	err = runner.Run(ctx, fl)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	dec := decoder.New(osRelease, runner)

	var bv BuildVars
	err = dec.DecodeVars(&bv)

	var notFoundErr decoder.VarNotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("Expected VarNotFoundError, got %T %v", err, err)
	}
}

func TestGetFunc(t *testing.T) {
	ctx := context.Background()

	fl, err := syntax.NewParser().Parse(strings.NewReader(testScript), "Staplerfile")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	runner, err := interp.New()
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	err = runner.Run(ctx, fl)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	dec := decoder.New(osRelease, runner)
	fn, ok := dec.GetFunc("test")
	if !ok {
		t.Fatalf("Expected test() function to exist")
	}

	buf := &bytes.Buffer{}
	err = fn(ctx, interp.StdIO(os.Stdin, buf, buf))
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if buf.String() != "Test\n" {
		t.Fatalf(`Expected "Test\n", got %#v`, buf.String())
	}
}
