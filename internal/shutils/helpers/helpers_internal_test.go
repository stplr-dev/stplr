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
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type symlink struct {
	linkPath   string
	targetPath string
}

type testCase struct {
	name             string
	dirsToCreate     []string
	filesToCreate    []string
	expectedOutput   []string
	symlinksToCreate []symlink
	args             string
	expectedError    error
}

func testFilesFindCommon(t *testing.T, command string, tc testCase) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", fmt.Sprintf("test-%s", command))
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		for _, dir := range tc.dirsToCreate {
			dirPath := filepath.Join(tempDir, dir)
			err := os.MkdirAll(dirPath, 0o755)
			assert.NoError(t, err)
		}

		for _, file := range tc.filesToCreate {
			filePath := filepath.Join(tempDir, file)
			err := os.WriteFile(filePath, []byte("test content"), 0o644)
			assert.NoError(t, err)
		}

		for _, sl := range tc.symlinksToCreate {
			linkFullPath := filepath.Join(tempDir, sl.linkPath)
			targetFullPath := sl.targetPath

			// make sure parent dir exists
			err := os.MkdirAll(filepath.Dir(linkFullPath), 0o755)
			assert.NoError(t, err)

			err = os.Symlink(targetFullPath, linkFullPath)
			assert.NoError(t, err)
		}

		buf := &bytes.Buffer{}
		runner, err := interp.New(
			interp.Dir(tempDir),
			interp.StdIO(os.Stdin, buf, os.Stderr),
			interp.ExecHandlers(func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
				return Helpers.ExecHandler(interp.DefaultExecHandler(1000))
			}),
		)
		assert.NoError(t, err)

		scriptContent := command + " " + tc.args

		script, err := syntax.NewParser().Parse(strings.NewReader(scriptContent), "")
		assert.NoError(t, err)

		err = runner.Run(context.Background(), script)
		if tc.expectedError != nil {
			assert.ErrorAs(t, err, &tc.expectedError)
		} else {
			assert.NoError(t, err)
		}

		contents, err := shlex.Split(buf.String())
		assert.NoError(t, err)
		assert.ElementsMatch(t, tc.expectedOutput, contents)
	})
}

func TestFindFilesDoc(t *testing.T) {
	tests := []testCase{
		{
			name: "All dirs",
			dirsToCreate: []string{
				"usr/share/doc/yandex-browser-stable/subdir",
				"usr/share/doc/firefox",
			},
			filesToCreate: []string{
				"usr/share/doc/yandex-browser-stable/README.md",
				"usr/share/doc/yandex-browser-stable/subdir/nested-file.txt",
				"usr/share/doc/firefox/README.md",
			},
			expectedOutput: []string{
				"./usr/share/doc/yandex-browser-stable",
				"./usr/share/doc/yandex-browser-stable/README.md",
				"./usr/share/doc/yandex-browser-stable/subdir",
				"./usr/share/doc/yandex-browser-stable/subdir/nested-file.txt",
				"./usr/share/doc/firefox",
				"./usr/share/doc/firefox/README.md",
			},
			args: "",
		},
		{
			name: "Only selected dir",
			dirsToCreate: []string{
				"usr/share/doc/yandex-browser-stable/subdir",
				"usr/share/doc/firefox",
				"usr/share/doc/foo/yandex-browser-stable",
			},
			filesToCreate: []string{
				"usr/share/doc/yandex-browser-stable/README.md",
				"usr/share/doc/yandex-browser-stable/subdir/nested-file.txt",
				"usr/share/doc/firefox/README.md",
				"usr/share/doc/firefox/yandex-browser-stable",
				"usr/share/doc/foo/yandex-browser-stable/README.md",
			},
			expectedOutput: []string{
				"./usr/share/doc/yandex-browser-stable",
				"./usr/share/doc/yandex-browser-stable/README.md",
				"./usr/share/doc/yandex-browser-stable/subdir",
				"./usr/share/doc/yandex-browser-stable/subdir/nested-file.txt",
			},
			args: "yandex-browser-stable",
		},
	}

	for _, tc := range tests {
		testFilesFindCommon(t, "files-find-doc", tc)
	}
}

