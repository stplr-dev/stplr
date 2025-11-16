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

package repoprocessor

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type RepoProcessor struct{}

func New() *RepoProcessor {
	return &RepoProcessor{}
}

func (rp *RepoProcessor) Process(ctx context.Context, repo types.Repo, repoDir string) ([]*staplerfile.Package, error) {
	rootScript := filepath.Join(repoDir, "Staplerfile")
	if fi, err := os.Stat(rootScript); err == nil && !fi.IsDir() {
		return rp.processFiles(ctx, repo, []string{rootScript})
	}

	glob := filepath.Join(repoDir, "*/Staplerfile")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("error globbing for Staplerfile files: %w", err)
	}

	return rp.processFiles(ctx, repo, matches)
}

func (rp *RepoProcessor) processFiles(ctx context.Context, repo types.Repo, files []string) ([]*staplerfile.Package, error) {
	var all []*staplerfile.Package
	for _, match := range files {
		f, err := os.Open(match)
		if err != nil {
			return nil, fmt.Errorf("failed to open %q: %w", match, err)
		}
		pkgs, err := rp.parseScript(ctx, repo, f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse script %q: %w", match, err)
		}
		all = slices.Concat(all, pkgs)
	}

	return all, nil
}

func (rp *RepoProcessor) parseScript(
	ctx context.Context,
	repo types.Repo,
	r io.ReadCloser,
) ([]*staplerfile.Package, error) {
	f, err := staplerfile.ReadFromIOReader(r, "/tmp")
	if err != nil {
		return nil, err
	}
	_, pkgs, err := f.ParseBuildVars(ctx, &distro.OSRelease{}, []string{})
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		pkg.Repository = repo.Name
	}
	return pkgs, nil
}
