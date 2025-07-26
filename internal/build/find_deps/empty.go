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
	"log/slog"

	"github.com/goreleaser/nfpm/v2"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/pkg/types"
)

type EmptyFindProvReq struct{}

func (o *EmptyFindProvReq) FindProvides(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist []string) error {
	slog.Info(gotext.Get("AutoProv is not implemented for this package format, so it's skipped"))
	return nil
}

func (o *EmptyFindProvReq) FindRequires(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist []string) error {
	slog.Info(gotext.Get("AutoReq is not implemented for this package format, so it's skipped"))
	return nil
}

func (o *EmptyFindProvReq) BuildDepends(ctx context.Context) ([]string, error) {
	return []string{}, nil
}
