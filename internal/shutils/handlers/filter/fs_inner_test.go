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
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// setupTestFs creates a MemMapFs with test files and directories and a filter.Fs with a predicate.
func setupTestFs(t *testing.T) (afero.Fs, afero.Fs) {
	t.Helper()
	memFs := afero.NewMemMapFs()
	predicate := func(p string) bool {
		name := filepath.Base(p)
		return name == "allowed.txt" || name == "allowed_dir"
	}
	filterFs := NewFs(memFs, predicate)

	// Create test files and directories
	if err := memFs.Mkdir("allowed_dir", 0o755); err != nil {
		t.Fatalf("Failed to create allowed_dir: %v", err)
	}
	if err := memFs.Mkdir("denied_dir", 0o755); err != nil {
		t.Fatalf("Failed to create denied_dir: %v", err)
	}
	if _, err := memFs.Create("allowed.txt"); err != nil {
		t.Fatalf("Failed to create allowed.txt: %v", err)
	}
	if _, err := memFs.Create("denied.txt"); err != nil {
		t.Fatalf("Failed to create denied.txt: %v", err)
	}
	return memFs, filterFs
}

func TestChmod(t *testing.T) {
	_, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		mode        fs.FileMode
		expectError bool
	}{
		{"allowed.txt", 0o644, false},
		{"denied.txt", 0o644, true},
		{"allowed_dir", 0o755, false},
		{"denied_dir", 0o755, true},
		{"nonexistent.txt", 0o644, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filterFs.Chmod(tt.name, tt.mode)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
			}
		})
	}
}

func TestChown(t *testing.T) {
	_, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		uid, gid    int
		expectError bool
	}{
		{"allowed.txt", 1000, 1000, false},
		{"denied.txt", 1000, 1000, true},
		{"allowed_dir", 1000, 1000, false},
		{"denied_dir", 1000, 1000, true},
		{"nonexistent.txt", 1000, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filterFs.Chown(tt.name, tt.uid, tt.gid)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
			}
		})
	}
}

func TestChtimes(t *testing.T) {
	_, filterFs := setupTestFs(t)
	now := time.Now()
	tests := []struct {
		name         string
		atime, mtime time.Time
		expectError  bool
	}{
		{"allowed.txt", now, now, false},
		{"denied.txt", now, now, true},
		{"allowed_dir", now, now, false},
		{"denied_dir", now, now, true},
		{"nonexistent.txt", now, now, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filterFs.Chtimes(tt.name, tt.atime, tt.mtime)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	_, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		expectError bool
	}{
		{"allowed.txt", false},
		{"denied.txt", true},
		{"allowed_dir/new.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := filterFs.Create(tt.name)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
				assert.Nil(t, file, "Expected nil file for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
				assert.NotNil(t, file, "Expected non-nil file for path: %s", tt.name)
			}
		})
	}
}

func TestMkdir(t *testing.T) {
	memFs, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		perm        fs.FileMode
		expectError bool
	}{
		{"allowed_dir", 0o755, false},
		{"denied_dir", 0o755, true},
		{"nonexistent_dir", 0o755, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = memFs.RemoveAll(tt.name)
			err := filterFs.Mkdir(tt.name, tt.perm)
			if tt.expectError {
				assert.Error(t, err, "Expected error for dir: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for dir: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for dir: %s", tt.name)
			}
		})
	}
}

func TestMkdirAll(t *testing.T) {
	_, filterFs := setupTestFs(t)
	err := filterFs.MkdirAll("allowed_dir/subdir", 0o755)
	assert.NoError(t, err, "Unexpected error for MkdirAll")
}

func TestName(t *testing.T) {
	_, filterFs := setupTestFs(t)
	assert.Equal(t, "Filter: MemMapFS", filterFs.Name(), "Unexpected Name output")
}

func TestOpen(t *testing.T) {
	_, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		expectError bool
	}{
		{"allowed.txt", false},
		{"denied.txt", true},
		{"allowed_dir", false},
		{"denied_dir", true},
		{"nonexistent.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := filterFs.Open(tt.name)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
				assert.Nil(t, file, "Expected nil file for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
				assert.NotNil(t, file, "Expected non-nil file for path: %s", tt.name)
			}
		})
	}
}

func TestOpenFile(t *testing.T) {
	_, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		flag        int
		perm        fs.FileMode
		expectError bool
	}{
		{"allowed.txt", os.O_RDWR, 0o644, false},
		{"denied.txt", os.O_RDWR, 0o644, true},
		{"allowed_dir", os.O_RDWR, 0o644, false},
		{"denied_dir", os.O_RDWR, 0o644, true},
		{"nonexistent.txt", os.O_RDWR, 0o644, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := filterFs.OpenFile(tt.name, tt.flag, tt.perm)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
				assert.Nil(t, file, "Expected nil file for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
				assert.NotNil(t, file, "Expected non-nil file for path: %s", tt.name)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	memFs, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		expectError bool
	}{
		{"allowed.txt", false},
		{"denied.txt", true},
		{"allowed_dir", false},
		{"denied_dir", true},
		{"nonexistent.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "allowed_dir" || tt.name == "denied_dir" {
				_ = memFs.RemoveAll(tt.name)
				_ = memFs.Mkdir(tt.name, 0o755)
			} else if tt.name != "nonexistent.txt" {
				_, _ = memFs.Create(tt.name)
			}
			err := filterFs.Remove(tt.name)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
			}
		})
	}
}

