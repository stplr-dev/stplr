// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/knadh/koanf/v2"

	"go.stplr.dev/stplr/internal/config/common"
	"go.stplr.dev/stplr/internal/config/internal/sources"
	"go.stplr.dev/stplr/internal/config/savers"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/pkg/types"
)

type ALRConfig struct {
	cfg   *types.Config
	paths *Paths

	srcs []sources.Source
}

type Option func(*ALRConfig)

func WithSystemConfigWriter(w savers.SystemConfigWriterExecutor) Option {
	return func(cfg *ALRConfig) {
		for _, src := range cfg.srcs {
			if v, ok := src.(*sources.SystemConfig); ok {
				v.Writer = w
				return
			}
		}
	}
}

func New(opts ...Option) *ALRConfig {
	cfg := &ALRConfig{
		srcs: []sources.Source{
			sources.NewDefaultConfig(),
			sources.NewSystemConfig(),
			sources.NewEnvConfig(),
		},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func (c *ALRConfig) Load() error {
	merged := koanf.New(".")

	for _, src := range c.srcs {
		k, err := src.Load()
		if err != nil {
			return fmt.Errorf("failed to load %s config: %w", src.Name(), err)
		}
		if err := merged.Merge(k); err != nil {
			return fmt.Errorf("failed to merge %s config: %w", src.Name(), err)
		}
	}

	cfg := &types.Config{}
	if err := merged.Unmarshal("", cfg); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	c.cfg = cfg
	c.paths = &Paths{}
	c.paths.UserConfigPath = constants.SystemConfigPath
	c.paths.CacheDir = constants.SystemCachePath
	c.paths.RepoDir = filepath.Join(c.paths.CacheDir, "repo")
	c.paths.PkgsDir = filepath.Join(c.paths.CacheDir, "pkgs")
	c.paths.DBPath = filepath.Join(c.paths.CacheDir, "db")

	return nil
}

func (c *ALRConfig) SetTo(level, key string, value any) error {
	for _, src := range c.srcs {
		if src.Name() != level {
			continue
		}

		setter, ok := src.(sources.Setter)
		if !ok {
			return fmt.Errorf("%s config is not writable", src.Name())
		}

		return setter.Set(key, value)
	}
	return fmt.Errorf("unknown config level: %s", level)
}

// TODO: remove
func (c *ALRConfig) SetRepos(v []types.Repo) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var m []interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		panic(err)
	}
	if err := c.SetTo(common.SOURCE_SYSTEM, common.REPO, m); err != nil {
		panic(err)
	}
}

func (c *ALRConfig) Save(level string) error {
	for _, src := range c.srcs {
		if src.Name() != level {
			continue
		}

		saver, ok := src.(sources.Saver)
		if !ok {
			return nil
		}

		return saver.Save()
	}
	return nil
}

func (c *ALRConfig) SetToAndSave(level, key string, value any) error {
	if err := c.SetTo(level, key, value); err != nil {
		return err
	}

	return c.Save(level)
}

func (c *ALRConfig) ToYAML() (string, error) {
	data, err := yaml.Marshal(c.cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *ALRConfig) RootCmd() string                  { return c.cfg.RootCmd }
func (c *ALRConfig) PagerStyle() string               { return c.cfg.PagerStyle }
func (c *ALRConfig) AutoPull() bool                   { return c.cfg.AutoPull }
func (c *ALRConfig) Repos() []types.Repo              { return c.cfg.Repos }
func (c *ALRConfig) IgnorePkgUpdates() []string       { return c.cfg.IgnorePkgUpdates }
func (c *ALRConfig) LogLevel() string                 { return c.cfg.LogLevel }
func (c *ALRConfig) UseRootCmd() bool                 { return c.cfg.UseRootCmd }
func (c *ALRConfig) FirejailExclude() []string        { return c.cfg.FirejailExclude }
func (c *ALRConfig) HideFirejailExcludeWarning() bool { return c.cfg.HideFirejailExcludeWarning }
func (c *ALRConfig) ForbidSkipInChecksums() bool      { return c.cfg.ForbidSkipInChecksums }
func (c *ALRConfig) ForbidBuildCommand() bool         { return c.cfg.ForbidBuildCommand }
func (c *ALRConfig) GetPaths() *Paths                 { return c.paths }

// TODO: refactor
func PatchToUserDirs(c *ALRConfig) error {
	paths := c.GetPaths()
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	paths.CacheDir = filepath.Join(userCacheDir, "stplr")
	paths.PkgsDir = filepath.Join(paths.CacheDir, "pkgs")

	return nil
}

func AllowedKeys() []string {
	return []string{
		common.ROOT_CMD,
		common.PAGER_STYLE,
		common.LOG_LEVEL,
		common.USE_ROOT_CMD,
		common.AUTO_PULL,
		common.IGNORE_PKG_UPDATES,
		common.FORBID_SKIP_IN_CHECKSUMS,
		common.FORBID_BUILD_COMMAND,
		common.FIREJAIL_EXCLUDE,
		common.HIDE_FIREJAIL_EXCLUDE_WARNING,
	}
}

func ConvertValue(key, v string) (any, error) {
	switch key {
	case common.AUTO_PULL, common.USE_ROOT_CMD,
		common.FORBID_SKIP_IN_CHECKSUMS, common.FORBID_BUILD_COMMAND, common.HIDE_FIREJAIL_EXCLUDE_WARNING:
		val, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("expected boolean value, got: %s", v)
		}
		return val, nil

	case common.IGNORE_PKG_UPDATES, common.FIREJAIL_EXCLUDE:
		if v == "" {
			return []string{}, nil
		}
		updates := strings.Split(v, ",")
		for i := range updates {
			updates[i] = strings.TrimSpace(updates[i])
		}
		return updates, nil

	case common.ROOT_CMD, common.PAGER_STYLE, common.LOG_LEVEL:
		return v, nil

	default:
		return nil, fmt.Errorf("unknown config key: %s", key)
	}
}