func TestFindLang(t *testing.T) {
	tests := []testCase{
		{
			name: "All dirs",
			dirsToCreate: []string{
				"usr/share/locale/ru/LC_MESSAGES",
				"usr/share/locale/tr/LC_MESSAGES",
			},
			filesToCreate: []string{
				"usr/share/locale/ru/LC_MESSAGES/yandex-disk.mo",
				"usr/share/locale/ru/LC_MESSAGES/yandex-disk-indicator.mo",
				"usr/share/locale/tr/LC_MESSAGES/yandex-disk.mo",
			},
			expectedOutput: []string{
				"./usr/share/locale/ru/LC_MESSAGES/yandex-disk.mo",
				"./usr/share/locale/ru/LC_MESSAGES/yandex-disk-indicator.mo",
				"./usr/share/locale/tr/LC_MESSAGES/yandex-disk.mo",
			},
			args: "",
		},
		{
			name: "All dirs",
			dirsToCreate: []string{
				"usr/share/locale/ru/LC_MESSAGES",
				"usr/share/locale/tr/LC_MESSAGES",
			},
			filesToCreate: []string{
				"usr/share/locale/ru/LC_MESSAGES/yandex-disk.mo",
				"usr/share/locale/ru/LC_MESSAGES/yandex-disk-indicator.mo",
				"usr/share/locale/tr/LC_MESSAGES/yandex-disk.mo",
			},
			expectedOutput: []string{
				"./usr/share/locale/ru/LC_MESSAGES/yandex-disk.mo",
				"./usr/share/locale/tr/LC_MESSAGES/yandex-disk.mo",
			},
			args: "yandex-disk",
		},
	}

	for _, tc := range tests {
		testFilesFindCommon(t, "files-find-lang", tc)
	}
}

func TestFindFiles(t *testing.T) {
	tests := []testCase{
		{
			name: "With file and dir symlinks",
			dirsToCreate: []string{
				"usr/share/locale/ru/LC_MESSAGES",
				"usr/share/locale/tr/LC_MESSAGES",
				"opt/app",
				"opt/app/internal",
				"opt/app/with space",
				"usr/bin",
			},
			filesToCreate: []string{
				"usr/share/locale/ru/LC_MESSAGES/yandex-disk.mo",
				"usr/share/locale/ru/LC_MESSAGES/yandex-disk-indicator.mo",
				"usr/share/locale/tr/LC_MESSAGES/yandex-disk.mo",
				"opt/app/internal/test",
				"opt/app/with space/file",
			},
			symlinksToCreate: []symlink{
				{
					linkPath:   "/opt/app/etc",
					targetPath: "/etc",
				},
				{
					linkPath:   "/usr/bin/file",
					targetPath: "/not-existing",
				},
			},
			expectedOutput: []string{
				"./usr/share/locale/ru/LC_MESSAGES/yandex-disk.mo",
				"./usr/share/locale/ru/LC_MESSAGES/yandex-disk-indicator.mo",
				"./usr/share/locale/tr/LC_MESSAGES/yandex-disk.mo",
				"./opt/app/etc",
				"./opt/app/internal",
				"./opt/app/internal/test",
				"./opt/app/with space",
				"./opt/app/with space/file",
				"./usr/bin/file",
			},
			args:          "\"/usr/share/locale/*/LC_MESSAGES/*.mo\" \"/opt/app/**/*\" \"/usr/bin/file\"",
			expectedError: nil,
		},
		{
			name:          "Not existing paths should throw error",
			args:          "\"/opt/test/not-existing\"",
			expectedError: doublestar.ErrPatternNotExist,
		},
	}

	for _, tc := range tests {
		testFilesFindCommon(t, "files-find", tc)
	}
}

