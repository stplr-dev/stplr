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
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

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

	changes *GitChanges
}

type RepoProcessor struct {
	gm *GitManager
}

func (rp *RepoProcessor) ProcessFull(ctx context.Context, repo types.Repo, repoDir string) ([]PackageAction, error) {
	actions, err := rp.actionsFromDir(repoDir)
	if err != nil {
		return nil, err
	}
	if len(actions) == 0 {
		slog.Warn("No Staplerfile files found in repository", "repo", repo.Name)
		return nil, nil
	}
	return rp.processActions(ctx, actions, repo)
}

func (rp *RepoProcessor) ProcessChanges(ctx context.Context, repo types.Repo, r *git.Repository, old, new *plumbing.Reference) ([]PackageAction, error) {
	changes, err := rp.gm.GetChanges(r, old, new)
	if err != nil {
		return nil, err
	}
	actions := rp.actionsFromChanges(changes)
	return rp.processActions(ctx, actions, repo)
}

func (rp *RepoProcessor) actionsFromDir(repoDir string) ([]action, error) {
	rootScript := filepath.Join(repoDir, "Staplerfile")
	if fi, err := os.Stat(rootScript); err == nil && !fi.IsDir() {
		return []action{{
			Type: actionUpdate,
			File: rootScript,
		}}, nil
	}

	glob := filepath.Join(repoDir, "*/Staplerfile")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("error globbing for Staplerfile files: %w", err)
	}

	if len(matches) == 0 {
		return nil, nil
	}

	acts := make([]action, len(matches))

	for i, match := range matches {
		acts[i] = action{
			Type: actionUpdate,
			File: match,
		}
	}

	return acts, nil
}

func (rp *RepoProcessor) actionsFromChanges(changes GitChanges) []action {
	var actions []action
	for _, fp := range changes.Patch.FilePatches() {
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
				Type:    actionDelete,
				File:    from.Path(),
				changes: &changes,
			})
		case from == nil:
			actions = append(actions, action{
				Type:    actionUpdate,
				File:    to.Path(),
				changes: &changes,
			})
		case from.Path() != to.Path():
			actions = append(actions,
				action{
					Type:    actionDelete,
					File:    from.Path(),
					changes: &changes,
				},
				action{
					Type:    actionUpdate,
					File:    to.Path(),
					changes: &changes,
				},
			)
		default:
			slog.Debug("unexpected, but I'll try to do")
			actions = append(actions, action{
				Type:    actionUpdate,
				File:    to.Path(),
				changes: &changes,
			})
		}
	}

	return actions
}

func (rp *RepoProcessor) processActions(ctx context.Context, actions []action, repo types.Repo) ([]PackageAction, error) {
	res := []PackageAction{}
	for _, action := range actions {
		m, err := rp.actionToPackageAction(ctx, action, repo)
		if err != nil {
			return nil, err
		}
		if len(m) > 0 {
			res = append(res, m...)
		}
	}

	return res, nil
}

func readerFromAction(act action) (io.ReadCloser, error) {
	if act.changes == nil {
		return os.Open(act.File)
	}

	var c *object.Commit
	switch act.Type {
	case actionDelete:
		c = act.changes.Old
	case actionUpdate:
		c = act.changes.New
	}

	f, err := c.File(act.File)
	if err != nil {
		return nil, err
	}

	r, err := f.Reader()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func toPackageActionType(t actionType) PackageActionType {
	switch t {
	case actionDelete:
		return ActionDelete
	case actionUpdate:
		return ActionUpsert
	default:
		panic("unexpected actionType")
	}
}

func (rp *RepoProcessor) actionToPackageAction(ctx context.Context, act action, repo types.Repo) ([]PackageAction, error) {
	res := []PackageAction{}

	t := toPackageActionType(act.Type)
	r, err := readerFromAction(act)
	if err != nil {
		return nil, err
	}

	pkgs, err := parseScript(ctx, repo, r)
	if err != nil {
		return nil, fmt.Errorf("error parsing deleted script %s: %w", act.File, err)
	}

	for _, pkg := range pkgs {
		res = append(res, PackageAction{
			Type: t,
			Pkg:  pkg,
		})
	}

	return res, nil
}
