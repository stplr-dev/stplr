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
	"fmt"
	"log/slog"

	stdErrors "errors"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/config/savers"
	"go.stplr.dev/stplr/internal/copier"
	"go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/installer"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/plugins"
	"go.stplr.dev/stplr/internal/scripter"
	"go.stplr.dev/stplr/internal/search"
	"go.stplr.dev/stplr/internal/service/updater"
	"go.stplr.dev/stplr/internal/sys"
	"go.stplr.dev/stplr/internal/utils"

	"go.stplr.dev/stplr/internal/service/repos"

	"go.stplr.dev/stplr/pkg/distro"
)

type Cleanup func()

type Deps struct {
	UID int
	GID int
	WD  string

	Output output.Output

	Cfg      *config.ALRConfig
	Manager  manager.Manager
	DB       *db.Database
	Repos    *repos.Repos
	Builder  *build.Builder
	Info     *distro.OSRelease
	Searcher *search.Searcher
	Updater  *updater.Updater

	PluginProvider     *plugins.Provider
	RootPluginProvider *plugins.Provider

	Puller             repos.PullExecutor
	Copier             copier.CopierExecutor
	Installer          installer.InstallerExecutor
	Scripter           scripter.ScriptExecutor
	SystemConfigWriter savers.SystemConfigWriterExecutor

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
	sys  *sys.Sys
}

func Start(ctx context.Context) *builder {
	return &builder{
		ctx: ctx,
		deps: &Deps{
			Output: output.NewConsoleOutput(),
		},
		sys: &sys.Sys{},
	}
}

