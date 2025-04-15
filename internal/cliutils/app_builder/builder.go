// ALR - Any Linux Repository
// Copyright (C) 2025 Евгений Храмов
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

package appbuilder

import (
	"context"
	"errors"
	"log/slog"

	"github.com/leonelquinteros/gotext"

	"gitea.plemya-x.ru/Plemya-x/ALR/internal/cliutils"
	"gitea.plemya-x.ru/Plemya-x/ALR/internal/config"
	"gitea.plemya-x.ru/Plemya-x/ALR/internal/db"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/distro"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/manager"
	"gitea.plemya-x.ru/Plemya-x/ALR/pkg/repos"
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

type AppBuilder struct {
	deps AppDeps
	err  error
	ctx  context.Context
}

func New(ctx context.Context) *AppBuilder {
	return &AppBuilder{ctx: ctx}
}

func (b *AppBuilder) UseConfig(cfg *config.ALRConfig) *AppBuilder {
	if b.err != nil {
		return b
	}
	b.deps.Cfg = cfg
	return b
}

func (b *AppBuilder) WithConfig() *AppBuilder {
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

func (b *AppBuilder) WithDB() *AppBuilder {
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

func (b *AppBuilder) WithRepos() *AppBuilder {
	b.withRepos(true, false)
	return b
}

func (b *AppBuilder) WithReposForcePull() *AppBuilder {
	b.withRepos(true, true)
	return b
}

func (b *AppBuilder) WithReposNoPull() *AppBuilder {
	b.withRepos(false, false)
	return b
}

func (b *AppBuilder) withRepos(enablePull, forcePull bool) *AppBuilder {
	if b.err != nil {
		return b
	}

	cfg := b.deps.Cfg
	db := b.deps.DB
	if cfg == nil || db == nil {
		b.err = errors.New("config and db are required before initializing repos")
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

func (b *AppBuilder) WithDistroInfo() *AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.Info, b.err = distro.ParseOSRelease(b.ctx)
	if b.err != nil {
		b.err = cliutils.FormatCliExit(gotext.Get("Error parsing os release"), b.err)
	}

	return b
}

func (b *AppBuilder) WithManager() *AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.Manager = manager.Detect()
	if b.deps.Manager == nil {
		b.err = cliutils.FormatCliExit(gotext.Get("Unable to detect a supported package manager on the system"), nil)
	}

	return b
}

func (b *AppBuilder) Build() (*AppDeps, error) {
	if b.err != nil {
		return nil, b.err
	}
	return &b.deps, nil
}
