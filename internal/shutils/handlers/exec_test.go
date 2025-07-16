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

package handlers_test

import (
	"context"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/shutils/decoder"
	"go.stplr.dev/stplr/internal/shutils/handlers"
	"go.stplr.dev/stplr/pkg/distro"
)

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
		test-cmd "Hello, World"
		test-fb
	}

	package() {
		install-binary test
	}
`

var osRelease = &distro.OSRelease{
	ID:   "test_os",
	Like: []string{"arch"},
}

func TestExecFuncs(t *testing.T) {
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

	eh := handlers.ExecFuncs{
		"test-cmd": func(hc interp.HandlerContext, name string, args []string) error {
			if name != "test-cmd" {
				t.Errorf("Expected name to be 'test-cmd', got '%s'", name)
			}

			if len(args) < 1 {
				t.Fatalf("Expected at least one argument, got %d", len(args))
			}

			if args[0] != "Hello, World" {
				t.Errorf("Expected first argument to be 'Hello, World', got '%s'", args[0])
			}

			return nil
		},
	}

	fbInvoked := false
	fbHandler := func(context.Context, []string) error {
		fbInvoked = true
		return nil
	}

	err = fn(ctx, interp.ExecHandler(eh.ExecHandler(fbHandler)))
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if !fbInvoked {
		t.Errorf("Expected fallback handler to be invoked")
	}
}
