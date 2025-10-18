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

package build

import (
	"context"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/utils"
)

type mainBuilderConfig interface {
	Config
	checksRunnerConfig
}

func NewMainBuilder(
	cfg mainBuilderConfig,
	mgr manager.Manager,
	repos PackageFinder,
	scriptExecutor ScriptExecutor,
	installerExecutor InstallerExecutor,
	out output.Output,
) (*Builder, error) {
	builder := NewBuilder(
		NewScriptResolver(cfg),
		scriptExecutor,
		NewLocalCacheExecutor(cfg),
		installerExecutor,
		NewLocalSourceDownloader(cfg, out),
		NewChecksRunner(mgr, cfg),
		NewNonFreeViewer(cfg),
		repos,
		NewScriptViewer(cfg),
	)

	return builder, nil
}

type PrepareResult struct {
	Installer InstallerExecutor
	Scripter  ScriptExecutor
}

func PrepareInstallerAndScripter(ctx context.Context) (res *PrepareResult, cleanup func(), err error) {
	var installerClose func()
	var scripterClose func()

	installer, installerClose, err := GetSafeInstaller(ctx)
	if err != nil {
		return nil, nil, err
	}

	if utils.IsRoot() {
		if err := utils.ExitIfCantDropCapsToBuilderUserNoPrivs(); err != nil {
			installerClose()
			return nil, nil, err
		}
	}

	scripter, scripterClose, err := GetSafeScriptExecutor(ctx)
	if err != nil {
		installerClose()
		return nil, nil, err
	}

	cleanup = func() {
		scripterClose()
		installerClose()
	}

	res = &PrepareResult{
		Installer: installer,
		Scripter:  scripter,
	}

	return res, cleanup, nil
}
