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

package dl

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// If the checksum does not match, returns ErrChecksumMismatch
func VerifyHashFromLocal(path string, opts Options) error {
	if opts.Hash != nil {
		h, err := opts.NewHash()
		if err != nil {
			return err
		}

		err = HashLocal(filepath.Join(opts.Destination, path), h)
		if err != nil {
			return err
		}

		sum := h.Sum(nil)

		slog.Debug("validate checksum", "real", hex.EncodeToString(sum), "expected", hex.EncodeToString(opts.Hash))

		if !bytes.Equal(sum, opts.Hash) {
			return ErrChecksumMismatch
		}
	}

	return nil
}

// hashFile hashes a single file using the provided hash function.
func hashFile(path string, h hash.Hash) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(h, f)
	return err
}

// hashDir walks a directory and hashes all regular files, skipping ".git" directories.
func hashDir(path string, h hash.Hash) error {
	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return hashFile(p, h)
	})
}

// HashLocal hashes a file or directory using the provided hash function.
func HashLocal(path string, h hash.Hash) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().IsRegular() {
		return hashFile(path, h)
	}
	if info.IsDir() {
		return hashDir(path, h)
	}
	return fmt.Errorf("unsupported file type: %s", path)
}
