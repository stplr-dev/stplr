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

package local

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"go.stplr.dev/stplr/internal/osutils"
	"go.stplr.dev/stplr/pkg/dl/cache"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

const (
	metadataRepository         = "repository"
	metadataPackage            = "package"
	metadataVersion            = "version"
	metadataSFE249NewExtractor = "sfe_249_new_extractor"
)

type LocalCache struct {
	baseDir string
	engine  *xorm.Engine
}

func NewLocalCache(baseDir string) *LocalCache {
	return &LocalCache{baseDir: baseDir}
}

type cacheRecord struct {
	ID                 int64  `xorm:"pk autoincr"`
	Hash               string `xorm:"index"`
	Repo               string `xorm:"index"`
	Pkg                string `xorm:"index"`
	Ver                string `xorm:"index"`
	SFE249NewExtractor bool   `xorm:"'sfe_249_new_extractor'"`
	Name               string
	Type               cache.Type
}

func (c *LocalCache) Init() error {
	if err := c.Connect(); err != nil {
		return err
	}
	return c.engine.Sync(&cacheRecord{})
}

func (c *LocalCache) Connect() error {
	err := os.MkdirAll(c.baseDir, 0o755)
	if err != nil {
		return err
	}
	engine, err := xorm.NewEngine("sqlite", filepath.Join(c.baseDir, "db"))
	if err != nil {
		return err
	}
	c.engine = engine
	return nil
}

func (c *LocalCache) Reset(ctx context.Context) error {
	if c.engine == nil {
		if err := c.Connect(); err != nil {
			return err
		}
	}

	if err := c.engine.DropTables(&cacheRecord{}); err != nil {
		slog.Debug("failed to drop cacheRecord", "err", err)
	}

	if err := os.RemoveAll(c.baseDir); err != nil {
		return errors.New("unable to remove cache")
	}

	return nil
}

func (c *LocalCache) Resolve(ctx context.Context, url string, metadata cache.Metadata) (cache.CacheID, error) {
	m := ParseMetadata(metadata)
	repo, pkg, ver := m.Repository, m.Package, m.Version
	urlHash := hashUrl(url)

	exactMatch := &cacheRecord{
		Hash: urlHash,
		Repo: repo,
		Pkg:  pkg,
		Ver:  ver,
	}

	has, err := c.engine.Context(ctx).Get(exactMatch)
	if err != nil {
		return "", fmt.Errorf("failed to resolve cache ID: %w", err)
	}
	if has {
		return toCacheId(exactMatch.ID), nil
	}

	hashOnlyMatch := &cacheRecord{Hash: urlHash}
	has, err = c.engine.Context(ctx).Get(hashOnlyMatch)
	if err != nil {
		return "", fmt.Errorf("failed to resolve cache ID: %w", err)
	}
	if !has {
		return "", cache.ErrEntryNotFound
	}

	newRecord := &cacheRecord{
		Hash:               hashOnlyMatch.Hash,
		Type:               hashOnlyMatch.Type,
		SFE249NewExtractor: hashOnlyMatch.SFE249NewExtractor,
		Name:               hashOnlyMatch.Name,
		Repo:               repo,
		Pkg:                pkg,
		Ver:                ver,
	}
	if v, ok := metadata[cache.MetdataRestoreName]; ok {
		newRecord.Name = v
	}

	_, err = c.engine.Context(ctx).Insert(newRecord)
	if err != nil {
		return "", fmt.Errorf("failed to create cache record copy: %w", err)
	}

	return toCacheId(newRecord.ID), nil
}

func (c *LocalCache) Get(ctx context.Context, cid cache.CacheID) (cache.CachedSource, error) {
	src := cache.CachedSource{}

	id, err := fromCacheId(cid)
	if err != nil {
		return src, err
	}

	record := &cacheRecord{
		ID: id,
	}

	has, err := c.engine.Context(ctx).Get(record)
	if err != nil {
		return src, err
	}
	if !has {
		return src, cache.ErrEntryNotFound
	}

	src.Id = toCacheId(record.ID)
	src.Manifest.Name = record.Name
	src.Manifest.Type = record.Type
	src.Metadata = BuildMetadata(LocalCacheMetadata{
		Repository:         record.Repo,
		Package:            record.Pkg,
		Version:            record.Ver,
		SFE249NewExtractor: record.SFE249NewExtractor,
	})

	return src, nil
}

