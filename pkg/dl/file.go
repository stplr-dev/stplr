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

package dl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v4"
)

// FileDownloader загружает файлы с использованием HTTP
type FileDownloader struct{}

// Name всегда возвращает "file"
func (FileDownloader) Name() string {
	return "file"
}

// MatchURL всегда возвращает true, так как FileDownloader
// используется как резерв, если ничего другого не соответствует
func (FileDownloader) MatchURL(string) bool {
	return true
}

func IsLocalUrl(u *url.URL) bool { return u.Scheme == "local" }

// Download загружает файл с использованием HTTP. Если файл
// сжат в поддерживаемом формате, он будет распакован
func (FileDownloader) Download(ctx context.Context, opts Options) (Type, string, error) {
	// Разбор URL
	u, err := url.Parse(opts.URL)
	if err != nil {
		return 0, "", err
	}

	// Получение параметров запроса
	query := u.Query()

	// Получение имени файла из параметров запроса
	name := query.Get("~name")
	query.Del("~name")

	// Получение параметра архивации
	archive := query.Get("~archive")
	query.Del("~archive")

	// Кодирование измененных параметров запроса обратно в URL
	u.RawQuery = query.Encode()

	var r io.ReadCloser
	var size int64

	if IsLocalUrl(u) {
		localFl, err := os.Open(filepath.Join(opts.LocalDir, u.Path))
		if err != nil {
			return 0, "", err
		}
		fi, err := localFl.Stat()
		if err != nil {
			return 0, "", err
		}
		size = fi.Size()
		if name == "" {
			name = fi.Name()
		}
		r = localFl
	} else {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return 0, "", fmt.Errorf("failed to create request: %w", err)
		}
		// Выполнение HTTP GET запроса
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, "", err
		}
		size = res.ContentLength
		if name == "" {
			name = getFilename(res)
		}
		r = res.Body
	}
	defer r.Close()

	postprocDisabled := opts.PostprocDisabled || archive == "false"

	path := filepath.Join(opts.Destination, name)
	fl, err := os.Create(path)
	if err != nil {
		return 0, "", err
	}

	var out io.WriteCloser
	// Настройка индикатора прогресса
	if opts.Progress != nil {
		out = NewProgressWriter(fl, size, name, opts.Progress)
	} else {
		out = fl
	}
	defer out.Close()

	h, err := opts.NewHash()
	if err != nil {
		return 0, "", err
	}

	var w io.Writer
	// Настройка MultiWriter для записи в файл, хеш и индикатор прогресса
	if opts.Hash != nil {
		w = io.MultiWriter(h, out)
	} else {
		w = io.MultiWriter(out)
	}

	// Копирование содержимого из источника в файл назначения
	_, err = io.Copy(w, r)
	if err != nil {
		return 0, "", err
	}
	r.Close()

	// Проверка контрольной суммы
	if opts.Hash != nil {
		sum := h.Sum(nil)
		if !bytes.Equal(sum, opts.Hash) {
			return 0, "", ErrChecksumMismatch
		}
	}

	// Проверка необходимости постобработки
	if postprocDisabled {
		return TypeFile, name, nil
	}

	_, err = fl.Seek(0, io.SeekStart)
	if err != nil {
		return 0, "", err
	}

	// Идентификация формата архива
	format, ar, err := archiver.Identify(name, fl)
	if err == archiver.ErrNoMatch {
		return TypeFile, name, nil
	} else if err != nil {
		return 0, "", err
	}

	// Распаковка архива
	err = extractFile(ar, format, name, opts)
	if err != nil {
		return 0, "", err
	}

	// Удаление исходного архива
	err = os.Remove(path)
	return TypeDir, "", err
}

// extractFile извлекает архив или распаковывает файл
func extractFile(r io.Reader, format archiver.Format, name string, opts Options) (err error) {
	fname := format.Name()

	// Проверка типа формата архива
	switch format := format.(type) {
	case archiver.Extractor:
		// Извлечение файлов из архива
		err = format.Extract(context.Background(), r, nil, func(ctx context.Context, f archiver.File) error {
			fr, err := f.Open()
			if err != nil {
				return err
			}
			defer fr.Close()
			fi, err := f.Stat()
			if err != nil {
				return err
			}
			fm := fi.Mode()

			path := filepath.Join(opts.Destination, f.NameInArchive)

			err = os.MkdirAll(filepath.Dir(path), 0o755)
			if err != nil {
				return err
			}

			if f.IsDir() {
				err = os.MkdirAll(path, 0o755)
				if err != nil {
					return err
				}
			} else {
				outFl, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fm.Perm())
				if err != nil {
					return err
				}
				defer outFl.Close()

				_, err = io.Copy(outFl, fr)
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	case archiver.Decompressor:
		// Распаковка сжатого файла
		rc, err := format.OpenReader(r)
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(opts.Destination, name)
		path = strings.TrimSuffix(path, fname)

		outFl, err := os.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(outFl, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

// getFilename пытается разобрать заголовок Content-Disposition
// HTTP-ответа и извлечь имя файла. Если заголовок отсутствует,
// используется последний элемент пути.
func getFilename(res *http.Response) (name string) {
	_, params, err := mime.ParseMediaType(res.Header.Get("Content-Disposition"))
	if err != nil {
		return path.Base(res.Request.URL.Path)
	}
	if filename, ok := params["filename"]; ok {
		return filename
	} else {
		return path.Base(res.Request.URL.Path)
	}
}
