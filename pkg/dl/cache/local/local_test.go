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
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.stplr.dev/stplr/pkg/dl/cache"
)

func TestNewLocalCache(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	assert.NotNil(t, localCache)
	assert.Equal(t, tempDir, localCache.baseDir)
	assert.Nil(t, localCache.engine)
}

func TestLocalCacheConnect(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Connect()
	require.NoError(t, err)
	assert.NotNil(t, localCache.engine)

	// Test that directory is created
	_, err = os.Stat(tempDir)
	assert.NoError(t, err)
}

func TestLocalCacheInit(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)
	assert.NotNil(t, localCache.engine)

	// Test that we can use the engine
	_, err = localCache.engine.IsTableExist(&cacheRecord{})
	assert.NoError(t, err)
}

func TestLocalCachePutAndGet(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Put the file in cache
	ctx := context.Background()
	req := cache.CachePutRequest{
		Path: testFile,
		URL:  "http://example.com/test.txt",
	}

	cacheID, err := localCache.Put(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, cacheID)

	// Get the cached source
	cachedSource, err := localCache.Get(ctx, cacheID)
	require.NoError(t, err)
	assert.Equal(t, cacheID, cachedSource.Id)
	assert.Equal(t, "test.txt", cachedSource.Manifest.Name)
	assert.Equal(t, cache.TypeFile, cachedSource.Manifest.Type)
}

func TestLocalCachePutAndGetDirectory(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)

	// Create a test directory with files
	testDir := filepath.Join(tempDir, "testdir")
	err = os.Mkdir(testDir, 0o755)
	require.NoError(t, err)

	testFile := filepath.Join(testDir, "file.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Put the directory in cache
	ctx := context.Background()
	req := cache.CachePutRequest{
		Path: testDir,
		URL:  "http://example.com/testdir",
	}

	cacheID, err := localCache.Put(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, cacheID)

	// Get the cached source
	cachedSource, err := localCache.Get(ctx, cacheID)
	require.NoError(t, err)
	assert.Equal(t, cacheID, cachedSource.Id)
	assert.Equal(t, "testdir", cachedSource.Manifest.Name)
	assert.Equal(t, cache.TypeDir, cachedSource.Manifest.Type)
}

func TestLocalCacheResolve(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)

	ctx := context.Background()
	testURL := "http://example.com/test.txt"

	// Try to resolve non-existent entry
	metadata := BuildMetadata(LocalCacheMetadata{
		Repository: "test-repo",
		Package:    "test-package",
		Version:    "1.0.0",
	})

	_, err = localCache.Resolve(ctx, testURL, metadata)
	assert.Error(t, err)
	assert.Equal(t, cache.ErrEntryNotFound, err)
}

func TestLocalCacheDelete(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Put the file in cache
	ctx := context.Background()
	req := cache.CachePutRequest{
		Path: testFile,
		URL:  "http://example.com/test.txt",
	}

	cacheID, err := localCache.Put(ctx, req)
	require.NoError(t, err)

	// Delete the cached entry
	err = localCache.Delete(ctx, cacheID)
	require.NoError(t, err)

	// Try to get the deleted entry
	_, err = localCache.Get(ctx, cacheID)
	assert.Error(t, err)
	assert.Equal(t, cache.ErrEntryNotFound, err)
}

func TestLocalCacheRestore(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Put the file in cache
	ctx := context.Background()
	req := cache.CachePutRequest{
		Path: testFile,
		URL:  "http://example.com/test.txt",
	}

	cacheID, err := localCache.Put(ctx, req)
	require.NoError(t, err)

	// Create restore directory
	restoreDir := filepath.Join(tempDir, "restore")
	err = os.Mkdir(restoreDir, 0o755)
	require.NoError(t, err)

	// Restore the cached file
	err = localCache.Restore(ctx, cacheID, restoreDir)
	require.NoError(t, err)

	// Check that file was restored
	restoredFile := filepath.Join(restoreDir, "test.txt")
	content, err := os.ReadFile(restoredFile)
	require.NoError(t, err)
	assert.Equal(t, "test content", string(content))
}

