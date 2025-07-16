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
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitDownloader downloads Git repositories
type GitDownloader struct{}

// Name always returns "git"
func (GitDownloader) Name() string {
	return "git"
}

// MatchURL matches any URLs that start with "git+"
func (GitDownloader) MatchURL(u string) bool {
	return strings.HasPrefix(u, "git+")
}

// Download uses git to clone the repository from the specified URL.
// It allows specifying the revision, depth and recursion options
// via query string
func (d *GitDownloader) Download(ctx context.Context, opts Options) (Type, string, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return 0, "", err
	}
	u.Scheme = strings.TrimPrefix(u.Scheme, "git+")

	query := u.Query()

	rev := query.Get("~rev")
	query.Del("~rev")

	// Right now, this only affects the return value of name,
	// which will be used by dl_cache.
	// It seems wrong, but for now it's better to leave it as it is.
	name := query.Get("~name")
	query.Del("~name")

	depthStr := query.Get("~depth")
	query.Del("~depth")

	recursive := query.Get("~recursive")
	query.Del("~recursive")

	u.RawQuery = query.Encode()

	depth := 0
	if depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil {
			return 0, "", err
		}
	}

	co := &git.CloneOptions{
		URL:               u.String(),
		Depth:             depth,
		Progress:          opts.Progress,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}

	if recursive == "true" {
		co.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	r, err := git.PlainCloneContext(ctx, opts.Destination, false, co)
	if err != nil {
		return 0, "", err
	}

	err = r.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{"+refs/*:refs/*"},
	})
	if err != git.NoErrAlreadyUpToDate && err != nil {
		return 0, "", err
	}

	if rev != "" {
		h, err := r.ResolveRevision(plumbing.Revision(rev))
		if err != nil {
			return 0, "", err
		}

		w, err := r.Worktree()
		if err != nil {
			return 0, "", err
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: *h,
		})
		if err != nil {
			return 0, "", err
		}
	}

	err = VerifyHashFromLocal("", opts)
	if err != nil {
		return 0, "", err
	}

	if name == "" {
		name = strings.TrimSuffix(path.Base(u.Path), ".git")
	}

	return TypeDir, name, nil
}

// Update uses git to pull the repository and update it
// to the latest revision. It allows specifying the depth
// and recursion options via query string. It returns
// true if update was successful and false if the
// repository is already up-to-date
func (d *GitDownloader) Update(opts Options) (bool, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return false, err
	}
	u.Scheme = strings.TrimPrefix(u.Scheme, "git+")

	query := u.Query()
	rev := query.Get("~rev")
	query.Del("~rev")

	depthStr := query.Get("~depth")
	query.Del("~depth")

	recursive := query.Get("~recursive")
	query.Del("~recursive")

	u.RawQuery = query.Encode()

	r, err := git.PlainOpen(opts.Destination)
	if err != nil {
		return false, err
	}

	w, err := r.Worktree()
	if err != nil {
		return false, err
	}

	depth := 0
	if depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil {
			return false, err
		}
	}

	// First, we do a fetch to get all the revisions.
	fo := &git.FetchOptions{
		Depth:    depth,
		Progress: opts.Progress,
	}

	m, err := getManifest(opts.Destination)
	manifestOK := err == nil

	err = r.Fetch(fo)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return false, err
	}

	// If a revision is specified, switch to it.
	if rev != "" {
		// We are trying to find the revision as a hash of the commit
		hash, err := r.ResolveRevision(plumbing.Revision(rev))
		if err != nil {
			return false, fmt.Errorf("failed to resolve revision %s: %w", rev, err)
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: *hash,
		})
		if err != nil {
			return false, fmt.Errorf("failed to checkout revision %s: %w", rev, err)
		}

		if recursive == "true" {
			submodules, err := w.Submodules()
			if err == nil {
				err = submodules.Update(&git.SubmoduleUpdateOptions{
					Init: true,
				})
				if err != nil {
					return false, fmt.Errorf("failed to update submodules %s: %w", rev, err)
				}
			}
		}
	} else {
		// If the revision is not specified, we do a regular pull.
		po := &git.PullOptions{
			Depth:             depth,
			Progress:          opts.Progress,
			RecurseSubmodules: git.NoRecurseSubmodules,
		}

		if recursive == "true" {
			po.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
		}

		err = w.Pull(po)
		if err != nil {
			if errors.Is(err, git.NoErrAlreadyUpToDate) {
				return false, nil
			}
			return false, err
		}
	}

	err = VerifyHashFromLocal("", opts)
	if err != nil {
		return false, err
	}

	if manifestOK {
		err = writeManifest(opts.Destination, m)
	}

	return true, err
}
