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

package get

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/config/common"
	"go.stplr.dev/stplr/pkg/types"
)

type useCase struct {
	cfg ConfigGetter
	out io.Writer
}

type ConfigGetter interface {
	RootCmd() string
	PagerStyle() string
	AutoPull() bool
	Repos() []types.Repo
	IgnorePkgUpdates() []string
	LogLevel() string
	UseRootCmd() bool
	FirejailExclude() []string
	HideFirejailExcludeWarning() bool
	ForbidSkipInChecksums() bool
	ForbidBuildCommand() bool
	GetPaths() *config.Paths
}

func New(cfg ConfigGetter) *useCase {
	return &useCase{
		cfg: cfg,
		out: os.Stdout,
	}
}

func (u *useCase) Run(ctx context.Context, key string) error {
	stringGetters := map[string]func() string{
		common.ROOT_CMD:    u.cfg.RootCmd,
		common.PAGER_STYLE: u.cfg.PagerStyle,
		common.LOG_LEVEL:   u.cfg.LogLevel,
	}

	boolGetters := map[string]func() bool{
		common.USE_ROOT_CMD:                  u.cfg.UseRootCmd,
		common.AUTO_PULL:                     u.cfg.AutoPull,
		common.FORBID_SKIP_IN_CHECKSUMS:      u.cfg.ForbidSkipInChecksums,
		common.FORBID_BUILD_COMMAND:          u.cfg.ForbidBuildCommand,
		common.HIDE_FIREJAIL_EXCLUDE_WARNING: u.cfg.HideFirejailExcludeWarning,
	}

	listGetters := map[string]func() []string{
		common.IGNORE_PKG_UPDATES: u.cfg.IgnorePkgUpdates,
		common.FIREJAIL_EXCLUDE:   u.cfg.FirejailExclude,
	}

	switch key {
	// TODO: remove legacy keys
	case common.REPO, "repos":
		return u.handleReposKey()
	default:
		return u.handleConfigKey(key, boolGetters, stringGetters, listGetters)
	}
}

func (u *useCase) handleReposKey() error {
	repos := u.cfg.Repos()
	if len(repos) == 0 {
		fmt.Fprintln(u.out, "[]")
		return nil
	}

	repoData, err := yaml.Marshal(repos)
	if err != nil {
		return fmt.Errorf("failed to serialize repos: %w", err)
	}
	fmt.Fprint(u.out, string(repoData))
	return nil
}

func (u *useCase) handleConfigKey(
	key string,
	boolGetters map[string]func() bool,
	stringGetters map[string]func() string,
	listGetters map[string]func() []string,
) error {
	if getter, ok := boolGetters[key]; ok {
		fmt.Fprintln(u.out, getter())
		return nil
	}

	if getter, ok := stringGetters[key]; ok {
		fmt.Fprintln(u.out, getter())
		return nil
	}

	if getter, ok := listGetters[key]; ok {
		u.printList(getter())
		return nil
	}

	return errors.NewI18nError(gotext.Get("unknown config key: %s", key))
}

func (u *useCase) printList(listValue []string) {
	if len(listValue) == 0 {
		fmt.Fprintln(u.out, "[]")
	} else {
		fmt.Fprintln(u.out, strings.Join(listValue, ", "))
	}
}
