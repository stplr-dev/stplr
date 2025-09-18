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

package set

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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

type Options struct {
	Field string
	Value string
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	key := opts.Field
	value := opts.Value

	stringSetters := map[string]func(string){
		"rootCmd":    u.cfg.System.SetRootCmd,
		"pagerStyle": u.cfg.System.SetPagerStyle,
		"logLevel":   u.cfg.System.SetLogLevel,
	}

	boolSetters := map[string]func(bool){
		"useRootCmd":            u.cfg.System.SetUseRootCmd,
		"autoPull":              u.cfg.System.SetAutoPull,
		"forbidSkipInChecksums": u.cfg.System.SetForbidSkipInChecksums,
		"forbidBuildCommand":    u.cfg.System.SetForbidBuildCommand,
	}

	switch key {
	case "ignorePkgUpdates":
		var updates []string
		if value != "" {
			updates = strings.Split(value, ",")
			for i := range updates {
				updates[i] = strings.TrimSpace(updates[i])
			}
		}
		u.cfg.System.SetIgnorePkgUpdates(updates)

	case "repo", "repos":
		return errors.NewI18nError(gotext.Get("use 'repo add/remove' commands to manage repositories"))

	default:
		if setter, ok := stringSetters[key]; ok {
			setter(value)
		} else if setter, ok := boolSetters[key]; ok {
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return errors.WrapIntoI18nError(err, gotext.Get("invalid boolean value for %s: %s", key, value))
			}
			setter(boolValue)
		} else {
			return errors.NewI18nError(gotext.Get("unknown config key: %s", key))
		}
	}

	if err := u.cfg.System.Save(); err != nil {
		return errors.NewI18nError(gotext.Get("failed to save config"))
	}

	fmt.Println(gotext.Get("Successfully set %s = %s", key, value))
	return nil
}
