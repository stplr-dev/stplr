// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
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
	"errors"
	"fmt"
	"io"
	"os"

	ktoml "github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/pkg/types"
)

type SystemConfig struct {
	k   *koanf.Koanf
	cfg *types.Config

	Writer io.Writer
}

func NewSystemConfig() *SystemConfig {
	return &SystemConfig{
		k:   koanf.New("."),
		cfg: &types.Config{},
	}
}

func (c *SystemConfig) Name() string { return "system" }

func (c *SystemConfig) Load() (*koanf.Koanf, error) {
	if _, err := os.Stat(constants.SystemConfigPath); errors.Is(err, os.ErrNotExist) {
		return c.k, nil
	}

	if err := c.k.Load(file.Provider(constants.SystemConfigPath), ktoml.Parser()); err != nil {
		return c.k, err
	}

	return c.k, c.k.Unmarshal("", c.cfg)
}

func (c *SystemConfig) Set(key string, val any) error {
	return c.k.Set(key, val)
}

func (c *SystemConfig) Save() error {
	bytes, err := c.k.Marshal(ktoml.Parser())
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if c.Writer != nil {
		_, err = c.Writer.Write(bytes)
		return err
	} else {
		return errors.New("system config is not ready for write")
	}
}
