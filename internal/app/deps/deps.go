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

package deps

import (
	"context"

	"go.stplr.dev/stplr/internal/app/deps/internal/builder"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/build"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/config/savers"
	"go.stplr.dev/stplr/internal/copier"
	"go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/installer"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/scripter"
	"go.stplr.dev/stplr/internal/search"
	"go.stplr.dev/stplr/internal/service/repos"
	"go.stplr.dev/stplr/internal/service/updater"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/dlcache"
)

type WithRepos struct {
	Repos *repos.Repos
}

type Cleanup func()

type RemoveDeps struct {
	Mgr manager.Manager
}

func ForRemoveAction(ctx context.Context) (*RemoveDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Manager().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &RemoveDeps{
		Mgr: b.Manager,
	}, b.Cleanup, nil
}

type RemoveShellCompDeps struct {
	Cfg *config.ALRConfig
	DB  *db.Database
	Mgr manager.Manager
}

func ForRemoveShellComp(ctx context.Context) (*RemoveShellCompDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		DB().
		Manager().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &RemoveShellCompDeps{
		Cfg: b.Cfg,
		DB:  b.DB,
		Mgr: b.Manager,
	}, b.Cleanup, nil
}

type InstallActionDeps struct {
	Builder *build.Builder
	Manager manager.Manager
	Info    *distro.OSRelease
}

func ForInstallAction(ctx context.Context) (*InstallActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		RootPluginProvider().
		InstallerFromPlugin().
		// Drop caps
		DropCaps().
		PluginProvider().
		ScripterFromPlugin().
		Config().
		Manager().
		DB().
		Repos().
		Info().
		Builder().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &InstallActionDeps{
		Builder: b.Builder,
		Manager: b.Manager,
		Info:    b.Info,
	}, b.Cleanup, nil
}

type InstallShellCompDeps struct {
	DB *db.Database
}

func ForInstallShellComp(ctx context.Context) (*InstallShellCompDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		DB().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &InstallShellCompDeps{
		DB: b.DB,
	}, b.Cleanup, nil
}

type UpgradeDeps struct {
	Builder *build.Builder
	Manager manager.Manager
	Info    *distro.OSRelease
	DB      *db.Database
	Updater *updater.Updater
}

func ForUpgradeAction(ctx context.Context) (*UpgradeDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		RootPluginProvider().
		InstallerFromPlugin().
		DropCaps().
		PluginProvider().
		ScripterFromPlugin().
		Config().
		Manager().
		DB().
		Repos().
		Info().
		Builder().
		Searcher().
		Updater().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &UpgradeDeps{
		Builder: b.Builder,
		Manager: b.Manager,
		Info:    b.Info,
		DB:      b.DB,
		Updater: b.Updater,
	}, b.Cleanup, nil
}

type ListActionDeps struct {
	Config  *config.ALRConfig
	DB      *db.Database
	Info    *distro.OSRelease
	Updater *updater.Updater
}

func ForListAction(ctx context.Context) (*ListActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		DB().
		Info().
		Manager().
		Searcher().
		Updater().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &ListActionDeps{
		Config:  b.Cfg,
		DB:      b.DB,
		Info:    b.Info,
		Updater: b.Updater,
	}, b.Cleanup, nil
}

type ConfigShowActionDeps struct {
	Config *config.ALRConfig
}

func ForConfigShowAction(ctx context.Context) (*ConfigShowActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &ConfigShowActionDeps{
		Config: b.Cfg,
	}, b.Cleanup, nil
}

type ConfigSetActionDeps struct {
	Config *config.ALRConfig
}

func ForConfigSetAction(ctx context.Context) (*ConfigSetActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		SystemConfigWriter().
		ConfigRW().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &ConfigSetActionDeps{
		Config: b.Cfg,
	}, b.Cleanup, nil
}

type ConfigGetActionDeps struct {
	Config *config.ALRConfig
}

func ForConfigGetAction(ctx context.Context) (*ConfigGetActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &ConfigGetActionDeps{
		Config: b.Cfg,
	}, b.Cleanup, nil
}

type InfoActionDeps struct {
	Repos *repos.Repos
	Info  *distro.OSRelease
}

func ForInfoAction(ctx context.Context) (*InfoActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		DB().
		Info().
		Repos().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &InfoActionDeps{
		Repos: b.Repos,
		Info:  b.Info,
	}, b.Cleanup, nil
}

type InfoShellCompDeps struct {
	DB *db.Database
}

func ForInfoShellComp(ctx context.Context) (*InfoShellCompDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		DB().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &InfoShellCompDeps{
		DB: b.DB,
	}, b.Cleanup, nil
}

type FixActionDeps struct {
	Config *config.ALRConfig
}

func ForFixAction(ctx context.Context) (*FixActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		DropCaps().
		Config().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &FixActionDeps{
		Config: b.Cfg,
	}, b.Cleanup, nil
}

func ReposGetter(ctx context.Context) (*repos.Repos, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		DropCaps().
		Config().
		DB().
		PluginProvider().
		PullerFromPlugin().
		Repos().
		End()
	if err != nil {
		return nil, nil, err
	}

	return b.Repos, b.Cleanup, nil
}

type SearchDeps struct {
	Searcher *search.Searcher
	Info     *distro.OSRelease
}

