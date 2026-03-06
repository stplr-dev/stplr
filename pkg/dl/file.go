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
	"errors"
	"fmt"
	"hash"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"golift.io/xtractr"

	"go.stplr.dev/stplr/internal/experimental/xtract"
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

// parseURLAndParams parses the URL and extracts special query parameters.
func (fd FileDownloader) parseURLAndParams(opts Options) (*url.URL, string, string, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, "", "", err
	}

	query := u.Query()

	name := query.Get("~name")
	query.Del("~name")

	archive := query.Get("~archive")
	query.Del("~archive")

	u.RawQuery = query.Encode()

	return u, name, archive, nil
}

var commonTimeoutDuration = 30 * time.Second

// getSource retrieves the source reader, size, and filename based on whether the URL is local or remote.
func (fd FileDownloader) getSource(ctx context.Context, u *url.URL, opts Options, name string) (io.ReadCloser, int64, string, error) {
	if IsLocalUrl(u) {
		localPath := filepath.Join(opts.LocalDir, u.Path)
		localFl, err := os.Open(localPath)
		if err != nil {
			return nil, 0, "", err
		}
		fi, err := localFl.Stat()
		if err != nil {
			localFl.Close()
			return nil, 0, "", err
		}
		size := fi.Size()
		if name == "" {
			name = fi.Name()
		}
		return localFl, size, name, nil
	} else {
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   commonTimeoutDuration,
					KeepAlive: commonTimeoutDuration,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: commonTimeoutDuration,
				ExpectContinueTimeout: 1 * time.Second,
				IdleConnTimeout:       3 * commonTimeoutDuration,
			},
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to create request: %w", err)
		}
		res, err := client.Do(req)
		if err != nil {
			return nil, 0, "", err
		}
		size := res.ContentLength
		if name == "" {
			name = getFilename(res)
		}
		return res.Body, size, name, nil
	}
}

// setupOutput configures the output writer, using a progress writer if applicable.
func (fd FileDownloader) setupOutput(ctx context.Context, fl *os.File, opts Options, size int64, name string) io.WriteCloser {
	if opts.Progress != nil {
		return NewProgressWriter(ctx, fl, size, name, opts.Progress)
	}
	return fl
}

// postProcess handles archive detection and extraction if post-processing is enabled.
func (fd FileDownloader) postProcess(path string, fl *os.File, name string, opts Options, postprocDisabled bool) (Type, string, error) {
	if postprocDisabled {
		return TypeFile, name, nil
	}

	_, err := xtract.ExtractArchive(path, opts.Destination)
	if errors.Is(err, xtractr.ErrUnknownArchiveType) {
		return TypeFile, "", nil
	}
	if err != nil {
		return 0, "", fmt.Errorf("failed to extract with new extractor: %w", err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		return 0, "", fmt.Errorf("failed to remove original archive: %w", err)
	}
	return TypeDir, "", nil
}

// Download downloads a file using HTTP. If the file is compressed in a supported format, it will be unpacked.
func (fd FileDownloader) Download(ctx context.Context, opts Options) (Type, string, error) {
	u, name, archive, err := fd.parseURLAndParams(opts)
	if err != nil {
		return 0, "", err
	}

	postprocDisabled := opts.PostprocDisabled || archive == "false"

	r, size, name, err := fd.getSource(ctx, u, opts, name)
	if err != nil {
		return 0, "", err
	}
	defer r.Close()

	path := filepath.Join(opts.Destination, name)
	fl, err := os.Create(path)
	if err != nil {
		return 0, "", err
	}

	out := fd.setupOutput(ctx, fl, opts, size, name)
	defer out.Close()

	var h hash.Hash
	w := io.Writer(out)
	if opts.Hash != nil {
		h, err = opts.NewHash()
		if err != nil {
			return 0, "", err
		}
		w = io.MultiWriter(out, h)
	}

	_, err = io.Copy(w, r)
	if err != nil {
		return 0, "", err
	}
	r.Close()

	if opts.Hash != nil {
		sum := h.Sum(nil)
		if !bytes.Equal(sum, opts.Hash) {
			return 0, "", ErrChecksumMismatch
		}
	}

	return fd.postProcess(path, fl, name, opts, postprocDisabled)
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
