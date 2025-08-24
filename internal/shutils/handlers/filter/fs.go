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
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"time"

	"github.com/spf13/afero"
)

type Predicate func(string) bool

type Fs struct {
	src  afero.Fs
	pred Predicate
}

// Chmod implements afero.Fs.
func (f *Fs) Chmod(name string, mode fs.FileMode) error {
	if err := f.matchesName(name); err != nil {
		return err
	}

	return f.src.Chmod(name, mode)
}

// Chown implements afero.Fs.
func (f *Fs) Chown(name string, uid, gid int) error {
	if err := f.matchesName(name); err != nil {
		return err
	}

	return f.src.Chown(name, uid, gid)
}

// Chtimes implements afero.Fs.
func (f *Fs) Chtimes(name string, atime, mtime time.Time) error {
	if err := f.matchesName(name); err != nil {
		return err
	}

	return f.src.Chtimes(name, atime, mtime)
}

// Create implements afero.Fs.
func (f *Fs) Create(name string) (afero.File, error) {
	if err := f.matchesName(name); err != nil {
		return nil, err
	}

	return f.src.Create(name)
}

// Mkdir implements afero.Fs.
func (f *Fs) Mkdir(name string, perm fs.FileMode) error {
	if err := f.matchesName(name); err != nil {
		return err
	}

	return f.src.Mkdir(name, perm)
}

// MkdirAll implements afero.Fs.
func (f *Fs) MkdirAll(path string, perm fs.FileMode) error {
	return f.src.Mkdir(path, perm)
}

// Name implements afero.Fs.
func (f *Fs) Name() string {
	return fmt.Sprintf("Filter: %s", f.src.Name())
}

// Open implements afero.Fs.
func (f *Fs) Open(name string) (afero.File, error) {
	if err := f.matchesName(name); err != nil {
		return nil, err
	}

	file, err := f.src.Open(name)
	if err != nil {
		return nil, err
	}

	return &File{
		file: file,
		pred: f.pred,
	}, nil
}

// OpenFile implements afero.Fs.
func (f *Fs) OpenFile(name string, flag int, perm fs.FileMode) (afero.File, error) {
	if err := f.matchesName(name); err != nil {
		return nil, err
	}

	return f.src.OpenFile(name, flag, perm)
}

// Remove implements afero.Fs.
func (f *Fs) Remove(name string) error {
	if err := f.matchesName(name); err != nil {
		return err
	}
	return f.src.Remove(name)
}

// RemoveAll implements afero.Fs.
func (f *Fs) RemoveAll(path string) error {
	if err := f.matchesName(path); err != nil {
		return err
	}

	return f.src.RemoveAll(path)
}

// Rename implements afero.Fs.
func (f *Fs) Rename(oldname, newname string) error {
	if err := f.matchesName(oldname); err != nil {
		return err
	}
	if err := f.matchesName(newname); err != nil {
		return err
	}

	return f.src.Rename(oldname, newname)
}

// LstatIfPossible implements afero.Fs.
func (f *Fs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if err := f.matchesName(name); err != nil {
		return nil, false, err
	}
	if lsf, ok := f.src.(afero.Lstater); ok {
		return lsf.LstatIfPossible(name)
	}
	fi, err := f.Stat(name)
	return fi, false, err
}

// Stat implements afero.Fs.
func (f *Fs) Stat(name string) (fs.FileInfo, error) {
	if err := f.matchesName(name); err != nil {
		return nil, err
	}

	return f.src.Stat(name)
}

func (f *Fs) matchesName(name string) error {
	if f.pred == nil || f.pred(name) {
		return nil
	} else {
		return syscall.ENOENT
	}
}

func NewFs(src afero.Fs, predicate Predicate) afero.Fs {
	return &Fs{src: src, pred: predicate}
}
