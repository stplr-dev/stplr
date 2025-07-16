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

package repos

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/leonelquinteros/gotext"
	"github.com/pelletier/go-toml/v2"
	"go.elara.ws/vercmp"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/shutils/handlers"
	"go.stplr.dev/stplr/pkg/types"
)

type actionType uint8

const (
	actionDelete actionType = iota
	actionUpdate
)

type action struct {
	Type actionType
	File string
}

// Pull pulls the provided repositories. If a repo doesn't exist, it will be cloned
// and its packages will be written to the DB. If it does exist, it will be pulled.
// In this case, only changed packages will be processed if possible.
// If repos is set to nil, the repos in the ALR config will be used.
func (rs *Repos) Pull(ctx context.Context, repos []types.Repo) error {
	if repos == nil {
		repos = rs.cfg.Repos()
	}

	for _, repo := range repos {
		err := rs.pullRepo(ctx, &repo, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rs *Repos) PullOneAndUpdateFromConfig(ctx context.Context, repo *types.Repo) error {
	err := rs.pullRepo(ctx, repo, true)
	if err != nil {
		return err
	}

	return nil
}

func (rs *Repos) pullRepo(ctx context.Context, repo *types.Repo, updateRepoFromToml bool) error {
	urls := []string{repo.URL}
	urls = append(urls, repo.Mirrors...)

	var lastErr error

	for i, repoURL := range urls {
		if i > 0 {
			slog.Info(gotext.Get("Trying mirror"), "repo", repo.Name, "mirror", repoURL)
		}

		err := rs.pullRepoFromURL(ctx, repoURL, repo, updateRepoFromToml)
		if err != nil {
			lastErr = err
			slog.Warn(gotext.Get("Failed to pull from URL"), "repo", repo.Name, "url", repoURL, "error", err)
			continue
		}

		// Success
		return nil
	}

	return fmt.Errorf("failed to pull repository %s from any URL: %w", repo.Name, lastErr)
}

func readGitRepo(repoDir, repoUrl string) (*git.Repository, bool, error) {
	gitDir := filepath.Join(repoDir, ".git")
	if fi, err := os.Stat(gitDir); err == nil && fi.IsDir() {
		r, err := git.PlainOpen(repoDir)
		if err == nil {
			err = updateRemoteURL(r, repoUrl)
			if err == nil {
				_, err := r.Head()
				if err == nil {
					return r, false, nil
				}

				if errors.Is(err, plumbing.ErrReferenceNotFound) {
					return r, true, nil
				}

				slog.Debug("error getting HEAD, reinitializing...", "err", err)
			}
		}

		slog.Debug("error while reading repo, reinitializing...", "err", err)
	}

	if err := os.RemoveAll(repoDir); err != nil {
		return nil, false, fmt.Errorf("failed to remove repo directory: %w", err)
	}

	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		return nil, false, fmt.Errorf("failed to create repo directory: %w", err)
	}

	r, err := git.PlainInit(repoDir, false)
	if err != nil {
		return nil, false, fmt.Errorf("failed to initialize git repo: %w", err)
	}

	_, err = r.CreateRemote(&gitConfig.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{repoUrl},
	})
	if err != nil {
		return nil, false, err
	}

	return r, true, nil
}