func TestFindFilesPrefix(t *testing.T) {
	tests := []testCase{
		{
			name: "With file and dir symlinks",
			dirsToCreate: []string{
				"usr/bin",
			},
			filesToCreate: []string{
				"usr/bin/foo",
			},
			symlinksToCreate: []symlink{
				{
					linkPath:   "/usr/bin/file",
					targetPath: "/not-existing",
				},
			},
			expectedOutput: []string{
				"./usr/bin/file",
				"./usr/bin/foo",
			},
			args:          "\"f*\"",
			expectedError: nil,
		},
		{
			name:          "Not existing paths should throw error",
			args:          "\"/opt/test/not-existing\"",
			expectedError: doublestar.ErrPatternNotExist,
		},
	}

	for _, tc := range tests {
		testFilesFindCommon(t, "files-find-binary", tc)
	}
}

func TestFindFilesLicense(t *testing.T) {
	tests := []testCase{
		{
			name: "All dirs",
			dirsToCreate: []string{
				"./usr/share/licenses/yandex-disk",
				"./usr/share/licenses/chrome",
				"./usr/share/licenses/vk",
			},
			filesToCreate: []string{
				"./usr/share/licenses/yandex-disk/LICENSE",
				"./usr/share/licenses/chrome/LICENSE",
				"./usr/share/licenses/vk/LICENSE",
			},
			expectedOutput: []string{
				"./usr/share/licenses/yandex-disk",
				"./usr/share/licenses/yandex-disk/LICENSE",
				"./usr/share/licenses/chrome",
				"./usr/share/licenses/chrome/LICENSE",
				"./usr/share/licenses/vk",
				"./usr/share/licenses/vk/LICENSE",
			},
			args: "",
		},
		{
			name: "Only one",
			dirsToCreate: []string{
				"./usr/share/licenses/yandex-disk",
				"./usr/share/licenses/chrome",
			},
			filesToCreate: []string{
				"./usr/share/licenses/yandex-disk/LICENSE",
				"./usr/share/licenses/chrome/LICENSE",
			},
			expectedOutput: []string{
				"./usr/share/licenses/yandex-disk",
				"./usr/share/licenses/yandex-disk/LICENSE",
			},
			args: "yandex-disk",
		},
		{
			name: "Multiple",
			dirsToCreate: []string{
				"./usr/share/licenses/yandex-disk",
				"./usr/share/licenses/chrome",
				"./usr/share/licenses/vk",
			},
			filesToCreate: []string{
				"./usr/share/licenses/yandex-disk/LICENSE",
				"./usr/share/licenses/chrome/LICENSE",
				"./usr/share/licenses/vk/LICENSE",
			},
			expectedOutput: []string{
				"./usr/share/licenses/yandex-disk",
				"./usr/share/licenses/yandex-disk/LICENSE",
				"./usr/share/licenses/vk",
				"./usr/share/licenses/vk/LICENSE",
			},
			args: "vk yandex-disk",
		},
	}

	for _, tc := range tests {
		testFilesFindCommon(t, "files-find-license", tc)
	}
}

func TestFindFilesBinary(t *testing.T) {
	tests := []testCase{
		{
			name: "All dirs",
			dirsToCreate: []string{
				"./usr/bin",
			},
			filesToCreate: []string{
				"./usr/bin/yandex-disk",
				"./usr/bin/chrome",
				"./usr/bin/vk",
			},
			expectedOutput: []string{
				"./usr/bin/yandex-disk",
				"./usr/bin/chrome",
				"./usr/bin/vk",
			},
			args: "",
		},
		{
			name: "Only one",
			dirsToCreate: []string{
				"./usr/bin",
			},
			filesToCreate: []string{
				"./usr/bin/yandex-disk",
				"./usr/bin/chrome",
				"./usr/bin/vk",
			},
			expectedOutput: []string{
				"./usr/bin/yandex-disk",
			},
			args: "yandex-disk",
		},
		{
			name: "Multiple",
			dirsToCreate: []string{
				"./usr/bin",
			},
			filesToCreate: []string{
				"./usr/bin/yandex-disk",
				"./usr/bin/chrome",
				"./usr/bin/vk",
			},
			expectedOutput: []string{
				"./usr/bin/yandex-disk",
				"./usr/bin/vk",
			},
			args: "vk yandex-disk",
		},
	}

	for _, tc := range tests {
		testFilesFindCommon(t, "files-find-binary", tc)
	}
}
