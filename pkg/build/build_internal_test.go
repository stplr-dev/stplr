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

package build

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"mvdan.cc/sh/v3/syntax"

	"gitea.plemya-x.ru/Plemya-x/ALR/internal/config"
	"gitea.plemya-x.ru/Plemya-x/ALR/internal/db"
	"gitea.plemya-x.ru/Plemya-x/ALR/internal/types"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/distro"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/manager"
)

type TestPackageFinder struct {
	FindPkgsFunc func(ctx context.Context, pkgs []string) (map[string][]db.Package, []string, error)
}

func (pf *TestPackageFinder) FindPkgs(ctx context.Context, pkgs []string) (map[string][]db.Package, []string, error) {
	if pf.FindPkgsFunc != nil {
		return pf.FindPkgsFunc(ctx, pkgs)
	}
	return map[string][]db.Package{}, []string{}, nil
}

type TestManager struct {
	NameFunc          func() string
	FormatFunc        func() string
	ExistsFunc        func() bool
	SetRootCmdFunc    func(cmd string)
	SyncFunc          func(opts *manager.Opts) error
	InstallFunc       func(opts *manager.Opts, pkgs ...string) error
	RemoveFunc        func(opts *manager.Opts, pkgs ...string) error
	UpgradeFunc       func(opts *manager.Opts, pkgs ...string) error
	InstallLocalFunc  func(opts *manager.Opts, files ...string) error
	UpgradeAllFunc    func(opts *manager.Opts) error
	ListInstalledFunc func(opts *manager.Opts) (map[string]string, error)
	IsInstalledFunc   func(pkg string) (bool, error)
}

func (m *TestManager) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "TestManager"
}

func (m *TestManager) Format() string {
	if m.FormatFunc != nil {
		return m.FormatFunc()
	}
	return "testpkg"
}

func (m *TestManager) Exists() bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc()
	}
	return true
}

func (m *TestManager) SetRootCmd(cmd string) {
	if m.SetRootCmdFunc != nil {
		m.SetRootCmdFunc(cmd)
	}
}

func (m *TestManager) Sync(opts *manager.Opts) error {
	if m.SyncFunc != nil {
		return m.SyncFunc(opts)
	}
	return nil
}

func (m *TestManager) Install(opts *manager.Opts, pkgs ...string) error {
	if m.InstallFunc != nil {
		return m.InstallFunc(opts, pkgs...)
	}
	return nil
}

func (m *TestManager) Remove(opts *manager.Opts, pkgs ...string) error {
	if m.RemoveFunc != nil {
		return m.RemoveFunc(opts, pkgs...)
	}
	return nil
}

func (m *TestManager) Upgrade(opts *manager.Opts, pkgs ...string) error {
	if m.UpgradeFunc != nil {
		return m.UpgradeFunc(opts, pkgs...)
	}
	return nil
}

func (m *TestManager) InstallLocal(opts *manager.Opts, files ...string) error {
	if m.InstallLocalFunc != nil {
		return m.InstallLocalFunc(opts, files...)
	}
	return nil
}

func (m *TestManager) UpgradeAll(opts *manager.Opts) error {
	if m.UpgradeAllFunc != nil {
		return m.UpgradeAllFunc(opts)
	}
	return nil
}

func (m *TestManager) ListInstalled(opts *manager.Opts) (map[string]string, error) {
	if m.ListInstalledFunc != nil {
		return m.ListInstalledFunc(opts)
	}
	return map[string]string{}, nil
}

func (m *TestManager) IsInstalled(pkg string) (bool, error) {
	if m.IsInstalledFunc != nil {
		return m.IsInstalledFunc(pkg)
	}
	return true, nil
}

type TestConfig struct{}

func (c *TestConfig) PagerStyle() string {
	return "native"
}

func (c *TestConfig) GetPaths() *config.Paths {
	return &config.Paths{
		CacheDir: "/tmp",
	}
}

func TestExecuteFirstPassIsSecure(t *testing.T) {
	cfg := &TestConfig{}
	pf := &TestPackageFinder{}
	info := &distro.OSRelease{}
	m := &TestManager{}

	opts := types.BuildOpts{
		Manager:     m,
		Interactive: false,
	}

	ctx := context.Background()

	b := NewBuilder(
		ctx,
		opts,
		pf,
		info,
		cfg,
	)

	tmpFile, err := os.CreateTemp("", "testfile-")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	_, err = os.Stat(tmpFilePath)
	assert.NoError(t, err)

	testScript := fmt.Sprintf(`name='test'
version=1.0.0
release=1
rm -f %s`, tmpFilePath)

	fl, err := syntax.NewParser().Parse(strings.NewReader(testScript), "alr.sh")
	assert.NoError(t, err)

	_, _, err = b.executeFirstPass(fl)
	assert.NoError(t, err)

	_, err = os.Stat(tmpFilePath)
	assert.NoError(t, err)
}

func TestExecuteFirstPassIsCorrect(t *testing.T) {
	type testCase struct {
		Name     string
		Script   string
		Opts     types.BuildOpts
		Expected func(t *testing.T, vars []*types.BuildVars)
	}

	for _, tc := range []testCase{{
		Name: "single package",
		Script: `name='test'
version='1.0.0'
release=1
epoch=2
desc="Test package"
homepage='https://example.com'
maintainer='Ivan Ivanov'
`,
		Opts: types.BuildOpts{
			Manager:     &TestManager{},
			Interactive: false,
		},
		Expected: func(t *testing.T, vars []*types.BuildVars) {
			assert.Equal(t, 1, len(vars))
			assert.Equal(t, vars[0].Name, "test")
			assert.Equal(t, vars[0].Version, "1.0.0")
			assert.Equal(t, vars[0].Release, int(1))
			assert.Equal(t, vars[0].Epoch, uint(2))
			assert.Equal(t, vars[0].Description, "Test package")
		},
	}, {
		Name: "multiple packages",
		Script: `name=(
	foo
	bar
)

version='0.0.1'
release=1
epoch=2
desc="Test package"

meta_foo() {
	desc="Foo package"
}

meta_bar() {

}
`,
		Opts: types.BuildOpts{
			Packages:    []string{"foo"},
			Manager:     &TestManager{},
			Interactive: false,
		},
		Expected: func(t *testing.T, vars []*types.BuildVars) {
			assert.Equal(t, 1, len(vars))
			assert.Equal(t, vars[0].Name, "foo")
			assert.Equal(t, vars[0].Description, "Foo package")
		},
	}} {
		t.Run(tc.Name, func(t *testing.T) {
			cfg := &TestConfig{}
			pf := &TestPackageFinder{}
			info := &distro.OSRelease{}

			ctx := context.Background()

			b := NewBuilder(
				ctx,
				tc.Opts,
				pf,
				info,
				cfg,
			)

			fl, err := syntax.NewParser().Parse(strings.NewReader(tc.Script), "alr.sh")
			assert.NoError(t, err)

			_, allVars, err := b.executeFirstPass(fl)
			assert.NoError(t, err)

			tc.Expected(t, allVars)
		})
	}
}
