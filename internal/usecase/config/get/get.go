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
		"rootCmd":    u.cfg.RootCmd,
		"pagerStyle": u.cfg.PagerStyle,
		"logLevel":   u.cfg.LogLevel,
	}

	boolGetters := map[string]func() bool{
		"useRootCmd":            u.cfg.UseRootCmd,
		"autoPull":              u.cfg.AutoPull,
		"forbidSkipInChecksums": u.cfg.ForbidSkipInChecksums,
		"forbidBuildCommand":    u.cfg.ForbidBuildCommand,
	}

	switch key {
	case "ignorePkgUpdates":
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
