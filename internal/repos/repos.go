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

package repos

import (
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/config"
	database "go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/pkg/types"
)

type Config interface {
	GetPaths() *config.Paths
	Repos() []types.Repo
}

type Repos struct {
	cfg Config
	db  *database.Database

	rp *RepoProcessor
	gm *GitManager

	out output.Output
}

func New(
	cfg Config,
	db *database.Database,
	out output.Output,
) *Repos {
	gm := &GitManager{}
	return &Repos{
		cfg,
		db,
		&RepoProcessor{
			gm,
		},
		gm,
		out,
	}
}
