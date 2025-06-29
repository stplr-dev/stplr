// ALR - Any Linux Repository
// Copyright (C) 2025 The ALR Authors
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
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type EnvConfig struct {
	k *koanf.Koanf
}

func NewEnvConfig() *EnvConfig {
	return &EnvConfig{
		k: koanf.New("."),
	}
}

func (c *EnvConfig) koanf() *koanf.Koanf {
	return c.k
}

func (c *EnvConfig) Load() error {
	allowedKeys := map[string]struct{}{
		"ALR_LOG_LEVEL":   {},
		"ALR_PAGER_STYLE": {},
		"ALR_AUTO_PULL":   {},
	}
	err := c.k.Load(env.Provider("ALR_", ".", func(s string) string {
		_, ok := allowedKeys[s]
		if !ok {
			return ""
		}
		withoutPrefix := strings.TrimPrefix(s, "ALR_")
		lowered := strings.ToLower(withoutPrefix)
		dotted := strings.ReplaceAll(lowered, "__", ".")
		parts := strings.Split(dotted, ".")
		for i, part := range parts {
			if strings.Contains(part, "_") {
				parts[i] = toCamelCase(part)
			}
		}
		return strings.Join(parts, ".")
	}), nil)

	return err
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = cases.Title(language.Und, cases.NoLower).String(parts[i])
		}
	}
	return strings.Join(parts, "")
}
