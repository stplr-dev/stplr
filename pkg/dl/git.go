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
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
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

// parseGitURL parses the URL and extracts query parameters
func parseGitURL(rawURL string) (*url.URL, string, int, bool, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, "", 0, false, err
	}
	u.Scheme = strings.TrimPrefix(u.Scheme, "git+")

	query := u.Query()
	rev := query.Get("~rev")
	query.Del("~rev")
	depthStr := query.Get("~depth")
	query.Del("~depth")
	recursive := query.Get("~recursive") == "true"
	query.Del("~recursive")
	u.RawQuery = query.Encode()

	depth := 0
	if depthStr != "" {
		depth, err = strconv.Atoi(depthStr)
		if err != nil {
			return nil, "", 0, false, err
		}
	}

	return u, rev, depth, recursive, nil
}

// performFetch executes a git fetch operation
func performFetch(repo *git.Repository, depth int, progress io.Writer) error {
	fo := &git.FetchOptions{
		Depth:    depth,
		Progress: progress,
		// RefSpecs: []config.RefSpec{"+refs/*:refs/*"},
	}
	err := repo.Fetch(fo)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

func checkoutRevision(repo *git.Repository, rev string, recursive bool) error {
	if rev == "" {
		return nil
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return fmt.Errorf("failed to resolve revision %s: %w", rev, err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: *hash,
	})
	if err != nil {
		return fmt.Errorf("failed to checkout revision %s: %w", rev, err)
	}

	if recursive {
		submodules, err := w.Submodules()
		if err != nil {
			return nil // Ignore submodule errors
		}
		err = submodules.Update(&git.SubmoduleUpdateOptions{
			Init: true,
		})
		if err != nil {
			return fmt.Errorf("failed to update submodules %s: %w", rev, err)
		}
	}

	return nil
}

// performPull executes a git pull operation
func performPull(w *git.Worktree, depth int, recursive bool, progress io.Writer) (bool, error) {
	po := &git.PullOptions{
		Depth:             depth,
		Progress:          progress,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}
	if recursive {
		po.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	err := w.Pull(po)
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Download uses git to clone the repository from the specified URL.
// It allows specifying the revision, depth and recursion options
// via query string
func (d *GitDownloader) Download(ctx context.Context, opts Options) (Type, string, error) {
	u, rev, depth, recursive, err := parseGitURL(opts.URL)
	if err != nil {
		return 0, "", err
	}

	query := u.Query()
	name := query.Get("~name")
	query.Del("~name")
	u.RawQuery = query.Encode()

	if name == "" {
		name = strings.TrimSuffix(path.Base(u.Path), ".git")
	}

	co := &git.CloneOptions{
		URL:               u.String(),
		Depth:             depth,
		Progress:          opts.Progress,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}
	if recursive {
		co.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	r, err := git.PlainCloneContext(ctx, filepath.Join(opts.Destination, name), false, co)
	if err != nil {
		return 0, "", err
	}

	err = performFetch(r, depth, opts.Progress)
	if err != nil {
		return 0, "", err
	}

	if rev != "" {
		err = checkoutRevision(r, rev, recursive)
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
	u, rev, depth, recursive, err := parseGitURL(opts.URL)
	if err != nil {
		return false, err
	}

	query := u.Query()
	name := query.Get("~name")
	query.Del("~name")
	u.RawQuery = query.Encode()

	if name == "" {
		name = strings.TrimSuffix(path.Base(u.Path), ".git")
	}

	r, err := git.PlainOpen(filepath.Join(opts.Destination, name))
	if err != nil {
		return false, err
	}

	w, err := r.Worktree()
	if err != nil {
		return false, err
	}

	err = performFetch(r, depth, opts.Progress)
	if err != nil {
		return false, err
	}

	var updated bool
	if rev != "" {
		err = checkoutRevision(r, rev, recursive)
		if err != nil {
			return false, err
		}
		updated = true // Assume checkout changes the state
	} else {
		updated, err = performPull(w, depth, recursive, opts.Progress)
		if err != nil {
			return false, err
		}
	}

	err = VerifyHashFromLocal(name, opts)
	if err != nil {
		return false, err
	}

	return updated, err
}
