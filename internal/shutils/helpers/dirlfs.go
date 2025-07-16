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

package helpers

import (
	"io/fs"
	"os"
	"path/filepath"
)

// dirLfs implements fs.FS like os.DirFS but uses LStat instead of Stat.
// This means symbolic links are treated as links themselves rather than
// being followed to their targets.
type dirLfs struct {
	fs.FS
	dir string
}

func NewDirLFS(dir string) *dirLfs {
	return &dirLfs{
		FS:  os.DirFS(dir),
		dir: dir,
	}
}

func (d *dirLfs) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}

	fullPath := filepath.Join(d.dir, filepath.FromSlash(name))

	info, err := os.Lstat(fullPath)
	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}

	return info, nil
}