func (rs *Repos) pullRepoFromURL(ctx context.Context, rawRepoUrl string, repo *types.Repo, update bool) error {
	repoURL, err := url.Parse(rawRepoUrl)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", rawRepoUrl, err)
	}

	slog.Info(gotext.Get("Pulling repository"), "name", repo.Name)
	repoDir := filepath.Join(rs.cfg.GetPaths().RepoDir, repo.Name)

	var repoFS billy.Filesystem

	r, freshGit, err := readGitRepo(repoDir, repoURL.String())
	if err != nil {
		return fmt.Errorf("failed to open repo")
	}

	err = r.FetchContext(ctx, &git.FetchOptions{
		Progress: os.Stderr,
		Force:    true,
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	var old *plumbing.Reference

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	revHash, err := resolveHash(r, repo.Ref)
	if err != nil {
		return fmt.Errorf("error resolving hash: %w", err)
	}

	if !freshGit {
		old, err = r.Head()
		if err != nil {
			return err
		}

		if old.Hash() == *revHash {
			slog.Info(gotext.Get("Repository up to date"), "name", repo.Name)
		}
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(revHash.String()),
		Force: true,
	})
	if err != nil {
		return err
	}
	repoFS = w.Filesystem

	new, err := r.Head()
	if err != nil {
		return err
	}

	// If the DB was not present at startup, that means it's
	// empty. In this case, we need to update the DB fully
	// rather than just incrementally.
	if rs.db.IsEmpty() || freshGit {
		err = rs.processRepoFull(ctx, *repo, repoDir)
		if err != nil {
			return err
		}
	} else {
		err = rs.processRepoChanges(ctx, *repo, r, w, old, new)
		if err != nil {
			return err
		}
	}

	fl, err := repoFS.Open(constants.RepoConfigFile)
	if err != nil {
		slog.Warn(gotext.Get("Git repository does not appear to be a valid Stapler repo"), "repo", repo.Name)
		return nil
	}

	var repoCfg types.RepoConfig
	err = toml.NewDecoder(fl).Decode(&repoCfg)
	if err != nil {
		return err
	}
	fl.Close()

	// If the version doesn't have a "v" prefix, it's not a standard version.
	// It may be "unknown" or a git version, but either way, there's no way
	// to compare it to the repo version, so only compare versions with the "v".
	if strings.HasPrefix(config.Version, "v") {
		if vercmp.Compare(config.Version, repoCfg.Repo.MinVersion) == -1 {
			slog.Warn(gotext.Get("Stapler repo's minimum Stapler version is greater than the current version. Try updating Stapler if something doesn't work."), "repo", repo.Name)
		}
	}

	if update {
		if repoCfg.Repo.URL != "" {
			repo.URL = repoCfg.Repo.URL
		}
		if repoCfg.Repo.Ref != "" {
			repo.Ref = repoCfg.Repo.Ref
		}
		if len(repoCfg.Repo.Mirrors) > 0 {
			repo.Mirrors = repoCfg.Repo.Mirrors
		}
	}

	return nil
}

func updateRemoteURL(r *git.Repository, newURL string) error {
	cfg, err := r.Config()
	if err != nil {
		return err
	}

	remote, ok := cfg.Remotes[git.DefaultRemoteName]
	if !ok || len(remote.URLs) == 0 {
		return fmt.Errorf("no remote '%s' found", git.DefaultRemoteName)
	}

	currentURL := remote.URLs[0]
	if currentURL == newURL {
		return nil
	}

	slog.Debug("Updating remote URL", "old", currentURL, "new", newURL)

	err = r.DeleteRemote(git.DefaultRemoteName)
	if err != nil {
		return fmt.Errorf("failed to delete old remote: %w", err)
	}

	_, err = r.CreateRemote(&gitConfig.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{newURL},
	})
	if err != nil {
		return fmt.Errorf("failed to create new remote: %w", err)
	}

	return nil
}