func ForSearchAction(ctx context.Context) (*SearchDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		DB().
		Searcher().
		Info().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &SearchDeps{
		Searcher: b.Searcher,
		Info:     b.Info,
	}, b.Cleanup, nil
}

type RepoAddDeps struct {
	Config *config.ALRConfig
	// Puller repos.PullExecutor
	Repos *repos.Repos
}

func ForRepoAddAction(ctx context.Context) (*RepoAddDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		RootPluginProvider().
		SystemConfigWriterFromRootPlugin().
		ConfigRW().
		DropCaps().
		PluginProvider().
		PullerFromPlugin().
		DB().
		Repos().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &RepoAddDeps{
		Config: b.Cfg,
		Repos:  b.Repos,
	}, b.Cleanup, nil
}

type RepoRemoveDeps struct {
	Config *config.ALRConfig
	DB     *db.Database
}

func ForRepoRemoveAction(ctx context.Context) (*RepoRemoveDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		SystemConfigWriter().
		ConfigRW().
		DB().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &RepoRemoveDeps{
		Config: b.Cfg,
		DB:     b.DB,
	}, b.Cleanup, nil
}

type UniversalReposModificationActionDeps struct {
	Config *config.ALRConfig
	Repos  *repos.Repos
}

func ForUniversalReposModificationActionDeps(ctx context.Context) (*UniversalReposModificationActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		RootPluginProvider().
		SystemConfigWriterFromRootPlugin().
		ConfigRW().
		DB().
		DropCaps().
		PluginProvider().
		PullerFromPlugin().
		Repos().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &UniversalReposModificationActionDeps{
		Config: b.Cfg,
		Repos:  b.Repos,
	}, b.Cleanup, nil
}

type RepoSetRefDeps struct {
	Config *config.ALRConfig
	Repos  *repos.Repos
}

type PluginServeDeps struct {
	Puller   repos.PullExecutor
	Scripter scripter.ScriptExecutor
}

func ForPluginsServe(ctx context.Context) (*PluginServeDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		SetupPluginOutput().
		Config().
		Scripter().
		DB().
		Puller().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &PluginServeDeps{
		b.Puller,
		b.Scripter,
	}, b.Cleanup, nil
}

type RootPluginServeDeps struct {
	Installer          installer.InstallerExecutor
	Copier             copier.CopierExecutor
	SystemConfigWriter savers.SystemConfigWriterExecutor
}

func ForPluginsServeRoot(ctx context.Context) (*RootPluginServeDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Manager().
		Config().
		UserContext().
		CopierForRootPlugin().
		SystemConfigWriter().
		End()
	if err != nil {
		return nil, nil, err
	}

	needRootCmd := false
	rootCmd := ""

	if !utils.IsRoot() {
		needRootCmd = b.Cfg.UseRootCmd()
		rootCmd = b.Cfg.RootCmd()
	}

	return &RootPluginServeDeps{
		Installer:          installer.New(b.Manager, needRootCmd, rootCmd),
		Copier:             b.Copier,
		SystemConfigWriter: b.SystemConfigWriter,
	}, b.Cleanup, nil
}

type BuildActionDeps struct {
	Config  *config.ALRConfig
	Builder *build.Builder
	Info    *distro.OSRelease
	Copier  copier.CopierExecutor
	Manager manager.Manager
	Repos   *repos.Repos
}

func ForBuildAction(ctx context.Context) (*BuildActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		RootPluginProvider().
		InstallerFromPlugin().
		CopierFromRootPlugin().
		// PatchConfigToUserDirs().
		//
		// Drop to builder user
		DropCaps().
		PatchToUserDirs().
		PluginProvider().
		ScripterFromPlugin().
		Manager().
		DB().
		Repos().
		Info().
		Builder().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &BuildActionDeps{
		Builder: b.Builder,
		Info:    b.Info,
		Copier:  b.Copier,
		Manager: b.Manager,
		Repos:   b.Repos,
		Config:  b.Cfg,
	}, b.Cleanup, nil
}

type RefreshActionDeps struct {
	Repos *repos.Repos
}

func ForRefreshAction(ctx context.Context) (*RefreshActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		DropCaps().
		DB().
		PluginProvider().
		PullerFromPlugin().
		Repos().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &RefreshActionDeps{
		Repos: b.Repos,
	}, b.Cleanup, nil
}

type MigrateActionDeps struct {
	DbResetter      *db.Resetter
	DlCacheResetter *dlcache.Resetter
	Repos           *repos.Repos
}

func ForMigrateAction(ctx context.Context) (*MigrateActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		Config().
		End()
	if err != nil {
		return nil, nil, err
	}

	return &MigrateActionDeps{
		DbResetter:      db.NewResetter(b.Cfg),
		DlCacheResetter: dlcache.NewResetter(b.Cfg),
	}, b.Cleanup, nil
}

type SupportActionDeps struct {
	Out output.Output
}

func ForSupportAction(ctx context.Context) (*SupportActionDeps, Cleanup, error) {
	b, err := builder.
		Start(ctx).
		End()
	if err != nil {
		return nil, nil, err
	}

	return &SupportActionDeps{
		Out: b.Output,
	}, b.Cleanup, nil
}
