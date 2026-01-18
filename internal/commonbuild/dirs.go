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

package commonbuild

import (
	"path/filepath"

	"go.stplr.dev/stplr/pkg/types"
)

func GetDirs(
	cfg Config,
	scriptPath string,
	basePkg string,
) (types.Directories, error) {
	pkgsDir := cfg.GetPaths().PkgsDir

	scriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return types.Directories{}, err
	}
	baseDir := filepath.Join(pkgsDir, basePkg)
	return types.Directories{
		BaseDir:   GetBaseDir(cfg, basePkg),
		SrcDir:    GetSrcDir(cfg, basePkg),
		PkgDir:    filepath.Join(baseDir, "pkg"),
		ScriptDir: GetScriptDir(scriptPath),
	}, nil
}

func GetBaseDir(cfg Config, basePkg string) string {
	return filepath.Join(cfg.GetPaths().PkgsDir, basePkg)
}

func GetSrcDir(cfg Config, basePkg string) string {
	return filepath.Join(GetBaseDir(cfg, basePkg), "src")
}

func GetScriptDir(scriptPath string) string {
	return filepath.Dir(scriptPath)
}
