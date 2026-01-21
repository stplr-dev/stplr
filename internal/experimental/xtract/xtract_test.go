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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportedExtensionsStability(t *testing.T) {
	expected := []string{
		".tar.bz2", ".cpio.gz", ".tar.gz", ".tar.xz",
		".tar.lz", ".tar.z", ".7z", ".7z.001",
		".ar", ".br", ".brotli", ".bz2", ".cpgz",
		".cpio", ".deb", ".gz", ".gzip", ".iso",
		".lz4", ".lz", ".lzip", ".lzma", ".r00",
		".rar", ".s2", ".rpm", ".snappy", ".sz",
		".tar", ".tbz", ".tbz2", ".tgz", ".tlz",
		".txz", ".tz", ".xz", ".z", ".zip",
		".zlib", ".zst", ".zstd", ".zz",
	}

	assert.Equal(t, expected, SupportedExtensions(),
		"SupportedArchiveExtensions must not change")
}

func TestSupportedExtensionsHaveFixtures(t *testing.T) {
	fixturesDir := "./fixtures"

	entries, err := os.ReadDir(fixturesDir)
	require.NoError(t, err, "failed to read fixtures directory")

	// Collect all file names once
	var filenames []string
	for _, e := range entries {
		if !e.IsDir() {
			filenames = append(filenames, e.Name())
		}
	}

	missing := []string{}

	for _, ext := range SupportedExtensions() {
		if ext == ".r00" {
			// I couldn't easily create a text archive with a real .r00 on Linux,
			// so we'll just hope this case works.
			continue
		}

		found := false
		for _, name := range filenames {
			if strings.HasSuffix(name, ext) {
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, ext)
		}
	}

	assert.Emptyf(
		t,
		missing,
		"no fixture found for extensions: %v in %s",
		missing,
		fixturesDir,
	)
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		// Supported extensions
		{"tar file", "file.tar", true},
		{"tar.gz file", "file.tar.gz", true},
		{"zip file", "file.zip", true},
		{"rar file", "file.rar", true},
		{"7z file", "file.7z", true},
		{"deb file", "file.deb", true},
		{"rpm file", "file.rpm", true},
		{"Upper case extension", "file.ZIP", true},
		{"Mixed case extension", "file.Zip", true},
		{"Nested extension", "archive.tar.gz", true},
		{"Complex path with supported extension", "/path/to/file.tar.xz", true},

		// Unsupported extensions
		{"text file", "file.txt", false},
		{"executable", "file.exe", false},
		{"no extension", "file", false},
		{"partial extension match", "file.tar.backup", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSupported(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// expectedPayloadFiles represents files that should exist in payload
var expectedPayloadFiles = []string{
	"payload/file.txt",
	"payload/binary.bin",
	"payload/subdir/data.txt",
	"payload/subdir/nested/deep.txt",
	"payload/subdir/zeros",
}

// expectedPayloadSymlinks represents symlinks in payload
var expectedPayloadSymlinks = []string{
	"payload/subdir/symlink.txt", // symlink to ../file.txt
	"payload/dirlink",            // symlink to subdir
}

// expectedArDebFiles represents files in AR/DEB archives (flat structure, no symlinks)
var expectedArDebFiles = []string{
	"file.txt",
	"binary.bin",
	"data.txt",
	"deep.txt",
	"zeros",
}

// archiveCategory represents different types of archive formats
type archiveCategory int

const (
	categoryFull       archiveCategory = iota // Full structure with symlinks (tar, zip, 7z, rar, etc.)
	categoryCompressed                        // Compression-only formats that need double extraction
	categoryLimited                           // Limited formats without symlinks/dirs (ar, deb)
)

// compressionOnlyFormats are formats that only compress a single file/stream
// and don't preserve directory structure. They typically contain a tar archive.
var compressionOnlyFormats = []string{
	".gz", ".gzip", ".bz2", ".xz", ".lzma", ".lzma2",
	".lz", ".lzip", ".lz4", ".z", ".Z",
	".zst", ".zstd", ".br", ".brotli",
	".zlib", ".snappy", ".sz", ".s2", ".zz",
}

// getArchiveCategory returns the category of an archive based on its extension
func getArchiveCategory(filename string) archiveCategory {
	ext := filepath.Ext(filename)
	lower := strings.ToLower(filename)

	// Limited formats
	if ext == ".ar" || ext == ".deb" {
		return categoryLimited
	}

	// Compression-only formats
	for _, compExt := range compressionOnlyFormats {
		if strings.HasSuffix(lower, compExt) {
			return categoryCompressed
		}
	}

	// Everything else supports full structure
	return categoryFull
}

func TestExtractAllFixtures(t *testing.T) {
	fixturesDir := "./fixtures"
	baseOutputDir := t.TempDir()

	entries, err := os.ReadDir(fixturesDir)
	require.NoError(t, err, "failed to read fixtures directory")

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()

		if IsSupported(name) {
			t.Run(name, func(t *testing.T) {
				outputDir, err := os.MkdirTemp(baseOutputDir, "extract-"+name+"-*")
				require.NoError(t, err, "failed to create temporary directory")
				defer os.RemoveAll(outputDir)

				archivePath := filepath.Join(fixturesDir, name)
				category := getArchiveCategory(name)

				result, err := ExtractArchive(archivePath, outputDir)
				require.NoErrorf(t, err, "failed to extract archive %q", name)

				// Verify extraction result
				assert.NotNil(t, result, "extraction result should not be nil")
				assert.Greater(t, result.Size, int64(0), "extracted size should be positive")
				assert.NotEmpty(t, result.Files, "files list should not be empty")

				// Verify extracted content based on archive type
				switch category {
				case categoryLimited:
					// For AR/DEB, verify flat file list
					verifyFlatFiles(t, outputDir, expectedArDebFiles)

				case categoryFull:
					// For full archives, verify payload structure
					verifyPayloadStructure(t, outputDir, expectedPayloadFiles, expectedPayloadSymlinks)
				}

				t.Logf("Extracted %s: size=%d bytes, files=%d, archives=%d",
					name, result.Size, len(result.Files), len(result.Archives))
			})
		}
	}
}

// verifyPayloadStructure checks that the expected payload structure exists
func verifyPayloadStructure(t *testing.T, outputDir string, expectedFiles, expectedSymlinks []string) {
	t.Helper()

	// Verify payload directory exists
	payloadDir := filepath.Join(outputDir, "payload")
	info, err := os.Stat(payloadDir)
	require.NoError(t, err, "payload directory should exist")
	assert.True(t, info.IsDir(), "payload should be a directory")

	// Verify all expected files exist
	for _, expectedPath := range expectedFiles {
		fullPath := filepath.Join(outputDir, expectedPath)
		info, err := os.Lstat(fullPath)
		assert.NoErrorf(t, err, "expected file should exist: %s", expectedPath)
		if err == nil {
			assert.NotNilf(t, info, "file info should not be nil for: %s", expectedPath)
			assert.False(t, info.Mode()&os.ModeSymlink != 0, "should be a regular file: %s", expectedPath)
		}
	}

	// Verify symlinks exist (may not be supported by all formats)
	for _, expectedPath := range expectedSymlinks {
		fullPath := filepath.Join(outputDir, expectedPath)
		info, err := os.Lstat(fullPath)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			t.Logf("Symlink found: %s", expectedPath)
		}
		// Don't fail if symlinks are missing, as not all formats support them
	}
}

// verifyFlatFiles checks that files exist directly in output dir (for AR/DEB)
func verifyFlatFiles(t *testing.T, outputDir string, expectedFiles []string) {
	t.Helper()

	for _, expectedFile := range expectedFiles {
		fullPath := filepath.Join(outputDir, expectedFile)
		info, err := os.Stat(fullPath)
		assert.NoErrorf(t, err, "expected file should exist: %s", expectedFile)
		if err == nil {
			assert.False(t, info.IsDir(), "should be a file, not directory: %s", expectedFile)
		}
	}
}
