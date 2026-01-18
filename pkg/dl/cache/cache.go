// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package cache

import (
	"context"
	"errors"
)

// CacheID is the ID of the cache.
type CacheID string

// Metadata is the metadata of the cache.
//
// Metadata is an additional information about the URL.
// It is useful for separate same URL to different cache id.
type Metadata map[string]string

const (
	MetdataRestoreName = "core::restore-name"
)

var ErrEntryNotFound = errors.New("entry not found")

// Type is the type of the cache.
type Type uint8

const (
	TypeUnset Type = iota
	TypeFile
	TypeDir
)

func (t Type) String() string {
	switch t {
	case TypeFile:
		return "file"
	case TypeDir:
		return "dir"
	}
	return "<unknown>"
}

type Manifest struct {
	Type Type
	Name string
}

type CachedSource struct {
	Id       CacheID
	Manifest Manifest
}

type CachePutRequest struct {
	Id       CacheID
	Path     string
	URL      string
	Metadata Metadata
}

type DlCache interface {
	// Resolve resolves the URL to a cache ID.
	Resolve(ctx context.Context, url string, metadata Metadata) (CacheID, error)

	// Get returns the cached source for the given ID.
	Get(ctx context.Context, cid CacheID) (CachedSource, error)

	// Put puts the source to the cache.
	Put(ctx context.Context, req CachePutRequest) (CacheID, error)

	// Restore restores the cached source to the given directory.
	Restore(ctx context.Context, id CacheID, dir string) error

	// Delete deletes the cached source from the cache.
	Delete(ctx context.Context, id CacheID) error
}