func (rs *Repos) updatePkg(ctx context.Context, repo types.Repo, runner *interp.Runner, scriptFl io.ReadCloser) error {
	parser := syntax.NewParser()

	pkgs, err := parseScript(ctx, repo, parser, runner, scriptFl)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		err = rs.db.InsertPackage(ctx, *pkg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rs *Repos) processRepoChangesRunner(repoDir, scriptDir string) (*interp.Runner, error) {
	env := append(os.Environ(), "scriptdir="+scriptDir)
	return interp.New(
		interp.Env(expand.ListEnviron(env...)),
		interp.ExecHandler(handlers.NopExec),
		interp.ReadDirHandler2(handlers.RestrictedReadDir(repoDir)),
		interp.StatHandler(handlers.RestrictedStat(repoDir)),
		interp.OpenHandler(handlers.RestrictedOpen(repoDir)),
		interp.StdIO(handlers.NopRWC{}, handlers.NopRWC{}, handlers.NopRWC{}),
		// Use temp dir instead script dir because runner may be for deleted file
		interp.Dir(os.TempDir()),
	)
}

func (rs *Repos) processRepoChanges(ctx context.Context, repo types.Repo, r *git.Repository, w *git.Worktree, old, new *plumbing.Reference) error {
	oldCommit, err := r.CommitObject(old.Hash())
	if err != nil {
		return err
	}

	newCommit, err := r.CommitObject(new.Hash())
	if err != nil {
		return err
	}

	patch, err := oldCommit.Patch(newCommit)
	if err != nil {
		return fmt.Errorf("error to create patch: %w", err)
	}

	var actions []action
	for _, fp := range patch.FilePatches() {
		from, to := fp.Files()

		var isValidPath bool
		if from != nil {
			isValidPath = isValidScriptPath(from.Path())
		}
		if to != nil {
			isValidPath = isValidPath || isValidScriptPath(to.Path())
		}

		if !isValidPath {
			continue
		}

		switch {
		case to == nil:
			actions = append(actions, action{
				Type: actionDelete,
				File: from.Path(),
			})
		case from == nil:
			actions = append(actions, action{
				Type: actionUpdate,
				File: to.Path(),
			})
		case from.Path() != to.Path():
			actions = append(actions,
				action{
					Type: actionDelete,
					File: from.Path(),
				},
				action{
					Type: actionUpdate,
					File: to.Path(),
				},
			)
		default:
			slog.Debug("unexpected, but I'll try to do")
			actions = append(actions, action{
				Type: actionUpdate,
				File: to.Path(),
			})
		}
	}

	repoDir := w.Filesystem.Root()
	parser := syntax.NewParser()

	for _, action := range actions {
		var scriptDir string
		if filepath.Dir(action.File) == "." {
			scriptDir = repoDir
		} else {
			scriptDir = filepath.Dir(filepath.Join(repoDir, action.File))
		}

		runner, err := rs.processRepoChangesRunner(repoDir, scriptDir)
		if err != nil {
			return fmt.Errorf("error creating process repo changes runner: %w", err)
		}

		switch action.Type {
		case actionDelete:
			scriptFl, err := oldCommit.File(action.File)
			if err != nil {
				slog.Warn("Failed to get deleted file from old commit", "file", action.File, "error", err)
				continue
			}

			r, err := scriptFl.Reader()
			if err != nil {
				slog.Warn("Failed to read deleted file", "file", action.File, "error", err)
				continue
			}

			pkgs, err := parseScript(ctx, repo, parser, runner, r)
			if err != nil {
				return fmt.Errorf("error parsing deleted script %s: %w", action.File, err)
			}

			for _, pkg := range pkgs {
				err = rs.db.DeletePkgs(ctx, "name = ? AND repository = ?", pkg.Name, repo.Name)
				if err != nil {
					return fmt.Errorf("error deleting package %s: %w", pkg.Name, err)
				}
			}

		case actionUpdate:
			scriptFl, err := newCommit.File(action.File)
			if err != nil {
				slog.Warn("Failed to get updated file from new commit", "file", action.File, "error", err)
				continue
			}

			r, err := scriptFl.Reader()
			if err != nil {
				slog.Warn("Failed to read updated file", "file", action.File, "error", err)
				continue
			}

			err = rs.updatePkg(ctx, repo, runner, r)
			if err != nil {
				return fmt.Errorf("error updating package from %s: %w", action.File, err)
			}
		}
	}

	return nil
}

func isValidScriptPath(path string) bool {
	if filepath.Base(path) != "Staplerfile" {
		return false
	}

	dir := filepath.Dir(path)
	return dir == "." || !strings.Contains(strings.TrimPrefix(dir, "./"), "/")
}

func (rs *Repos) processRepoFull(ctx context.Context, repo types.Repo, repoDir string) error {
	rootScript := filepath.Join(repoDir, "Staplerfile")
	if fi, err := os.Stat(rootScript); err == nil && !fi.IsDir() {
		slog.Debug("Found root Staplerfile, processing single-script repository", "repo", repo.Name)

		runner, err := rs.processRepoChangesRunner(repoDir, repoDir)
		if err != nil {
			return fmt.Errorf("error creating runner for root Staplerfile: %w", err)
		}

		scriptFl, err := os.Open(rootScript)
		if err != nil {
			return fmt.Errorf("error opening root Staplerfile: %w", err)
		}
		defer scriptFl.Close()

		err = rs.updatePkg(ctx, repo, runner, scriptFl)
		if err != nil {
			return fmt.Errorf("error processing root Staplerfile: %w", err)
		}

		return nil
	}

	glob := filepath.Join(repoDir, "*/Staplerfile")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("error globbing for Staplerfile files: %w", err)
	}

	if len(matches) == 0 {
		slog.Warn("No Staplerfile files found in repository", "repo", repo.Name)
		return nil
	}

	slog.Debug("Found multiple Staplerfile files, processing multi-package repository",
		"repo", repo.Name, "count", len(matches))

	for _, match := range matches {
		runner, err := rs.processRepoChangesRunner(repoDir, filepath.Dir(match))
		if err != nil {
			return fmt.Errorf("error creating runner for %s: %w", match, err)
		}

		scriptFl, err := os.Open(match)
		if err != nil {
			return fmt.Errorf("error opening %s: %w", match, err)
		}

		err = rs.updatePkg(ctx, repo, runner, scriptFl)
		scriptFl.Close()
		if err != nil {
			return fmt.Errorf("error processing %s: %w", match, err)
		}
	}

	return nil
}
