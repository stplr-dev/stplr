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

package puller

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/leonelquinteros/gotext"
	"github.com/pelletier/go-toml/v2"
	"go.elara.ws/vercmp"

	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/constants"
	database "go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/plugins/shared"
	"go.stplr.dev/stplr/internal/repoprocessor"
	"go.stplr.dev/stplr/internal/service/repos/internal/gitmanager"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type packageActionType uint8

const (
	ActionDelete packageActionType = iota
	ActionUpsert
)

type PackageAction struct {
	Type packageActionType
	Pkg  *staplerfile.Package
}

// Puller executes in plugin
type Puller struct {
	cfg *config.ALRConfig
	db  *database.Database

	rp *repoprocessor.RepoProcessor
	gm *gitmanager.GitManager
}

func NewPuller(cfg *config.ALRConfig, db *database.Database) *Puller {
	return &Puller{
		cfg,
		db,
		repoprocessor.New(),
		&gitmanager.GitManager{},
	}
}

func (p *Puller) Pull(ctx context.Context, repo types.Repo, report PullReporter) (types.Repo, error) {
	urls := []string{repo.URL}
	urls = append(urls, repo.Mirrors...)

	var lastErr error

	for i, repoURL := range urls {
		err := report.Notify(ctx, EventTryPull, map[string]string{
			"i":   strconv.Itoa(i),
			"url": repoURL,
		})
		if err != nil {
			slog.Warn("failed to notify", "err", err)
		}

		if err := p.pull(ctx, repoURL, &repo, report); err != nil {
			lastErr = err
			err = report.Notify(ctx, EventErrorPull, map[string]string{
				"i":   strconv.Itoa(i),
				"url": repoURL,
				"err": string(err.Error()),
			})
			if err != nil {
				slog.Warn("failed to notify", "err", err)
			}
			continue
		}
		return repo, nil
	}

	return repo, fmt.Errorf("failed to pull repository %s from any URL: %w", repo.Name, lastErr)
}

func (p *Puller) Read(ctx context.Context, repo types.Repo, report PullReporter) (types.Repo, error) {
	repoDir := filepath.Join(p.cfg.GetPaths().RepoDir, repo.Name)

	if err := p.processRepoChanges(ctx, repo, repoDir); err != nil {
		return repo, err
	}

	if err := p.loadAndUpdateConfig(repoDir, &repo); err != nil {
		return repo, err
	}

	return repo, nil
}

func (p *Puller) pull(ctx context.Context, rawRepoUrl string, repo *types.Repo, report PullReporter) error {
	repoURL, err := url.Parse(rawRepoUrl)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", rawRepoUrl, err)
	}

	repoDir := filepath.Join(p.cfg.GetPaths().RepoDir, repo.Name)

	r, isGitFresh, err := p.gm.ReadGitRepo(repoDir, repoURL.String())
	if err != nil {
		return fmt.Errorf("failed to open repo")
	}

	err = p.gm.FetchRepoWithProgress(ctx, r, repo.Ref, shared.ToIoWriter(report, EventGitPullProgress))
	if err != nil {
		return err
	}

	_, revHash, err := p.resolveRevision(r, *repo, isGitFresh)
	if err != nil {
		return err
	}

	if _, err = p.gm.CheckoutRevision(r, revHash); err != nil {
		return err
	}

	if err := p.processRepoChanges(ctx, *repo, repoDir); err != nil {
		return err
	}

	return p.loadAndUpdateConfig(repoDir, repo)
}

func (p *Puller) resolveRevision(r *git.Repository, repo types.Repo, isFresh bool) (*plumbing.Reference, *plumbing.Hash, error) {
	revHash, err := p.gm.ResolveHash(r, repo.Ref)
	if err != nil {
		return nil, nil, fmt.Errorf("error resolving hash: %w", err)
	}

	if isFresh {
		return nil, revHash, nil
	}

	head, err := r.Head()
	if err != nil {
		return nil, nil, err
	}

	return head, revHash, nil
}

func (p *Puller) processRepoChanges(ctx context.Context, repo types.Repo, repoDir string) error {
	if err := p.db.DeletePkgs(ctx, "repository = ?", repo.Name); err != nil {
		return fmt.Errorf("failed to remove pkgs: %w", err)
	}

	pkgs, err := p.rp.Process(ctx, repo, repoDir)
	if err != nil {
		return fmt.Errorf("failed to process %q repo: %w", repo.Name, err)
	}

	for _, pkg := range pkgs {
		if err := p.db.InsertPackage(ctx, *pkg); err != nil {
			return err
		}
	}

	return nil
}

func (rs *Puller) loadAndUpdateConfig(repoDir string, repo *types.Repo) error {
	fl, err := os.Open(filepath.Join(repoDir, constants.RepoConfigFile))
	if err != nil {
		// TODO:
		slog.Warn(gotext.Get("Git repository does not appear to be a valid Stapler repo"), "repo", repo.Name)
		return nil
	}
	defer fl.Close()

	var repoCfg types.RepoConfig
	if err := toml.NewDecoder(fl).Decode(&repoCfg); err != nil {
		return err
	}

	warnAboutVersion(*repo, repoCfg)

	if repoCfg.Repo.URL != "" {
		repo.URL = repoCfg.Repo.URL
	}
	if repoCfg.Repo.Ref != "" {
		repo.Ref = repoCfg.Repo.Ref
	}
	if len(repoCfg.Repo.Mirrors) > 0 {
		repo.Mirrors = repoCfg.Repo.Mirrors
	}
	repo.ReportUrl = repoCfg.Repo.ReportUrl
	return nil
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

var _ PullExecutor = &Puller{}
