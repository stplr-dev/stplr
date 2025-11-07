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

package sources

import (
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"

	"go.stplr.dev/stplr/internal/config/common"
	"go.stplr.dev/stplr/pkg/types"
)

type DefaultConfig struct {
	k *koanf.Koanf
}

func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		k: koanf.New("."),
	}
}

func (DefaultConfig) Name() string {
	return "default"
}

func (c *DefaultConfig) Load() (*koanf.Koanf, error) {
	defaults := map[string]interface{}{
		common.ROOT_CMD:           "sudo",
		common.USE_ROOT_CMD:       true,
		common.PAGER_STYLE:        "native",
		common.IGNORE_PKG_UPDATES: []string{},
		common.LOG_LEVEL:          "info",
		common.AUTO_PULL:          true,
		common.REPO:               []types.Repo{},
	}
	if err := c.k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		panic(err)
	}
	return c.k, nil
}

var _ Source = (*DefaultConfig)(nil)
