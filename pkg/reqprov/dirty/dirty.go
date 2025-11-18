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

package dirty

import (
	"bytes"
	"context"
	"debug/elf"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/goreleaser/nfpm/v2"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/pkg/types"
)

type Dirty struct{}

func New() *Dirty {
	return &Dirty{}
}

func (o *Dirty) FindProvides(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	out.Info(gotext.Get("AutoProv is not implemented for this package format, so it's skipped"))
	return nil
}

func (o *Dirty) FindRequires(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	provSet := make(map[string]struct{})
	needSet := make(map[string]struct{})

	paths, err := getPaths(pkgInfo, dirs, skiplist)
	if err != nil {
		return err
	}

	for _, p := range paths {
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		err = o.processFile(p, info, provSet, needSet)
		if err != nil {
			return err
		}
	}

	extern := diffSets(needSet, provSet)
	o.updatePackageDependencies(pkgInfo, extern, filter)

	return nil
}

func (o *Dirty) BuildDepends(ctx context.Context) ([]string, error) { return []string{}, nil }

func (o *Dirty) processFile(path string, fi os.FileInfo, provSet, needSet map[string]struct{}) error {
	if fi.IsDir() || !isELF(path) {
		return nil
	}

	execOrSO := (fi.Mode()&0o111 != 0) || looksLikeLib(path)
	if !execOrSO {
		return nil
	}

	return processELF(path, provSet, needSet)
}

func (o *Dirty) updatePackageDependencies(pkgInfo *nfpm.Info, extern map[string]struct{}, filter []string) {
	regexFilters := makeRegexList(filter)
	existing := ToSet(pkgInfo.Depends)
	newDeps := make([]string, 0, len(extern))
	for dep := range extern {
		if matchesRegexList(dep, regexFilters) {
			continue
		}
		if _, ok := existing[dep]; ok {
			continue
		}
		existing[dep] = struct{}{}
		newDeps = append(newDeps, dep)
	}
	pkgInfo.Depends = append(pkgInfo.Depends, newDeps...)
}

func relWithSlash(basepath, targpath string) (string, error) {
	p, err := filepath.Rel(basepath, targpath)
	if err != nil {
		return "", err
	}
	return filepath.Join("/", p), nil
}

func matchesAnyPattern(p string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		if matched, err := doublestar.PathMatch(pattern, p); err != nil {
			return false, err
		} else if matched {
			return true, nil
		}
	}
	return false, nil
}

func getPaths(pkgInfo *nfpm.Info, dirs types.Directories, skiplist []string) ([]string, error) {
	var paths []string

	for _, content := range pkgInfo.Contents {
		dest := content.Destination
		p, err := relWithSlash(dirs.PkgDir, dest)
		if err != nil {
			return paths, err
		}

		if matched, err := matchesAnyPattern(p, skiplist); err != nil {
			return paths, err
		} else if !matched {
			paths = append(paths, path.Join(dirs.PkgDir, dest))
		}
	}

	return paths, nil
}

func isELF(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 4)
	if _, err := f.Read(buf); err != nil {
		return false
	}
	return bytes.Equal(buf, []byte{0x7f, 'E', 'L', 'F'})
}

func processELF(path string, provSet, needSet map[string]struct{}) error {
	f, err := elf.Open(path)
	if err != nil {
		return nil // not ELF or unreadable – we think we're skipping
	}
	defer f.Close()

	if needed, err := f.DynString(elf.DT_NEEDED); err == nil {
		for _, n := range needed {
			if dep, err := formatDependencyString(f, n); err == nil {
				needSet[dep] = struct{}{}
			}
		}
	}

	if sonames, err := f.DynString(elf.DT_SONAME); err == nil {
		for _, s := range sonames {
			if dep, err := formatDependencyString(f, s); err == nil {
				provSet[dep] = struct{}{}
			}
		}
	}

	if dep, err := formatDependencyString(f, filepath.Base(path)); err == nil {
		provSet[dep] = struct{}{}
	}

	return nil
}

func formatDependencyString(f *elf.File, dep string) (string, error) {
	const (
		suffix32bit = ""
		suffix64bit = "()(64bit)"
	)

	switch f.Class {
	case elf.ELFCLASS32:
		return dep + suffix32bit, nil
	case elf.ELFCLASS64:
		return dep + suffix64bit, nil
	default:
		return "", errors.New("unknown ELF class")
	}
}

func bracketValue(s string) string {
	i := strings.IndexByte(s, '[')
	if i < 0 {
		return ""
	}
	j := strings.IndexByte(s[i+1:], ']')
	if j < 0 {
		return ""
	}
	return s[i+1 : i+1+j]
}
