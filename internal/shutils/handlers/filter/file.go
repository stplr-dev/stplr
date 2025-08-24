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

package filter

import (
	"io/fs"
	"path/filepath"

	"github.com/spf13/afero"
)

type File struct {
	file afero.File
	pred Predicate
}

// Close implements afero.File.
func (f *File) Close() error {
	return f.file.Close()
}

// Name implements afero.File.
func (f *File) Name() string {
	return f.file.Name()
}

// Read implements afero.File.
func (f *File) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

// ReadAt implements afero.File.
func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	return f.file.ReadAt(p, off)
}

// Readdir implements afero.File.
func (f *File) Readdir(count int) (res []fs.FileInfo, err error) {
	var infos []fs.FileInfo
	infos, err = f.file.Readdir(count)
	if err != nil {
		return nil, err
	}

	for _, i := range infos {
		fullPath := filepath.Join(f.file.Name(), i.Name())
		if f.pred(fullPath) {
			res = append(res, i)
		}
	}

	return res, nil
}

// Readdirnames implements afero.File.
func (f *File) Readdirnames(n int) (names []string, err error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}

	for _, i := range infos {
		names = append(names, i.Name())
	}

	return names, nil
}

// Seek implements afero.File.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

// Stat implements afero.File.
func (f *File) Stat() (fs.FileInfo, error) {
	return f.file.Stat()
}

// Sync implements afero.File.
func (f *File) Sync() error {
	return f.file.Sync()
}

// Truncate implements afero.File.
func (f *File) Truncate(size int64) error {
	return f.file.Truncate(size)
}

// Write implements afero.File.
func (f *File) Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}

// WriteAt implements afero.File.
func (f *File) WriteAt(p []byte, off int64) (n int, err error) {
	return f.file.WriteAt(p, off)
}

// WriteString implements afero.File.
func (f *File) WriteString(s string) (ret int, err error) {
	return f.file.WriteString(s)
}
