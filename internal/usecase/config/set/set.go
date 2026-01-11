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
	"slices"

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

type Options struct {
	Field string
	Value string
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	key := opts.Field
	valueStr := opts.Value

	if !slices.Contains(config.AllowedKeys(), key) {
		return errors.NewI18nError(gotext.Get("unknown config key: %s", key))
	}

	value, err := config.ConvertValue(key, valueStr)
	if err != nil {
		return errors.NewI18nError(gotext.Get("invalid value for %s: %v", key, err))
	}

	if err := u.cfg.SetToAndSave(common.SOURCE_SYSTEM, key, value); err != nil {
		return errors.NewI18nError(gotext.Get("failed to save config: %v", err))
	}

	fmt.Println(gotext.Get("Successfully set %s = %s", key, valueStr))
	return nil
}
