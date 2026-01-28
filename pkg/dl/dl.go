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

// Пакет dl содержит абстракции для загрузки файлов и каталогов
// из различных источников.
package dl

import (
	"context"
	"crypto/md5"  //gosec:disable G501 -- Allowed hash for files
	"crypto/sha1" //gosec:disable G505 -- Allowed hash for files
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/leonelquinteros/gotext"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"

	"go.stplr.dev/stplr/internal/app/output"

	"go.stplr.dev/stplr/pkg/dl/cache"
	"go.stplr.dev/stplr/pkg/dl/cache/local"
)

// Объявление ошибок для несоответствия контрольной суммы и отсутствия алгоритма хеширования
var (
	ErrChecksumMismatch = errors.New("dl: checksums did not match")
	ErrNoSuchHashAlgo   = errors.New("dl: invalid hashing algorithm")
)

// Массив доступных загрузчиков в порядке их проверки
var Downloaders = []Downloader{
	&GitDownloader{},
	TorrentDownloader{},
	FileDownloader{},
}

// Тип данных, представляющий тип загрузки (файл или каталог)
type Type uint8

// Объявление констант для типов загрузки
const (
	TypeFile Type = iota
	TypeDir
)

// Метод для получения строки, представляющей тип загрузки
func (t Type) String() string {
	switch t {
	case TypeFile:
		return "file"
	case TypeDir:
		return "dir"
	}
	return "<unknown>"
}

// Структура Options содержит параметры для загрузки файлов и каталогов
type Options struct {
	Hash             []byte
	HashAlgorithm    string
	Name             string
	URL              string
	Destination      string
	CacheDisabled    bool
	PostprocDisabled bool
	Progress         io.Writer
	LocalDir         string
	DlCache          cache.DlCache
	CacheMetadata    cache.Metadata
	Output           output.Output
	NewExtractor     bool
}

var _ cache.DlCache = new(local.LocalCache)

// Метод для создания нового хеша на основе указанного алгоритма хеширования
func (opts Options) NewHash() (hash.Hash, error) {
	switch opts.HashAlgorithm {
	case "", "sha256":
		return sha256.New(), nil
	case "sha224":
		return sha256.New224(), nil
	case "sha512":
		return sha512.New(), nil
	case "sha384":
		return sha512.New384(), nil
	case "sha1":
		return sha1.New(), nil //gosec:disable G401 -- Allowed hash for files
	case "md5":
		return md5.New(), nil //gosec:disable G401 -- Allowed hash for files
	case "blake2s-128":
		return blake2s.New256(nil)
	case "blake2s-256":
		return blake2s.New256(nil)
	case "blake2b-256":
		return blake2b.New(32, nil)
	case "blake2b-512":
		return blake2b.New(64, nil)
	default:
		return nil, fmt.Errorf("%w: %s", ErrNoSuchHashAlgo, opts.HashAlgorithm)
	}
}

// Структура Manifest хранит информацию о типе и имени загруженного файла или каталога
type Manifest struct {
	Type Type
	Name string
}

// Интерфейс Downloader для реализации различных загрузчиков
type Downloader interface {
	Name() string
	MatchURL(string) bool
	Download(context.Context, Options) (Type, string, error)
}

// Интерфейс UpdatingDownloader расширяет Downloader методом Update
type UpdatingDownloader interface {
	Downloader
	Update(Options) (bool, error)
}

