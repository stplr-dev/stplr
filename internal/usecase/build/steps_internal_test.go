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

package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/build"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/mocks"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

func TestPrepareScriptArgsExecute(t *testing.T) {
	ctx := context.Background()
	state := &stepState{
		input: stepInputState{
			opts: Options{
				Script:     "script.sh",
				Subpackage: "subpkg",
			},
			deps: &appbuilder.AppDeps{
				Manager: &manager.APT{},
			},
		},
	}

	w, _ := os.Getwd()

	step := &prepareScriptArgs{}
	err := step.Execute(ctx, state)
	assert.NoError(t, err)
	assert.NotNil(t, state.input.scriptArgs)
	assert.Equal(t, filepath.Join(w, "script.sh"), state.input.scriptArgs.Script)
	assert.Equal(t, []string{"subpkg"}, state.input.scriptArgs.Packages)
}

func TestPrepareDbArgsExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	testPkgName := "test-pkg"
	state := &stepState{
		input: stepInputState{
			opts: Options{
				Package:     testPkgName,
				Interactive: false,
			},
			deps: &appbuilder.AppDeps{
				Manager: &manager.APT{},
			},
		},
	}

	pkg := staplerfile.Package{
		Name: testPkgName,
	}

	finder := mocks.NewMockPackageFinder(ctrl)
	finder.EXPECT().
		FindPkgs(ctx, []string{testPkgName}).
		Return(map[string][]staplerfile.Package{testPkgName: {pkg}}, []string{}, nil)

	step := &prepareDbArgs{finder: finder}
	err := step.Execute(ctx, state)
	assert.NoError(t, err)
	assert.NotNil(t, state.input.dbArgs)
	assert.Equal(t, &pkg, state.input.dbArgs.Package)
}

func TestCopyOutStepExecute(t *testing.T) {
	const wd = "/test/dir"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	copier := mocks.NewMockScriptCopier(ctrl)

	ctx := context.Background()
	state := &stepState{
		input: stepInputState{
			wd:      wd,
			origUid: 1000,
			origGid: 1000,
		},
		runtime: stepRuntimeState{
			copier: copier,
		},
		output: stepOutputState{
			out: []*build.BuiltDep{{Path: "/tmp/pkg1.rpm"}},
		},
	}

	copier.EXPECT().
		CopyOut(ctx, "/tmp/pkg1.rpm", "/test/dir/pkg1.rpm", 1000, 1000).
		Return(nil)

	step := &copyOutStep{copy: copyOutViaCopier}
	err := step.Execute(ctx, state)
	assert.NoError(t, err)
}

func TestCopyScriptStepExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	copier := mocks.NewMockScriptCopier(ctrl)

	fsys := fstest.MapFS{
		"Staplerfile": &fstest.MapFile{
			Data: []byte("content"),
		},
	}

	ctx := context.Background()
	state := &stepState{
		input: stepInputState{
			wd:      "/test/dir",
			origUid: 1000,
			origGid: 1000,
			scriptArgs: &build.BuildPackageFromScriptArgs{
				Script: "Staplerfile",
			},
			deps: &appbuilder.AppDeps{
				Info: &distro.OSRelease{},
			},
		},
		runtime: stepRuntimeState{
			copier: copier,
		},
		output: stepOutputState{
			out: []*build.BuiltDep{{Path: "/tmp/pkg1.rpm"}},
		},
	}

	copier.EXPECT().
		Copy(ctx, gomock.Any(), state.input.deps.Info).
		Return("/tmp/copied/script", nil)

	step := &copyScript{fsys}
	err := step.Execute(ctx, state)
	assert.NoError(t, err)
	assert.Equal(t, "/tmp/copied/script", state.input.scriptArgs.Script)
	assert.Len(t, state.runtime.cleanups, 1)
}

func TestCopyOutStepExecuteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	copier := mocks.NewMockScriptCopier(ctrl)

	ctx := context.Background()
	state := &stepState{
		input: stepInputState{
			wd:      "/test/dir",
			origUid: 1000,
			origGid: 1000,
		},
		runtime: stepRuntimeState{
			copier: copier,
		},
		output: stepOutputState{
			out: []*build.BuiltDep{{Path: "/tmp/pkg1.rpm"}},
		},
	}

	copyErr := fmt.Errorf("copy failed")
	copier.EXPECT().
		CopyOut(ctx, "/tmp/pkg1.rpm", "/test/dir/pkg1.rpm", 1000, 1000).
		Return(copyErr)

	step := &copyOutStep{copy: copyOutViaCopier}
	err := step.Execute(ctx, state)

	require.Error(t, err)
	var i18nErr *errors.I18nError
	require.ErrorAs(t, err, &i18nErr)
	assert.Equal(t, "Error moving the package", i18nErr.Message)
	assert.Contains(t, err.Error(), "copy failed")
}
