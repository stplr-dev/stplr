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
	"os"
	"path/filepath"

	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type ScriptResolverExecutor interface {
	ResolveScript(ctx context.Context, pkg *staplerfile.Package) *ScriptInfo
}

type ScriptResolver struct {
	cfg commonbuild.Config
}

func NewScriptResolver(cfg commonbuild.Config) *ScriptResolver {
	return &ScriptResolver{cfg: cfg}
}

type ScriptInfo struct {
	Script     string
	Repository string
}

func (s *ScriptResolver) ResolveScript(
	ctx context.Context,
	pkg *staplerfile.Package,
) *ScriptInfo {
	var repository, script string

	repodir := s.cfg.GetPaths().RepoDir
	repository = pkg.Repository

	// First, we check if there is a root Staplerfile in the repository
	rootScriptPath := filepath.Join(repodir, repository, "Staplerfile")
	if _, err := os.Stat(rootScriptPath); err == nil {
		// A repository with a single Staplerfile at the root
		script = rootScriptPath
	} else {
		// Multi-package repository - we are looking for Staplerfile in the subfolder
		var scriptPath string
		if pkg.BasePkgName != "" {
			scriptPath = filepath.Join(repodir, repository, pkg.BasePkgName, "Staplerfile")
		} else {
			scriptPath = filepath.Join(repodir, repository, pkg.Name, "Staplerfile")
		}
		script = scriptPath
	}

	return &ScriptInfo{
		Repository: repository,
		Script:     script,
	}
}
