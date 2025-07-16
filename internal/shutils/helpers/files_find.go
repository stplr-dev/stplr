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
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func matchNamePattern(name, pattern string) bool {
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

func validateDir(dirPath, commandName string) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("%s: %w", commandName, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %s is not a directory", commandName, dirPath)
	}
	return nil
}

func outputFiles(hc interp.HandlerContext, files []string) error {
	for _, file := range files {
		v, err := syntax.Quote(file, syntax.LangAuto)
		if err != nil {
			return err
		}
		fmt.Fprintln(hc.Stdout, v)
	}
	return nil
}

func makeRelativePath(basePath, fullPath string) (string, error) {
	relPath, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return "", err
	}
	return "./" + relPath, nil
}

func filesFindLangCmd(hc interp.HandlerContext, cmd string, args []string) error {
	namePattern := "*.mo"
	if len(args) > 0 {
		namePattern = args[0] + ".mo"
	}

	localePath := "./usr/share/locale/"
	realPath := path.Join(hc.Dir, localePath)

	if err := validateDir(realPath, "files-find-lang"); err != nil {
		return err
	}

	var langFiles []string
	err := filepath.Walk(realPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && matchNamePattern(info.Name(), namePattern) {
			relPath, relErr := makeRelativePath(hc.Dir, p)
			if relErr != nil {
				return relErr
			}
			langFiles = append(langFiles, relPath)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("files-find-lang: %w", err)
	}

	return outputFiles(hc, langFiles)
}

func filesFindDocCmd(hc interp.HandlerContext, cmd string, args []string) error {
	namePattern := "*"
	if len(args) > 0 {
		namePattern = args[0]
	}

	docPath := "./usr/share/doc/"
	docRealPath := path.Join(hc.Dir, docPath)

	if err := validateDir(docRealPath, "files-find-doc"); err != nil {
		return err
	}

	var docFiles []string

	entries, err := os.ReadDir(docRealPath)
	if err != nil {
		return fmt.Errorf("files-find-doc: %w", err)
	}

	for _, entry := range entries {
		if matchNamePattern(entry.Name(), namePattern) {
			targetPath := filepath.Join(docRealPath, entry.Name())
			targetInfo, err := os.Stat(targetPath)
			if err != nil {
				return fmt.Errorf("files-find-doc: %w", err)
			}
			if targetInfo.IsDir() {
				err := filepath.Walk(targetPath, func(subPath string, subInfo os.FileInfo, subErr error) error {
					if subErr != nil {
						return subErr
					}
					relPath, err := makeRelativePath(hc.Dir, subPath)
					if err != nil {
						return err
					}
					docFiles = append(docFiles, relPath)
					return nil
				})
				if err != nil {
					return fmt.Errorf("files-find-doc: %w", err)
				}
			}
		}
	}

	return outputFiles(hc, docFiles)
}

func filesFindCmd(hc interp.HandlerContext, cmd string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("files-find: at least one glob pattern is required")
	}

	var foundFiles []string

	for _, globPattern := range args {
		searchPath := path.Join(hc.Dir, globPattern)

		basepath, pattern := doublestar.SplitPattern(searchPath)
		fsys := NewDirLFS(basepath)
		matches, err := doublestar.Glob(fsys, pattern, doublestar.WithNoFollow(), doublestar.WithFailOnPatternNotExist())
		if err != nil {
			return fmt.Errorf("files-find: glob pattern error: %w", err)
		}

		for _, match := range matches {
			relPath, err := makeRelativePath(hc.Dir, path.Join(basepath, match))
			if err != nil {
				continue
			}
			foundFiles = append(foundFiles, relPath)
		}
	}

	return outputFiles(hc, foundFiles)
}