func (c *LocalCache) Put(ctx context.Context, req cache.CachePutRequest) (cache.CacheID, error) {
	hash := hashUrl(req.URL)
	record := &cacheRecord{
		Hash: hash,
	}

	if req.UsePathAsRoot {
		record.Name = ""
		record.Type = cache.TypeDir
	} else {
		record.Name = filepath.Base(req.Path)
		record.Type = typeFromPath(req.Path)
	}

	if req.Metadata != nil {
		m := ParseMetadata(req.Metadata)
		record.Repo = m.Repository
		record.Pkg = m.Package
		record.Ver = m.Version
		record.SFE249NewExtractor = m.SFE249NewExtractor
	}

	dest := c.recordEntry(*record)

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}

	// copy to record entry
	err := osutils.CopyDirOrFile(req.Path, dest)
	if err != nil {
		return "", err
	}

	_, err = c.engine.Context(ctx).Insert(record)
	if err != nil {
		return "", err
	}

	return toCacheId(record.ID), nil
}

func (c *LocalCache) Restore(ctx context.Context, id cache.CacheID, dir string) error {
	dbId, err := fromCacheId(id)
	if err != nil {
		return err
	}

	record := &cacheRecord{
		ID: dbId,
	}

	_, err = c.engine.Context(ctx).Get(record)
	if err != nil {
		return err
	}

	_, err = handleCache(c.recordEntry(*record), filepath.Join(dir, record.Name), record.Type)
	if err != nil {
		return err
	}

	return nil
}

func (c *LocalCache) Delete(ctx context.Context, cid cache.CacheID) error {
	id, err := fromCacheId(cid)
	if err != nil {
		return err
	}

	record := &cacheRecord{
		ID: id,
	}

	has, err := c.engine.Context(ctx).Get(record)
	if err != nil {
		return err
	}
	if !has {
		return nil // nothing to delete
	}

	_, err = c.engine.Context(ctx).Delete(&cacheRecord{ID: id})
	if err != nil {
		return err
	}

	count, err := c.engine.Context(ctx).Where("hash = ?", record.Hash).Count(&cacheRecord{})
	if err != nil {
		return err
	}

	if count == 0 {
		err := os.RemoveAll(c.recordDir(*record))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *LocalCache) CleanupOldPackageSources(ctx context.Context, repository, basePackage, version string) error {
	var records []cacheRecord

	err := c.engine.
		Where("repo = ? AND pkg = ? AND ver != ?", repository, basePackage, version).
		Find(&records)
	if err != nil {
		return err
	}

	for _, record := range records {
		err := c.Delete(ctx, toCacheId(record.ID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *LocalCache) recordEntry(r cacheRecord) string {
	return filepath.Join(c.recordDir(r), r.Name)
}

func (c *LocalCache) recordDir(r cacheRecord) string {
	return filepath.Join(c.baseDir, r.Hash, fmt.Sprintf("%d", btoi(r.SFE249NewExtractor)))
}

type LocalCacheMetadata struct {
	Repository         string
	Package            string
	Version            string
	SFE249NewExtractor bool
}

func BuildMetadata(m LocalCacheMetadata) cache.Metadata {
	return cache.Metadata{
		metadataRepository:         m.Repository,
		metadataPackage:            m.Package,
		metadataVersion:            m.Version,
		metadataSFE249NewExtractor: fmt.Sprintf("%t", m.SFE249NewExtractor),
	}
}

func ParseMetadata(metadata cache.Metadata) LocalCacheMetadata {
	sfe249, _ := strconv.ParseBool(metadata[metadataSFE249NewExtractor])
	return LocalCacheMetadata{
		Repository:         metadata[metadataRepository],
		Package:            metadata[metadataPackage],
		Version:            metadata[metadataVersion],
		SFE249NewExtractor: sfe249,
	}
}

func typeFromPath(path string) cache.Type {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return cache.TypeDir
	}
	return cache.TypeFile
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
