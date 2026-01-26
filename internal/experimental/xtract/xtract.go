// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package xtract

import (
	"fmt"
	"strings"

	"golift.io/xtractr"
)

func SupportedExtensions() []string {
	return xtractr.SupportedExtensions()
}

// IsSupported checks if the given filename has a supported archive extension.
// The check is case-insensitive.
func IsSupported(filename string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range SupportedExtensions() {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// ExtractResult contains information about the extraction process.
type ExtractResult struct {
	Size     int64    // Total size of extracted data
	Files    []string // List of extracted files
	Archives []string // List of archives processed (including nested)
}

// ExtractArchive extracts an archive file using the xtractr library.
func ExtractArchive(filepath, outputDir string) (*ExtractResult, error) {
	x := &xtractr.XFile{
		FilePath:  filepath,
		OutputDir: outputDir,
		FileMode:  0o644,
		DirMode:   0o755,
	}

	size, files, archives, err := xtractr.ExtractFile(x)
	if err != nil {
		return nil, fmt.Errorf("xtractr.ExtractFile: %w", err)
	}

	return &ExtractResult{
		Size:     size,
		Files:    files,
		Archives: archives,
	}, nil
}
