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

package build

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/goreleaser/nfpm/v2"

	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/cpu"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/overrides"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
)

var RegexpALRPackageName = regexp.MustCompile(`^(?P<package>[^+]+)\+stplr-(?P<repo>.+)$`)

type getBasePkgInfoInput interface {
	commonbuild.OSReleaser
	commonbuild.RepositoryGetter
	commonbuild.BuildOptsProvider
}

func getBasePkgInfo(pkg *alrsh.Package, input getBasePkgInfoInput) *nfpm.Info {
	var name string
	if input.BuildOpts().NoSuffix {
		name = pkg.Name
	} else {
		name = fmt.Sprintf("%s+stplr-%s", pkg.Name, input.Repository())
	}
	return &nfpm.Info{
		Name:    name,
		Arch:    cpu.Arch(),
		Version: pkg.Version,
		Release: overrides.ReleasePlatformSpecific(pkg.Release, input.OSRelease()),
		Epoch:   strconv.FormatUint(uint64(pkg.Epoch), 10),
	}
}

// Функция getPkgFormat возвращает формат пакета из менеджера пакетов,
// или STPLR_PKG_FORMAT, если он установлен.
func GetPkgFormat(mgr manager.Manager) string {
	pkgFormat := mgr.Format()
	if format, ok := os.LookupEnv("STPLR_PKG_FORMAT"); ok {
		pkgFormat = format
	}
	return pkgFormat
}

// Функция removeDuplicates убирает любые дубликаты из предоставленного среза.
func removeDuplicates[T comparable](slice []T) []T {
	seen := map[T]struct{}{}
	result := []T{}

	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

func removeDuplicatesSources(sources, checksums []string) ([]string, []string) {
	seen := map[string]string{}
	keys := make([]string, 0)
	for i, s := range sources {
		if val, ok := seen[s]; !ok || strings.EqualFold(val, "SKIP") {
			if !ok {
				keys = append(keys, s)
			}
			seen[s] = checksums[i]
		}
	}

	newSources := make([]string, len(keys))
	newChecksums := make([]string, len(keys))
	for i, k := range keys {
		newSources[i] = k
		newChecksums[i] = seen[k]
	}
	return newSources, newChecksums
}