// Download handles downloading a file or directory with the given options.
func Download(ctx context.Context, opts Options) (err error) {
	normalized, err := normalizeURL(opts.URL)
	if err != nil {
		return err
	}
	opts.URL = normalized

	d := getDownloader(opts.URL)

	u, err := url.Parse(opts.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	query := u.Query()
	name := query.Get("~name")
	if name != "" {
		opts.Name = name
		if opts.CacheMetadata != nil {
			opts.CacheMetadata[cache.MetdataRestoreName] = name
		}
	}

	if opts.CacheDisabled {
		_, _, err = d.Download(ctx, opts)
		return err
	}

	return downloadWithCache(ctx, opts, d)
}

// downloadWithCache handles downloading with cache logic.
func downloadWithCache(ctx context.Context, opts Options, d Downloader) error {
	id, err := opts.DlCache.Resolve(ctx, opts.URL, opts.CacheMetadata)
	if err != nil {
		if errors.Is(err, cache.ErrEntryNotFound) {
			return performDownload(ctx, opts, d)
		}
		return fmt.Errorf("failed to resolve cache id: %w", err)
	}

	return handleCachedSource(ctx, opts, d, id)
}

// handleCachedSource processes a cached source, updating it if necessary.
func handleCachedSource(ctx context.Context, opts Options, d Downloader, cid cache.CacheID) error {
	s, err := opts.DlCache.Get(ctx, cid)
	if err != nil {
		if errors.Is(err, cache.ErrEntryNotFound) {
			return performDownload(ctx, opts, d)
		}
		return err
	}

	if local.ParseMetadata(s.Metadata).SFE249NewExtractor != opts.NewExtractor {
		err := opts.DlCache.Delete(ctx, cid)
		if err != nil {
			return err
		}

		return performDownload(ctx, opts, d)
	}

	updated, err := updateSourceIfNeeded(ctx, opts, d, cid)
	if err != nil {
		return err
	}

	err = opts.DlCache.Restore(ctx, cid, opts.Destination)
	if err != nil {
		if errors.Is(err, cache.ErrEntryNotFound) {
			return performDownload(ctx, opts, d)
		}
		return err
	}
	logCacheHit(opts, updated)
	return nil
}

func updateSourceIfNeeded(ctx context.Context, opts Options, d Downloader, cid cache.CacheID) (bool, error) {
	if updater, ok := d.(UpdatingDownloader); ok {
		if opts.Output != nil {
			opts.Output.Info(gotext.Get(
				"Source %q can be updated using %s — updating if required",
				opts.Name, d.Name(),
			))
		}

		tmpDir, err := os.MkdirTemp(opts.Destination, ".tmp.updater-*")
		if err != nil {
			return false, err
		}
		defer os.RemoveAll(tmpDir)

		err = opts.DlCache.Restore(ctx, cid, tmpDir)
		if err != nil {
			return false, err
		}

		newOpts := opts
		newOpts.Destination = tmpDir
		updated, err := updater.Update(newOpts)
		if err != nil {
			return false, fmt.Errorf("failed to update source: %w", err)
		}

		if !updated {
			return false, nil
		}

		src, err := opts.DlCache.Get(ctx, cid)
		if err != nil {
			return false, err
		}

		newMetadata := cache.Metadata{}
		maps.Copy(src.Metadata, newMetadata)
		maps.Copy(opts.CacheMetadata, newMetadata)

		if _, err = opts.DlCache.Put(ctx, cache.CachePutRequest{
			Id:       src.Id,
			Path:     filepath.Join(tmpDir, src.Manifest.Name),
			Metadata: newMetadata,
		}); err != nil {
			return false, err
		}

		return true, nil
	}
	return false, nil
}

// performDownload executes the download and writes to the cache.
func performDownload(ctx context.Context, opts Options, d Downloader) error {
	slog.Debug(
		"downloading source",
		"source", opts.Name,
		"url", opts.URL,
		"downloader", d.Name(),
	)
	if opts.Output != nil {
		opts.Output.Info("%s", gotext.Get(
			"Downloading source %s from %s using %s downloader",
			opts.Name, opts.URL, d.Name(),
		))
	}

	tmpDir, err := os.MkdirTemp(opts.Destination, ".dl-cache-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	newOpts := opts
	newOpts.Destination = tmpDir

	_, name, err := d.Download(ctx, newOpts)
	if err != nil {
		return err
	}

	creq := cache.CachePutRequest{
		URL:      opts.URL,
		Metadata: opts.CacheMetadata,
	}

	if name != "" {
		creq.Path = filepath.Join(tmpDir, name)
	} else {
		creq.Path = tmpDir
		creq.UsePathAsRoot = true
	}

	cid, err := opts.DlCache.Put(ctx, creq)
	if err != nil {
		return fmt.Errorf("failed to put cache: %w", err)
	}
	os.RemoveAll(tmpDir)

	err = opts.DlCache.Restore(ctx, cid, opts.Destination)
	if err != nil {
		return err
	}

	return nil
}

// logCacheHit logs when a cached source is used or updated.
func logCacheHit(opts Options, updated bool) {
	if opts.Output == nil {
		return
	}

	var msg string
	if updated {
		msg = gotext.Get("Source %q was updated and linked to destination", opts.Name)
	} else {
		msg = gotext.Get("Source %q found in cache and linked to destination", opts.Name)
	}
	opts.Output.Info("%s", msg)
}

// Функция getDownloader возвращает загрузчик, соответствующий URL
func getDownloader(u string) Downloader {
	for _, d := range Downloaders {
		if d.MatchURL(u) {
			return d
		}
	}
	return nil
}

// Функция normalizeURL нормализует строку URL, чтобы незначительные различия не изменяли хеш
func normalizeURL(u string) (string, error) {
	const normalizationFlags = purell.FlagRemoveTrailingSlash |
		purell.FlagRemoveDefaultPort |
		purell.FlagLowercaseHost |
		purell.FlagLowercaseScheme |
		purell.FlagRemoveDuplicateSlashes |
		purell.FlagRemoveFragment |
		purell.FlagRemoveUnnecessaryHostDots |
		purell.FlagSortQuery |
		purell.FlagDecodeHexHost |
		purell.FlagDecodeOctalHost |
		purell.FlagDecodeUnnecessaryEscapes |
		purell.FlagRemoveEmptyPortSeparator

	u, err := purell.NormalizeURLString(u, normalizationFlags)
	if err != nil {
		return "", err
	}

	// Исправление URL-адресов magnet после нормализации
	u = strings.Replace(u, "magnet://", "magnet:", 1)
	return u, nil
}
