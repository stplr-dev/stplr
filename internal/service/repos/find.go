// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package repos

import (
	"context"
	"fmt"
	"strings"

	"go.stplr.dev/stplr/pkg/staplerfile"
)

func (rs *Repos) FindPkgs(ctx context.Context, pkgs []string) (map[string][]staplerfile.Package, []string, error) {
	found := make(map[string][]staplerfile.Package)
	var notFound []string

	for _, pkgName := range pkgs {
		if pkgName == "" {
			continue
		}

		result, err := rs.lookupPkg(ctx, pkgName)
		if err != nil {
			return nil, nil, fmt.Errorf("FindPkgs: lookup for %q failed: %w", pkgName, err)
		}

		if len(result) == 0 {
			notFound = append(notFound, pkgName)
		} else {
			found[pkgName] = result
		}
	}

	return found, notFound, nil
}

func (rs *Repos) lookupPkg(ctx context.Context, pkgName string) ([]staplerfile.Package, error) {
	var result []staplerfile.Package
	var err error

	if name, repo, ok := extractNameAndRepo(pkgName); ok {
		result, err = rs.db.GetPkgs(ctx, "name = ? AND repository = ?", name, repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get by name and repo: %w", err)
		}
	} else {
		result, err = rs.db.GetPkgs(ctx, "json_array_contains(provides, ?)", pkgName)
		if err != nil {
			return nil, fmt.Errorf("get by provides: %w", err)
		}

		if len(result) == 0 {
			result, err = rs.db.GetPkgs(ctx, "name LIKE ?", pkgName)
			if err != nil {
				return nil, fmt.Errorf("failed to get by name: %w", err)
			}
		}
	}

	return result, nil
}

func extractNameAndRepo(pkgName string) (string, string, bool) {
	switch {
	case strings.Contains(pkgName, "/"):
		// repo/pkg
		parts := strings.SplitN(pkgName, "/", 2)
		return parts[1], parts[0], true
	case strings.Contains(pkgName, "+stplr-"):
		// pkg+stplr-repo
		parts := strings.SplitN(pkgName, "+stplr-", 2)
		return parts[0], parts[1], true
	default:
		return "", "", false
	}
}