func TestLocalCacheCleanupOldPackageSources(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)

	// Create test files
	testFile1 := filepath.Join(tempDir, "test1.txt")
	err = os.WriteFile(testFile1, []byte("test content 1"), 0o644)
	require.NoError(t, err)

	testFile2 := filepath.Join(tempDir, "test2.txt")
	err = os.WriteFile(testFile2, []byte("test content 2"), 0o644)
	require.NoError(t, err)

	ctx := context.Background()

	// Put first file in cache with version 1.0.0
	req1 := cache.CachePutRequest{
		Path: testFile1,
		URL:  "http://example.com/test.txt",
		Metadata: BuildMetadata(LocalCacheMetadata{
			Repository: "test-repo",
			Package:    "test-package",
			Version:    "1.0.0",
		}),
	}

	cacheID1, err := localCache.Put(ctx, req1)
	require.NoError(t, err)

	// Put second file in cache with version 2.0.0
	req2 := cache.CachePutRequest{
		Path: testFile2,
		URL:  "http://example.com/test.txt",
		Metadata: BuildMetadata(LocalCacheMetadata{
			Repository: "test-repo",
			Package:    "test-package",
			Version:    "2.0.0",
		}),
	}

	cacheID2, err := localCache.Put(ctx, req2)
	require.NoError(t, err)

	// Cleanup old package sources (should remove version 1.0.0)
	err = localCache.CleanupOldPackageSources(ctx, "test-repo", "test-package", "2.0.0")
	require.NoError(t, err)

	// Check that old version is deleted
	_, err = localCache.Get(ctx, cacheID1)
	assert.Error(t, err)
	assert.Equal(t, cache.ErrEntryNotFound, err)

	// Check that new version still exists
	_, err = localCache.Get(ctx, cacheID2)
	assert.NoError(t, err)
}

func TestLocalCacheReset(t *testing.T) {
	tempDir := t.TempDir()
	localCache := NewLocalCache(tempDir)

	err := localCache.Init()
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0o644)
	require.NoError(t, err)

	// Put the file in cache
	ctx := context.Background()
	req := cache.CachePutRequest{
		Path: testFile,
		URL:  "http://example.com/test.txt",
	}

	_, err = localCache.Put(ctx, req)
	require.NoError(t, err)

	// Reset the cache
	err = localCache.Reset(ctx)
	require.NoError(t, err)

	// Check that cache directory is removed
	_, err = os.Stat(tempDir)
	require.ErrorIs(t, err, fs.ErrNotExist)
	// Note: Reset removes the entire directory, but TempDir might still exist
	// The important thing is that the database connection is closed
	assert.NoError(t, localCache.Connect())
}

func TestBuildAndParseMetadata(t *testing.T) {
	metadata := LocalCacheMetadata{
		Repository: "test-repo",
		Package:    "test-package",
		Version:    "1.0.0",
	}

	cacheMetadata := BuildMetadata(metadata)
	parsedMetadata := ParseMetadata(cacheMetadata)

	assert.Equal(t, metadata, parsedMetadata)
}

func TestTypeFromPath(t *testing.T) {
	tempDir := t.TempDir()

	// Test with file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0o644)
	require.NoError(t, err)

	fileType := typeFromPath(testFile)
	assert.Equal(t, cache.TypeFile, fileType)

	// Test with directory
	testDir := filepath.Join(tempDir, "testdir")
	err = os.Mkdir(testDir, 0o755)
	require.NoError(t, err)

	dirType := typeFromPath(testDir)
	assert.Equal(t, cache.TypeDir, dirType)

	// Test with non-existent path
	nonExistent := filepath.Join(tempDir, "nonexistent")
	nonExistentType := typeFromPath(nonExistent)
	assert.Equal(t, cache.TypeFile, nonExistentType)
}
