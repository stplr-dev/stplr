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
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	urlMatchRegex       = regexp.MustCompile(`(magnet|torrent\+https?):.*`)
	ErrAria2NotFound    = errors.New("aria2 must be installed for torrent functionality")
	ErrDestinationEmpty = errors.New("the destination directory is empty")
)

type TorrentDownloader struct{}

// Name always returns "file"
func (TorrentDownloader) Name() string {
	return "torrent"
}

// MatchURL returns true if the URL is a magnet link
// or an http(s) link with a "torrent+" prefix
func (TorrentDownloader) MatchURL(u string) bool {
	return urlMatchRegex.MatchString(u)
}

// Download downloads a file over the BitTorrent protocol.
func (TorrentDownloader) Download(ctx context.Context, opts Options) (Type, string, error) {
	aria2Path, err := exec.LookPath("aria2c")
	if err != nil {
		return 0, "", ErrAria2NotFound
	}

	opts.URL = strings.TrimPrefix(opts.URL, "torrent+")

	cmd := exec.CommandContext(ctx, aria2Path, "--summary-interval=0", "--log-level=warn", "--seed-time=0", "--dir="+opts.Destination, opts.URL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return 0, "", fmt.Errorf("aria2c returned an error: %w", err)
	}

	err = removeTorrentFiles(opts.Destination)
	if err != nil {
		return 0, "", err
	}

	dlType, name, err := determineType(opts.Destination)
	if err != nil {
		return 0, "", err
	}

	err = VerifyHashFromLocal(name, opts)
	if err != nil {
		return 0, "", err
	}

	return dlType, name, nil
}

func removeTorrentFiles(path string) error {
	filePaths, err := filepath.Glob(filepath.Join(path, "*.torrent"))
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		err = os.Remove(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func determineType(path string) (Type, string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return 0, "", err
	}

	if len(files) > 1 {
		return TypeDir, "", nil
	} else if len(files) == 1 {
		if files[0].IsDir() {
			return TypeDir, files[0].Name(), nil
		} else {
			return TypeFile, files[0].Name(), nil
		}
	}

	return 0, "", ErrDestinationEmpty
}
