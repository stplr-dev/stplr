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
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/config/common"
)

type useCase struct {
	cfg *config.ALRConfig
}

func New(cfg *config.ALRConfig) *useCase {
	return &useCase{
		cfg: cfg,
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

	switch key {
	case common.FIREJAIL_EXCLUDE:
		updates := u.cfg.FirejailExclude()
		if len(updates) == 0 {
			fmt.Println("[]")
		} else {
			fmt.Println(strings.Join(updates, ", "))
		}
	case common.IGNORE_PKG_UPDATES:
		updates := u.cfg.IgnorePkgUpdates()
		if len(updates) == 0 {
			fmt.Println("[]")
		} else {
			fmt.Println(strings.Join(updates, ", "))
		}
	case "repo", "repos":
		repos := u.cfg.Repos()
		if len(repos) == 0 {
			fmt.Println("[]")
		} else {
			repoData, err := yaml.Marshal(repos)
			if err != nil {
				return fmt.Errorf("failed to serialize repos: %w", err)
			}
			fmt.Print(string(repoData))
		}
	default:
		if getter, ok := boolGetters[key]; ok {
			fmt.Println(getter())
		} else if getter, ok := stringGetters[key]; ok {
			fmt.Println(getter())
		} else {
			return errors.NewI18nError(gotext.Get("unknown config key: %s", key))
		}
	}

	return nil
}
