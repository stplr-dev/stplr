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

package config

import (
	"encoding/json"
	"errors"
	"fmt"
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
}

func NewSystemConfig() *SystemConfig {
	return &SystemConfig{
		k:   koanf.New("."),
		cfg: &types.Config{},
	}
}

func (c *SystemConfig) koanf() *koanf.Koanf {
	return c.k
}

func (c *SystemConfig) Load() error {
	if _, err := os.Stat(constants.SystemConfigPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err := c.k.Load(file.Provider(constants.SystemConfigPath), ktoml.Parser()); err != nil {
		return err
	}

	return c.k.Unmarshal("", c.cfg)
}

func (c *SystemConfig) Save() error {
	bytes, err := c.k.Marshal(ktoml.Parser())
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	file, err := os.Create(constants.SystemConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err := file.Write(bytes); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync config: %w", err)
	}

	return nil
}

func (c *SystemConfig) SetRootCmd(v string) {
	err := c.k.Set("rootCmd", v)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetUseRootCmd(v bool) {
	err := c.k.Set("useRootCmd", v)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetPagerStyle(v string) {
	err := c.k.Set("pagerStyle", v)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetIgnorePkgUpdates(v []string) {
	err := c.k.Set("ignorePkgUpdates", v)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetAutoPull(v bool) {
	err := c.k.Set("autoPull", v)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetLogLevel(v string) {
	err := c.k.Set("logLevel", v)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetRepos(v []types.Repo) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var m []interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		panic(err)
	}
	err = c.k.Set("repo", m)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetForbidSkipInChecksums(v bool) {
	err := c.k.Set("forbidSkipInChecksums", v)
	if err != nil {
		panic(err)
	}
}

func (c *SystemConfig) SetForbidBuildCommand(v bool) {
	err := c.k.Set("forbidBuildCommand", v)
	if err != nil {
		panic(err)
	}
}
