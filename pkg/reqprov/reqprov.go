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

package reqprov

import (
	"context"
	"fmt"

	"github.com/goreleaser/nfpm/v2"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/types"

	"go.stplr.dev/stplr/pkg/reqprov/dirty"
	"go.stplr.dev/stplr/pkg/reqprov/rpm"
)

type ReqProvFinder interface {
	FindRequires(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error
	FindProvides(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error
	BuildDepends(ctx context.Context) ([]string, error)
}

type ReqProvService struct {
	finder ReqProvFinder
}

func getFinder(info *distro.OSRelease, method string) (ReqProvFinder, error) {
	if method == "" {
		method = "rpm"
	}
	switch method {
	case "rpm":
		switch {
		case distro.IsIdEqualOrLike(info, "altlinux"):
			return rpm.NewALTLinux(), nil
		case distro.IsIdEqualOrLike(info, "fedora"):
			return rpm.NewFedora(), nil
		default:
			return nil, fmt.Errorf("unsupported RPM-based distro: %s", info.ID)
		}
	case "dirty":
		return dirty.New(), nil
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func New(info *distro.OSRelease, pkgFormat, method string) (*ReqProvService, error) {
	finder, err := getFinder(info, method)
	if err != nil {
		return nil, fmt.Errorf("cannot getFinder: %w", err)
	}

	return &ReqProvService{
		finder: finder,
	}, nil
}

func (s *ReqProvService) FindProvides(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	return s.finder.FindProvides(ctx, out, pkgInfo, dirs, skiplist, filter)
}

func (s *ReqProvService) FindRequires(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	return s.finder.FindRequires(ctx, out, pkgInfo, dirs, skiplist, filter)
}

func (s *ReqProvService) BuildDepends(ctx context.Context) ([]string, error) {
	return s.finder.BuildDepends(ctx)
}
