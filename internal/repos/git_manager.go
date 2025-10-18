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

package repos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
)

type gitRepository interface {
	FetchContext(context.Context, *git.FetchOptions) error
	DeleteTag(string) error
	CommitObject(h plumbing.Hash) (*object.Commit, error)
	Worktree() (*git.Worktree, error)
}

type GitManager struct{}

type GitChanges struct {
	Patch *object.Patch
	Old   *object.Commit
	New   *object.Commit
}

func (gm *GitManager) GetChanges(r gitRepository, old, new *plumbing.Reference) (GitChanges, error) {
	var changes GitChanges
	var err error

	changes.Old, err = r.CommitObject(old.Hash())
	if err != nil {
		return changes, err
	}

	changes.New, err = r.CommitObject(new.Hash())
	if err != nil {
		return changes, err
	}

	changes.Patch, err = changes.Old.Patch(changes.New)
	if err != nil {
		return changes, fmt.Errorf("error to create patch: %w", err)
	}

	return changes, nil
}

func (gm *GitManager) initRepo(repoDir string) (*git.Repository, error) {
	if err := os.RemoveAll(repoDir); err != nil {
		return nil, fmt.Errorf("failed to remove repo directory: %w", err)
	}

	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create repo directory: %w", err)
	}

	r, err := git.PlainInit(repoDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git repo: %w", err)
	}

	return r, nil
}

func (gm *GitManager) tryOpenRepo(repoDir, repoUrl string) (*git.Repository, bool, error) {
	gitDir := filepath.Join(repoDir, ".git")
	fi, err := os.Stat(gitDir)
	if err != nil || !fi.IsDir() {
		return nil, false, fmt.Errorf("git dir not found")
	}

	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open repo: %w", err)
	}

	if err := updateRemoteURL(r, repoUrl); err != nil {
		return nil, false, fmt.Errorf("failed to update remote: %w", err)
	}

	_, err = r.Head()
	switch {
	case err == nil:
		return r, false, nil
	case errors.Is(err, plumbing.ErrReferenceNotFound):
		return r, true, nil
	default:
		return nil, false, fmt.Errorf("failed to get HEAD: %w", err)
	}
}

func (gm *GitManager) ReadGitRepo(repoDir, repoUrl string) (*git.Repository, bool, error) {
	if repo, fresh, err := gm.tryOpenRepo(repoDir, repoUrl); err == nil {
		return repo, fresh, nil
	}

	return gm.initAndConfigureRepo(repoDir, repoUrl)
}

func (gm *GitManager) initAndConfigureRepo(repoDir, repoUrl string) (*git.Repository, bool, error) {
	r, err := gm.initRepo(repoDir)
	if err != nil {
		return nil, false, err
	}

	_, err = r.CreateRemote(&gitConfig.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{repoUrl},
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to create remote: %w", err)
	}

	return r, true, nil
}

func (gm *GitManager) defaultFetchOptions() *git.FetchOptions {
	return &git.FetchOptions{
		Progress: os.Stderr,
		Force:    true,
		RefSpecs: []gitConfig.RefSpec{
			gitConfig.RefSpec(gitConfig.DefaultPushRefSpec),
		},
	}
}

func (gm *GitManager) fetch(ctx context.Context, r gitRepository, opts *git.FetchOptions) error {
	err := r.FetchContext(ctx, opts)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

func (gm *GitManager) fetchRepoByRef(ctx context.Context, r gitRepository, ref string) error {
	tempTag := fmt.Sprintf("stplr-temp-tag-%s", ref)
	gitref := gitConfig.RefSpec(fmt.Sprintf("%s:refs/tags/%s", ref, tempTag))
	err := gitref.Validate()
	if err != nil {
		return err
	}

	opts := gm.defaultFetchOptions()
	opts.RefSpecs = append(opts.RefSpecs, gitref)
	defer func() {
		if err := r.DeleteTag(tempTag); err != nil {
			slog.Debug("failed to delete temp tag", "tag", tempTag, "err", err)
		}
	}()

	return gm.fetch(ctx, r, opts)
}

func (gm *GitManager) FetchRepo(ctx context.Context, r gitRepository, ref string) error {
	if ref != "" {
		err := gm.fetchRepoByRef(ctx, r, ref)
		if err == nil {
			return nil
		}
		slog.Debug("fetch by ref failed, fallback to default", "ref", ref, "err", err)
	}
	return gm.fetch(ctx, r, gm.defaultFetchOptions())
}

func (gm *GitManager) FetchRepoWithProgress(ctx context.Context, r gitRepository, ref string, progress sideband.Progress) error {
	if ref != "" {
		err := gm.fetchRepoByRef(ctx, r, ref)
		if err == nil {
			return nil
		}
		slog.Debug("fetch by ref failed, fallback to default", "ref", ref, "err", err)
	}
	options := gm.defaultFetchOptions()
	options.Progress = progress
	return gm.fetch(ctx, r, options)
}

func (gm *GitManager) CheckoutRevision(r gitRepository, revHash *plumbing.Hash) (billy.Filesystem, error) {
	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(revHash.String()),
		Force: true,
	})
	if err != nil {
		return nil, err
	}

	return w.Filesystem, nil
}
