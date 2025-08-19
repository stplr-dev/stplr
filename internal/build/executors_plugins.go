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

	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

//go:generate go run ../../generators/plugin-generator InstallerExecutor ScriptExecutor ReposExecutor ScriptReader PackagesParser ScriptCopier

// The Executors interfaces must use context.Context as the first parameter,
// because the plugin-generator cannot generate code without it.

type InstallerExecutor interface {
	InstallLocal(ctx context.Context, paths []string, opts *manager.Opts) error
	Install(ctx context.Context, pkgs []string, opts *manager.Opts) error
	Remove(ctx context.Context, pkgs []string, opts *manager.Opts) error
	RemoveAlreadyInstalled(ctx context.Context, pkgs []string) ([]string, error)
}

type ScriptCopier interface {
	Copy(ctx context.Context, f *staplerfile.ScriptFile, info *distro.OSRelease) (string, error)
	CopyOut(ctx context.Context, from, dest string, uid, gid int) error
}

type ScriptExecutor interface {
	Read(ctx context.Context, scriptPath string) (*staplerfile.ScriptFile, error)
	ParsePackages(
		ctx context.Context,
		file *staplerfile.ScriptFile,
		packages []string,
		info distro.OSRelease,
	) (string, []*staplerfile.Package, error)
	PrepareDirs(
		ctx context.Context,
		input *BuildInput,
		basePkg string,
	) error
	ExecuteSecondPass(
		ctx context.Context,
		input *BuildInput,
		sf *staplerfile.ScriptFile,
		varsOfPackages []*staplerfile.Package,
		repoDeps []string,
		builtDeps []*BuiltDep,
		basePkg string,
	) ([]*BuiltDep, error)
}

type ReposExecutor interface {
	PullOneAndUpdateFromConfig(ctx context.Context, repo *types.Repo) (types.Repo, error)
}

type ScriptReader interface {
	Read(ctx context.Context, path string) (*staplerfile.ScriptFile, error)
}

type PackagesParser interface {
	ParsePackages(
		ctx context.Context,
		file *staplerfile.ScriptFile,
		packages []string,
		info distro.OSRelease,
	) (string, []*staplerfile.Package, error)
}
