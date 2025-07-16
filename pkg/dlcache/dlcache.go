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

package dlcache

import (
	"context"
	"os"
	"path/filepath"

	"go.stplr.dev/stplr/internal/config"
)

type Config interface {
	GetPaths() *config.Paths
}

type DownloadCache struct {
	cacheDir string
}

func New(cacheDir string) *DownloadCache {
	return &DownloadCache{cacheDir}
}

func (dc *DownloadCache) BasePath(ctx context.Context) string {
	return filepath.Join(dc.cacheDir, "dl")
}

// New creates a new directory with the given ID in the cache.
// If a directory with the same ID already exists,
// it will be deleted before creating a new one.
func (dc *DownloadCache) New(ctx context.Context, id string) (string, error) {
	h, err := hashID(id)
	if err != nil {
		return "", err
	}
	itemPath := filepath.Join(dc.BasePath(ctx), h)

	fi, err := os.Stat(itemPath)
	if err == nil || (fi != nil && !fi.IsDir()) {
		err = os.RemoveAll(itemPath)
		if err != nil {
			return "", err
		}
	}

	err = os.MkdirAll(itemPath, 0o755)
	if err != nil {
		return "", err
	}

	return itemPath, nil
}

// Get checks if an entry with the given ID
// already exists in the cache, and if so,
// returns the directory and true. If it
// does not exist, it returns an empty string
// and false.
func (dc *DownloadCache) Get(ctx context.Context, id string) (string, bool) {
	h, err := hashID(id)
	if err != nil {
		return "", false
	}
	itemPath := filepath.Join(dc.BasePath(ctx), h)

	_, err = os.Stat(itemPath)
	if err != nil {
		return "", false
	}

	return itemPath, true
}
