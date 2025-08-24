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

package handlers

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"mvdan.cc/sh/v3/interp"

	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/shutils/handlers/filter"
)

type options struct {
	Filter          filter.Predicate
	BlockedPrefixes []string
	PathRedirects   map[string]string
}

type Option func(*options)

func WithFilter(f filter.Predicate) Option {
	return func(o *options) {
		o.Filter = f
	}
}

func WithPathRedirect(from, to string) Option {
	return func(o *options) {
		if o.PathRedirects == nil {
			o.PathRedirects = make(map[string]string)
		}
		o.PathRedirects[filepath.Clean(from)] = filepath.Clean(to)
	}
}

func hasPathPrefix(path, prefix string) bool {
	path = filepath.Clean(path)
	prefix = filepath.Clean(prefix)

	if path == prefix {
		return true
	}

	if strings.HasPrefix(path, prefix+string(os.PathSeparator)) {
		return true
	}

	return false
}

func RestrictSandbox(allowedList ...string) filter.Predicate {
	blacklisted := []string{
		constants.SystemCachePath,
		constants.SocketDirPath,
	}
	return func(path string) bool {
		path = filepath.Clean(path)
		ok := true
		for _, blacklistedPath := range blacklisted {
			if hasPathPrefix(path, blacklistedPath) {
				ok = false
			}
		}
		if ok {
			return true
		}
		for _, allowed := range allowedList {
			allowed = filepath.Clean(allowed)
			if hasPathPrefix(path, allowed) || hasPathPrefix(allowed, path) {
				return true
			}
		}
		return false
	}
}

func fsFromOpts(opts options) afero.Fs {
	baseFs := afero.NewOsFs()
	if opts.Filter != nil {
		return filter.NewFs(baseFs, opts.Filter)
	} else {
		return baseFs
	}
}

func RestrictedReadDir(opt ...Option) interp.ReadDirHandlerFunc2 {
	opts := options{}
	for _, o := range opt {
		o(&opts)
	}
	f := fsFromOpts(opts)
	return func(ctx context.Context, path string) ([]fs.DirEntry, error) {
		infos, err := afero.ReadDir(f, path)
		if err != nil {
			return nil, err
		}
		entries := make([]fs.DirEntry, len(infos))
		for i, info := range infos {
			entries[i] = fs.FileInfoToDirEntry(info)
		}
		return entries, nil
	}
}

func RestrictedStat(opt ...Option) interp.StatHandlerFunc {
	opts := options{}
	for _, o := range opt {
		o(&opts)
	}
	f := fsFromOpts(opts)
	return func(ctx context.Context, path string, followSymlinks bool) (fs.FileInfo, error) {
		if !followSymlinks {
			if lst, ok := f.(afero.Lstater); ok {
				info, _, err := lst.LstatIfPossible(path)
				return info, err
			}
		}
		return f.Stat(path)
	}
}

func RestrictedOpen(opt ...Option) interp.OpenHandlerFunc {
	opts := options{}
	for _, o := range opt {
		o(&opts)
	}
	f := fsFromOpts(opts)
	return func(ctx context.Context, path string, flag int, perm fs.FileMode) (io.ReadWriteCloser, error) {
		mc := interp.HandlerCtx(ctx)
		if path != "" && !filepath.IsAbs(path) {
			path = filepath.Join(mc.Dir, path)
		}
		slog.Warn("open", "path", path, "flag", flag, "perm", perm)
		return f.OpenFile(path, flag, perm)
	}
}
