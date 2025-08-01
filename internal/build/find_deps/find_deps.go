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

package finddeps

import (
	"context"
	"slices"

	"github.com/goreleaser/nfpm/v2"

	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/types"
)

type ProvReqFinder interface {
	FindProvides(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error
	FindRequires(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error
	BuildDepends(ctx context.Context) ([]string, error)
}

type ProvReqService struct {
	finder ProvReqFinder
}

func New(info *distro.OSRelease, pkgFormat, method string) *ProvReqService {
	s := &ProvReqService{
		finder: &EmptyFindProvReq{},
	}

	if method == "" {
		method = "rpm"
	}

	switch method {
	case "rpm":
		switch {
		case info.ID == "altlinux":
			s.finder = &ALTLinuxFindProvReq{}
		case info.ID == "fedora" || slices.Contains(info.Like, "fedora"):
			s.finder = &FedoraFindProvReq{}
		}
	case "dirty":
		s.finder = &DirtyFindProvReq{}
	}
	return s
}

func (s *ProvReqService) FindProvides(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	return s.finder.FindProvides(ctx, pkgInfo, dirs, skiplist, filter)
}

func (s *ProvReqService) FindRequires(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	return s.finder.FindRequires(ctx, pkgInfo, dirs, skiplist, filter)
}

func (s *ProvReqService) BuildDepends(ctx context.Context) ([]string, error) {
	return s.finder.BuildDepends(ctx)
}