func TestRemoveAll(t *testing.T) {
	memFs, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		expectError bool
	}{
		{"allowed.txt", false},
		{"denied.txt", true},
		{"allowed_dir", false},
		{"denied_dir", true},
		{"nonexistent.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "allowed_dir" || tt.name == "denied_dir" {
				_ = memFs.RemoveAll(tt.name)
				_ = memFs.Mkdir(tt.name, 0o755)
			} else if tt.name != "nonexistent.txt" {
				_, _ = memFs.Create(tt.name)
			}
			err := filterFs.RemoveAll(tt.name)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
			}
		})
	}
}

func TestRename(t *testing.T) {
	memFs, filterFs := setupTestFs(t)
	tests := []struct {
		oldname, newname string
		expectError      bool
	}{
		{"allowed.txt", "allowed_dir", false},
		{"allowed.txt", "denied.txt", true},
		{"denied.txt", "allowed.txt", true},
		{"nonexistent.txt", "allowed.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.oldname+"->"+tt.newname, func(t *testing.T) {
			if tt.oldname != "nonexistent.txt" {
				_, _ = memFs.Create(tt.oldname)
			}
			err := filterFs.Rename(tt.oldname, tt.newname)
			if tt.expectError {
				assert.Error(t, err, "Expected error for rename: %s -> %s", tt.oldname, tt.newname)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for rename: %s -> %s", tt.oldname, tt.newname)
			} else {
				assert.NoError(t, err, "Unexpected error for rename: %s -> %s", tt.oldname, tt.newname)
			}
		})
	}
}

func TestLstatIfPossible(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	osFs := afero.NewOsFs()
	predicate := func(p string) bool {
		name := filepath.Base(p)
		return name == "allowed.txt" || name == "allowed_dir"
	}
	filterFs := NewFs(osFs, predicate)

	assert.NoError(t, os.Mkdir(filepath.Join(tempDir, "allowed_dir"), 0o755))
	assert.NoError(t, os.Mkdir(filepath.Join(tempDir, "denied_dir"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(tempDir, "allowed.txt"), []byte{}, 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(tempDir, "denied.txt"), []byte{}, 0o644))

	tests := []struct {
		name        string
		expectError bool
	}{
		{"allowed.txt", false},
		{"denied.txt", true},
		{"allowed_dir", false},
		{"denied_dir", true},
		{"nonexistent.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tempDir, tt.name)
			lst, ok := filterFs.(afero.Lstater)
			assert.True(t, ok, "Expected Lstater interface for filterFs")
			fi, isLstat, err := lst.LstatIfPossible(path)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", path)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", path)
				assert.Nil(t, fi, "Expected nil FileInfo for path: %s", path)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", path)
				assert.NotNil(t, fi, "Expected non-nil FileInfo for path: %s", path)
				assert.True(t, isLstat, "Expected Lstat to be used for path: %s", path)
			}
		})
	}
}

func TestStat(t *testing.T) {
	_, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		expectError bool
	}{
		{"allowed.txt", false},
		{"denied.txt", true},
		{"allowed_dir", false},
		{"denied_dir", true},
		{"nonexistent.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fi, err := filterFs.Stat(tt.name)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
				assert.Nil(t, fi, "Expected nil FileInfo for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
				assert.NotNil(t, fi, "Expected non-nil FileInfo for path: %s", tt.name)
			}
		})
	}
}

func TestNilPredicate(t *testing.T) {
	memFs := afero.NewMemMapFs()
	nilPredFs := NewFs(memFs, nil)
	_, err := nilPredFs.Create("anyfile.txt")
	assert.NoError(t, err, "Expected no error with nil predicate")
	err = nilPredFs.Mkdir("anydir", 0o755)
	assert.NoError(t, err, "Expected no error with nil predicate for directory")
}

func TestMatchesName(t *testing.T) {
	_, filterFs := setupTestFs(t)
	tests := []struct {
		name        string
		expectError bool
	}{
		{"allowed.txt", false},
		{"denied.txt", true},
		{"allowed_dir", false},
		{"denied_dir", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filterFs.(*Fs).matchesName(tt.name)
			if tt.expectError {
				assert.Error(t, err, "Expected error for path: %s", tt.name)
				assert.Equal(t, syscall.ENOENT, err, "Expected ENOENT error for path: %s", tt.name)
			} else {
				assert.NoError(t, err, "Unexpected error for path: %s", tt.name)
			}
		})
	}
}
