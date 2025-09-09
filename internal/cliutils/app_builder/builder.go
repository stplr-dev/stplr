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

package appBuilder

import (
	"context"
	"errors"
	"log/slog"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/repos"
	"go.stplr.dev/stplr/pkg/distro"
)

type AppDeps struct {
	Cfg     *config.ALRConfig
	DB      *db.Database
	Repos   *repos.Repos
	Info    *distro.OSRelease
	Manager manager.Manager
}

func (d *AppDeps) Defer() {
	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			slog.Warn("failed to close db", "err", err)
		}
	}
}

type appBuilder struct {
	deps AppDeps
	err  error
	ctx  context.Context
}

type AppBuilder interface {
	UseConfig(cfg *config.ALRConfig) AppBuilder
	WithConfig() AppBuilder
	WithDB() AppBuilder
	WithRepos() AppBuilder
	WithReposForcePull() AppBuilder
	WithReposNoPull() AppBuilder
	WithManager() AppBuilder
	WithDistroInfo() AppBuilder
	Build() (*AppDeps, error)
}

func New(ctx context.Context) *appBuilder {
	return &appBuilder{ctx: ctx}
}

func (b *appBuilder) UseConfig(cfg *config.ALRConfig) AppBuilder {
	if b.err != nil {
		return b
	}
	b.deps.Cfg = cfg
	return b
}

func (b *appBuilder) WithConfig() AppBuilder {
	if b.err != nil {
		return b
	}

	cfg := config.New()
	if err := cfg.Load(); err != nil {
		b.err = cliutils.FormatCliExit(gotext.Get("Error loading config"), err)
		return b
	}

	b.deps.Cfg = cfg
	return b
}

func (b *appBuilder) WithDB() AppBuilder {
	if b.err != nil {
		return b
	}

	cfg := b.deps.Cfg
	if cfg == nil {
		b.err = errors.New("config is required before initializing DB")
		return b
	}

	db := db.New(cfg)
	if err := db.Init(b.ctx); err != nil {
		b.err = cliutils.FormatCliExit(gotext.Get("Error initialization database"), err)
		return b
	}

	b.deps.DB = db
	return b
}

func (b *appBuilder) WithRepos() AppBuilder {
	b.withRepos(true, false)
	return b
}

func (b *appBuilder) WithReposForcePull() AppBuilder {
	b.withRepos(true, true)
	return b
}

func (b *appBuilder) WithReposNoPull() AppBuilder {
	b.withRepos(false, false)
	return b
}

func (b *appBuilder) withRepos(enablePull, forcePull bool) AppBuilder {
	if b.err != nil {
		return b
	}

	cfg := b.deps.Cfg
	db := b.deps.DB
	info := b.deps.Info

	if info == nil {
		b.WithDistroInfo()
		info = b.deps.Info
	}

	if cfg == nil || db == nil || info == nil {
		b.err = errors.New("config, db and info are required before initializing repos")
		return b
	}

	rs := repos.New(cfg, db)

	if enablePull && (forcePull || cfg.AutoPull()) {
		if err := rs.Pull(b.ctx, cfg.Repos()); err != nil {
			b.err = cliutils.FormatCliExit(gotext.Get("Error pulling repositories"), err)
			return b
		}
	}

	b.deps.Repos = rs

	return b
}

func (b *appBuilder) WithDistroInfo() AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.Info, b.err = distro.ParseOSRelease(b.ctx)
	if b.err != nil {
		b.err = cliutils.FormatCliExit(gotext.Get("Error parsing os-release file"), b.err)
	}

	return b
}

func (b *appBuilder) WithManager() AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.Manager = manager.Detect()
	if b.deps.Manager == nil {
		b.err = cliutils.FormatCliExit(gotext.Get("Unable to detect a supported package manager on the system"), nil)
	}

	return b
}

func (b *appBuilder) Build() (*AppDeps, error) {
	if b.err != nil {
		return nil, b.err
	}
	return &b.deps, nil
}
