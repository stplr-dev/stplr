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

package builder

import (
	"context"

	stdErrors "errors"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/repos"
	"go.stplr.dev/stplr/internal/search"
	"go.stplr.dev/stplr/internal/service/updater"
	"go.stplr.dev/stplr/internal/sys"

	repos2 "go.stplr.dev/stplr/internal/service/repos"

	"go.stplr.dev/stplr/pkg/distro"
)

type Cleanup func()

type Deps struct {
	Cfg       *config.ALRConfig
	Manager   manager.Manager
	DB        *db.Database
	Repos     *repos.Repos
	Repos2    *repos2.Repos
	Installer build.InstallerExecutor
	Scripter  build.ScriptExecutor
	Builder   *build.Builder
	Info      *distro.OSRelease
	Searcher  *search.Searcher
	Updater   *updater.Updater

	cleanups []Cleanup
}

func (d *Deps) Cleanup() {
	for i := len(d.cleanups) - 1; i >= 0; i-- {
		d.cleanups[i]()
	}
}

type builder struct {
	ctx  context.Context
	err  error
	deps *Deps
}

func Start(ctx context.Context) *builder {
	return &builder{
		ctx:  ctx,
		deps: &Deps{},
	}
}

func (b *builder) Config() *builder {
	if b.err != nil {
		return b
	}

	cfg := config.New()
	if err := cfg.Load(); err != nil {
		b.err = errors.WrapIntoI18nError(err, gotext.Get("Error loading config"))
		return b
	}

	b.deps.Cfg = cfg
	return b
}

func (b *builder) Manager() *builder {
	if b.err != nil {
		return b
	}

	b.deps.Manager = manager.Detect()
	if b.deps.Manager == nil {
		b.err = errors.NewI18nError(gotext.Get("Unable to detect a supported package manager on the system"))
	}

	return b
}

func (b *builder) DB() *builder {
	if b.err != nil {
		return b
	}

	cfg := b.deps.Cfg
	if cfg == nil {
		b.err = stdErrors.New("config is required before initializing DB")
		return b
	}

	db := db.New(cfg)
	if err := db.Init(b.ctx); err != nil {
		b.err = errors.WrapIntoI18nError(err, gotext.Get("Error initialization database"))
		return b
	}

	b.deps.DB = db

	b.deps.cleanups = append(b.deps.cleanups, func() {
		_ = db.Close()
	})

	return b
}

func (b *builder) InstallerAndScripter() *builder {
	if b.err != nil {
		return b
	}

	res, cleanup, err := build.PrepareInstallerAndScripter(b.ctx)
	if err != nil {
		b.err = err
		return b
	}
	b.deps.cleanups = append(b.deps.cleanups, cleanup)

	b.deps.Installer, b.deps.Scripter = res.Installer, res.Scripter

	return b
}

func (b *builder) Repos() *builder {
	if b.err != nil {
		return b
	}

	cfg := b.deps.Cfg
	db := b.deps.DB

	if cfg == nil || db == nil {
		b.err = stdErrors.New("config and db are required for initializing repos")
		return b
	}

	b.deps.Repos = repos.New(cfg, db)

	return b
}

func (b *builder) Repos2() *builder {
	if b.err != nil {
		return b
	}

	b.deps.Repos2 = repos2.New(b.deps.Cfg, &sys.Sys{}, b.deps.Repos)

	return b
}

func (b *builder) Info() *builder {
	if b.err != nil {
		return b
	}

	info, err := distro.ParseOSRelease(b.ctx)
	if err != nil {
		b.err = errors.WrapIntoI18nError(err, gotext.Get("Error parsing os-release file"))
	}

	b.deps.Info = info

	return b
}

func (b *builder) Builder() *builder {
	if b.err != nil {
		return b
	}

	builder, err := build.NewMainBuilder(
		b.deps.Cfg,
		b.deps.Manager,
		b.deps.Repos,
		b.deps.Scripter,
		b.deps.Installer,
	)
	if err != nil {
		b.err = err
		return b
	}

	b.deps.Builder = builder

	return b
}

func (b *builder) Searcher() *builder {
	if b.err != nil {
		return b
	}

	b.deps.Searcher = search.New(b.deps.DB)

	return b
}

func (b *builder) Updater() *builder {
	if b.err != nil {
		return b
	}

	b.deps.Updater = updater.New(b.deps.Manager, b.deps.Info, b.deps.Searcher)

	return b
}

func (b *builder) End() (*Deps, error) {
	if b.err != nil {
		b.deps.Cleanup()
		return nil, b.err
	}
	return b.deps, nil
}
