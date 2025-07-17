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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/leonelquinteros/gotext"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/exp/slices"
)

// Константа для имени файла манифеста кэша
const manifestFileName = ".alr_cache_manifest"

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

type DlCache interface {
	Get(context.Context, string) (string, bool)
	New(context.Context, string) (string, error)
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
	DlCache          DlCache
}

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
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
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

// Функция Download загружает файл или каталог с использованием указанных параметров
func Download(ctx context.Context, opts Options) (err error) {
	normalized, err := normalizeURL(opts.URL)
	if err != nil {
		return err
	}
	opts.URL = normalized

	d := getDownloader(opts.URL)

	if opts.CacheDisabled {
		_, _, err = d.Download(ctx, opts)
		return err
	}

	var t Type
	cacheDir, ok := opts.DlCache.Get(ctx, opts.URL)
	if ok {
		var updated bool
		if d, ok := d.(UpdatingDownloader); ok {
			slog.Info(
				gotext.Get("Source can be updated, updating if required"),
				"source", opts.Name,
				"downloader", d.Name(),
			)

			newOpts := opts
			newOpts.Destination = cacheDir

			updated, err = d.Update(newOpts)
			if err != nil {
				return err
			}
		}

		m, err := getManifest(cacheDir)
		if err == nil {
			t = m.Type

			dest := filepath.Join(opts.Destination, m.Name)
			ok, err := handleCache(cacheDir, dest, m.Name, t)
			if err != nil {
				return err
			}

			if ok && !updated {
				slog.Info(
					gotext.Get("Source found in cache and linked to destination"),
					"source", opts.Name,
					"type", t,
				)
				return nil
			} else if ok {
				slog.Info(
					gotext.Get("Source updated and linked to destination"),
					"source", opts.Name,
					"type", t,
				)
				return nil
			}
		} else {
			err = os.RemoveAll(cacheDir)
			if err != nil {
				return err
			}
		}
	}

	slog.Info(gotext.Get("Downloading source"), "source", opts.Name, "url", opts.URL, "downloader", d.Name())

	cacheDir, err = opts.DlCache.New(ctx, opts.URL)
	if err != nil {
		return err
	}

	newOpts := opts
	newOpts.Destination = cacheDir

	t, name, err := d.Download(ctx, newOpts)
	if err != nil {
		return err
	}

	err = writeManifest(cacheDir, Manifest{t, name})
	if err != nil {
		return err
	}

	dest := filepath.Join(opts.Destination, name)
	_, err = handleCache(cacheDir, dest, name, t)
	return err
}

// Функция writeManifest записывает манифест в указанный каталог кэша
func writeManifest(cacheDir string, m Manifest) error {
	fl, err := os.Create(filepath.Join(cacheDir, manifestFileName))
	if err != nil {
		return err
	}
	defer fl.Close()
	return msgpack.NewEncoder(fl).Encode(m)
}

// Функция getManifest считывает манифест из указанного каталога кэша
func getManifest(cacheDir string) (m Manifest, err error) {
	fl, err := os.Open(filepath.Join(cacheDir, manifestFileName))
	if err != nil {
		return Manifest{}, err
	}
	defer fl.Close()

	err = msgpack.NewDecoder(fl).Decode(&m)
	return
}

// Функция handleCache создает жесткие ссылки для файлов из каталога кэша в каталог назначения
func handleCache(cacheDir, dest, name string, t Type) (bool, error) {
	switch t {
	case TypeFile:
		cd, err := os.Open(cacheDir)
		if err != nil {
			return false, err
		}

		names, err := cd.Readdirnames(0)
		if err == io.EOF {
			break
		} else if err != nil {
			return false, err
		}

		cd.Close()

		if slices.Contains(names, name) {
			err = os.Link(filepath.Join(cacheDir, name), dest)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	case TypeDir:
		err := linkDir(cacheDir, dest)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Функция linkDir рекурсивно создает жесткие ссылки для файлов из каталога src в каталог dest
func linkDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == manifestFileName {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		newPath := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(newPath, info.Mode())
		}

		return os.Link(path, newPath)
	})
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
