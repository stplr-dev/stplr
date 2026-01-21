// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
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

package build

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/pkg/dl"
	"go.stplr.dev/stplr/pkg/dl/cache/local"
)

type LocalSourceDownloader struct {
	cfg        commonbuild.Config
	localCache *local.LocalCache
	out        output.Output
}

func NewLocalSourceDownloader(cfg commonbuild.Config, out output.Output) *LocalSourceDownloader {
	localCache := local.NewLocalCache(filepath.Join(cfg.GetPaths().CacheDir, "dl"))
	return &LocalSourceDownloader{
		cfg,
		localCache,
		out,
	}
}

func (s *LocalSourceDownloader) DownloadSources(
	ctx context.Context,
	input *commonbuild.BuildInput,
	repo string,
	basePkg string,
	version string,
	si SourcesInput,
) error {
	if err := s.localCache.Init(); err != nil {
		return err
	}

	for i, src := range si.Sources {
		opts := dl.Options{
			Name:         fmt.Sprintf("[%d]", i),
			URL:          src,
			Progress:     os.Stderr,
			Destination:  commonbuild.GetSrcDir(s.cfg, basePkg),
			LocalDir:     commonbuild.GetScriptDir(input.Script),
			Output:       s.out,
			DlCache:      s.localCache,
			NewExtractor: si.NewExtractor,
			CacheMetadata: local.BuildMetadata(
				local.LocalCacheMetadata{
					Repository:         repo,
					Package:            basePkg,
					Version:            version,
					SFE249NewExtractor: si.NewExtractor,
				},
			),
		}

		err := s.setHashFromChecksum(si.Checksums[i], &opts)
		if err != nil {
			return err
		}

		err = dl.Download(ctx, opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *LocalSourceDownloader) RemoveOldSourcesFromCache(ctx context.Context, repo, basePkg, version string) error {
	return s.localCache.CleanupOldPackageSources(ctx, repo, basePkg, version)
}

func IsSkipChecksum(checksum string) bool {
	return strings.EqualFold(checksum, "SKIP")
}

func (s *LocalSourceDownloader) setHashFromChecksum(checksum string, opts *dl.Options) error {
	if !IsSkipChecksum(checksum) {
		// If the checksum contains a colon, use the part before the colon
		// as the algorithm, and the part after as the actual checksum.
		// Otherwise, use SHA-256 by default with the whole string as the checksum.
		algo, hashData, ok := strings.Cut(checksum, ":")
		if ok {
			checksum, err := hex.DecodeString(hashData)
			if err != nil {
				return err
			}
			opts.Hash = checksum
			opts.HashAlgorithm = algo
		} else {
			checksum, err := hex.DecodeString(checksum)
			if err != nil {
				return err
			}
			opts.Hash = checksum
		}
	}
	return nil
}
