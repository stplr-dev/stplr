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
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/leonelquinteros/gotext"
	"github.com/pelletier/go-toml/v2"
	"go.elara.ws/vercmp"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type PackageActionType uint8

const (
	ActionDelete PackageActionType = iota
	ActionUpsert
)

type PackageAction struct {
	Type PackageActionType
	Pkg  *staplerfile.Package
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

func (rs *Repos) pullRepoFromURL(ctx context.Context, rawRepoUrl string, repo *types.Repo, update bool) error {
	repoURL, err := url.Parse(rawRepoUrl)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", rawRepoUrl, err)
	}

	slog.Info(gotext.Get("Pulling repository"), "name", repo.Name)
	repoDir := filepath.Join(rs.cfg.GetPaths().RepoDir, repo.Name)

	r, freshGit, err := rs.gm.ReadGitRepo(repoDir, repoURL.String())
	if err != nil {
		return fmt.Errorf("failed to open repo")
	}

	err = rs.gm.FetchRepo(ctx, r)
	if err != nil {
		return err
	}

	old, revHash, err := resolveRevision(r, repo, freshGit)
	if err != nil {
		return err
	}

	repoFS, err := rs.gm.CheckoutRevision(r, revHash)
	if err != nil {
		return err
	}

	newRef, err := r.Head()
	if err != nil {
		return err
	}

	if err := rs.processRepoChanges(ctx, repo, repoDir, r, old, newRef, freshGit); err != nil {
		return err
	}

	return rs.loadAndUpdateConfig(repoFS, repo, update)
}

func (rs *Repos) loadAndUpdateConfig(repoFS billy.Filesystem, repo *types.Repo, update bool) error {
	fl, err := repoFS.Open(constants.RepoConfigFile)
	if err != nil {
		slog.Warn(gotext.Get("Git repository does not appear to be a valid Stapler repo"), "repo", repo.Name)
		return nil
	}
	defer fl.Close()

	var repoCfg types.RepoConfig
	if err := toml.NewDecoder(fl).Decode(&repoCfg); err != nil {
		return err
	}

	warnAboutVersion(*repo, repoCfg)

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

func (rs *Repos) processRepoChanges(ctx context.Context, repo *types.Repo, repoDir string, r *git.Repository, old, new *plumbing.Reference, freshGit bool) error {
	var actions []PackageAction
	var err error

	if rs.db.IsEmpty() || freshGit {
		actions, err = rs.rp.ProcessFull(ctx, *repo, repoDir)
	} else {
		actions, err = rs.rp.ProcessChanges(ctx, *repo, r, old, new)
	}
	if err != nil {
		return err
	}

	return rs.processActions(ctx, *repo, actions)
}

func warnAboutVersion(repo types.Repo, cfg types.RepoConfig) {
	// If the version doesn't have a "v" prefix, it's not a standard version.
	// It may be "unknown" or a git version, but either way, there's no way
	// to compare it to the repo version, so only compare versions with the "v".
	if strings.HasPrefix(config.Version, "v") {
		if vercmp.Compare(config.Version, cfg.Repo.MinVersion) == -1 {
			slog.Warn(gotext.Get("Stapler repo's minimum Stapler version is greater than the current version. Try updating Stapler if something doesn't work."), "repo", repo.Name)
		}
	}
}

func resolveRevision(r *git.Repository, repo *types.Repo, freshGit bool) (*plumbing.Reference, *plumbing.Hash, error) {
	revHash, err := resolveHash(r, repo.Ref)
	if err != nil {
		return nil, nil, fmt.Errorf("error resolving hash: %w", err)
	}

	if freshGit {
		return nil, revHash, nil
	}

	old, err := r.Head()
	if err != nil {
		return nil, nil, err
	}

	if old.Hash() == *revHash {
		slog.Info(gotext.Get("Repository up to date"), "name", repo.Name)
	}

	return old, revHash, nil
}

func (rs *Repos) processActions(ctx context.Context, repo types.Repo, actions []PackageAction) error {
	for _, action := range actions {
		switch action.Type {
		case ActionUpsert:
			err := rs.db.InsertPackage(ctx, *action.Pkg)
			if err != nil {
				return err
			}
		case ActionDelete:
			err := rs.db.DeletePkgs(ctx, "name = ? AND repository = ?", action.Pkg.Name, repo.Name)
			if err != nil {
				return fmt.Errorf("error deleting package %s: %w", action.Pkg.Name, err)
			}
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

func isValidScriptPath(path string) bool {
	if filepath.Base(path) != "Staplerfile" {
		return false
	}

	dir := filepath.Dir(path)
	return dir == "." || !strings.Contains(strings.TrimPrefix(dir, "./"), "/")
}
