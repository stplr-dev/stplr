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

package build

import (
	"context"
	"os"
	"path/filepath"

	"github.com/goreleaser/nfpm/v2"

	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/scripter"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type LocalCacheExecutor struct{ cfg commonbuild.Config }

func NewLocalCacheExecutor(cfg commonbuild.Config) *LocalCacheExecutor {
	return &LocalCacheExecutor{cfg: cfg}
}

type CheckForBuiltPackageInput interface {
	commonbuild.OSReleaser
	commonbuild.PkgFormatter
	commonbuild.RepositoryGetter
	commonbuild.BuildOptsProvider
}

func (c *LocalCacheExecutor) CheckForBuiltPackage(
	ctx context.Context,
	input CheckForBuiltPackageInput,
	pkg *staplerfile.Package,
) (string, bool, error) {
	filename, err := pkgFileName(input, pkg)
	if err != nil {
		return "", false, err
	}

	pkgPath := filepath.Join(getBaseDir(c.cfg, pkg.Name), filename)

	_, err = os.Stat(pkgPath)
	if err != nil {
		return "", false, nil
	}

	return pkgPath, true, nil
}

func pkgFileName(input CheckForBuiltPackageInput, pkg *staplerfile.Package) (string, error) {
	pkgInfo := scripter.GetBasePkgInfo(pkg, input)

	packager, err := nfpm.Get(input.PkgFormat())
	if err != nil {
		return "", err
	}

	return packager.ConventionalFileName(pkgInfo), nil
}