func (b *builder) UserContext() *builder {
	if b.err != nil {
		return b
	}

	uid := b.sys.Getuid()
	gid := b.sys.Getgid()
	wd, err := b.sys.Getwd()
	if err != nil {
		b.err = errors.WrapIntoI18nError(err, gotext.Get("Failed to get working directory"))
		return b
	}

	b.deps.UID = uid
	b.deps.GID = gid
	b.deps.WD = wd

	return b
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

func (b *builder) ConfigRW() *builder {
	if b.err != nil {
		return b
	}

	cfg := config.New(
		config.WithSystemConfigWriter(b.deps.SystemConfigWriter),
	)
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

func (b *builder) Scripter() *builder {
	if b.err != nil {
		return b
	}

	isBuilder, err := utils.IsBuilderUser()
	if err != nil {
		b.err = err
		return b
	}

	if !isBuilder {
		err = config.PatchToUserDirs(b.deps.Cfg)
		if err != nil {
			b.err = err
			return b
		}
	}

	b.deps.Scripter = scripter.NewLocalScriptExecutor(b.deps.Cfg, b.deps.Output)

	return b
}

func (b *builder) PatchToUserDirs() *builder {
	if b.err != nil {
		return b
	}

	isBuilder, err := utils.IsBuilderUser()
	if err != nil {
		b.err = err
		return b
	}

	if !isBuilder {
		err = config.PatchToUserDirs(b.deps.Cfg)
		if err != nil {
			b.err = err
			return b
		}
	}

	return b
}

func (b *builder) ScripterFromPlugin() *builder {
	if b.err != nil {
		return b
	}

	res, err := plugins.GetScripter(b.ctx, b.deps.PluginProvider)
	if err != nil {
		b.err = err
		return b
	}

	b.deps.Scripter = res

	return b
}

func (b *builder) Repos() *builder {
	if b.err != nil {
		return b
	}

	cfg := b.deps.Cfg
	if cfg == nil {
		b.err = stdErrors.New("config is required for initializing repos")
		return b
	}

	b.deps.Repos = repos.New(cfg, b.deps.DB, b.deps.Puller, b.deps.Output)

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
		b.deps.DB,
		b.deps.Manager,
		b.deps.Repos,
		b.deps.Scripter,
		b.deps.Installer,
		b.deps.Output,
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

	b.deps.Updater = updater.New(b.deps.Cfg, b.deps.Manager, b.deps.Info, b.deps.Searcher)

	return b
}

func (b *builder) DropCaps() *builder {
	if b.err != nil {
		return b
	}

	if utils.IsRoot() {
		if err := cliutils.ExitIfCantDropCapsToBuilderUserNoPrivs(); err != nil {
			b.err = err
			return nil
		}
	}

	return b
}

func (b *builder) PluginProvider() *builder {
	if b.err != nil {
		return b
	}

	b.deps.PluginProvider = plugins.NewProvider(b.deps.Output)
	err := b.deps.PluginProvider.SetupConnection()
	if err != nil {
		b.err = err
		return b
	}
	cleanup := func() {
		err := b.deps.PluginProvider.Cleanup()
		if err != nil {
			slog.Warn("failed to cleanup PluginProvider")
		}
	}
	b.deps.cleanups = append(b.deps.cleanups, cleanup)

	return b
}

func (b *builder) PullerFromPlugin() *builder {
	if b.err != nil {
		return b
	}

	var err error

	b.deps.Puller, err = plugins.GetPuller(b.ctx, b.deps.PluginProvider)
	if err != nil {
		b.err = fmt.Errorf("failed to get puller from plugin: %w", err)
		return b
	}

	return b
}

func (b *builder) Puller() *builder {
	if b.err != nil {
		return b
	}

	b.deps.Puller = repos.NewPuller(b.deps.Cfg, b.deps.DB)

	return b
}

func (b *builder) RootPluginProvider() *builder {
	if b.err != nil {
		return b
	}

	b.deps.RootPluginProvider = plugins.NewProvider(b.deps.Output)
	err := b.deps.RootPluginProvider.SetupRootConnection()
	if err != nil {
		b.err = err
		return b
	}
	cleanup := func() {
		err := b.deps.RootPluginProvider.Cleanup()
		if err != nil {
			slog.Warn("failed to cleanup RootPluginProvider")
		}
	}
	b.deps.cleanups = append(b.deps.cleanups, cleanup)

	return b
}

func (b *builder) InstallerFromPlugin() *builder {
	if b.err != nil {
		return b
	}

	var err error

	b.deps.Installer, err = plugins.GetInstaller(b.ctx, b.deps.RootPluginProvider)
	if err != nil {
		b.err = err
		return b
	}

	return b
}

func (b *builder) CopierFromRootPlugin() *builder {
	if b.err != nil {
		return b
	}

	var err error
	b.deps.Copier, err = plugins.GetCopier(b.ctx, b.deps.RootPluginProvider)
	if err != nil {
		b.err = fmt.Errorf("failed to get copier: %w", err)
		return b
	}

	return b
}

func (b *builder) CopierForRootPlugin() *builder {
	if b.err != nil {
		return b
	}

	var err error
	b.deps.Copier, err = copier.New(b.deps.UID, b.deps.GID, b.deps.WD)
	if err != nil {
		b.err = fmt.Errorf("failed to init copier: %w", err)
		return b
	}

	return b
}

func (b *builder) SystemConfigWriterFromRootPlugin() *builder {
	if b.err != nil {
		return b
	}

	var err error
	b.deps.SystemConfigWriter, err = plugins.GetSystemConfigWriter(b.ctx, b.deps.RootPluginProvider)
	if err != nil {
		b.err = fmt.Errorf("failed to get system-config-writer: %w", err)
		return b
	}

	return b
}

func (b *builder) SystemConfigWriter() *builder {
	if b.err != nil {
		return b
	}

	b.deps.SystemConfigWriter = &savers.SystemConfigWriter{}

	return b
}

func (b *builder) SetupPluginOutput() *builder {
	if b.err != nil {
		return b
	}

	b.deps.Output = output.NewPluginOutput()

	return b
}

func (b *builder) End() (*Deps, error) {
	if b.err != nil {
		b.deps.Cleanup()
		return nil, b.err
	}
	return b.deps, nil
}
