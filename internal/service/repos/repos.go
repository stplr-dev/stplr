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
	"syscall"

	stdErrors "errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/config"
	"go.stplr.dev/stplr/internal/db"
	"go.stplr.dev/stplr/internal/plugins/shared"
	"go.stplr.dev/stplr/internal/service/repos/internal/puller"
	"go.stplr.dev/stplr/pkg/types"
)

type (
	PullExecutor       = puller.PullExecutor
	Puller             = puller.Puller
	PullExecutorPlugin = puller.PullExecutorPlugin
	PullOptions        = puller.PullOptions
)

var NewPuller = puller.NewPuller

type SysDropper interface {
	DropCapsToBuilderUser() error
}

type Repos struct {
	cfg *config.ALRConfig
	db  *db.Database
	sys SysDropper
	rp  PullExecutor
}

func New(cfg *config.ALRConfig, db *db.Database, sys SysDropper, rp PullExecutor) *Repos {
	return &Repos{
		cfg: cfg,
		db:  db,
		sys: sys,
		rp:  rp,
	}
}

// Modify updates the configured repositories using the provided callback.
// The updated list is saved into the system configuration.
//
// WARNING: Any root-level operations must be performed before this method,
// because process privileges are permanently dropped to the builder user.
//
// The callback can return nil to exclude a repo from the final list.
func (r *Repos) Modify(ctx context.Context, c func(repo types.Repo) *types.Repo) error {
	return r.ModifySlice(ctx, func(repos []types.Repo) ([]types.Repo, error) {
		newRepos := make([]types.Repo, 0, len(repos))
		for _, repo := range repos {
			if r := c(repo); r != nil {
				newRepos = append(newRepos, *r)
			}
		}
		return newRepos, nil
	}, true)
}

// WARNING: Any root-level operations must be performed before this method,
// because process privileges are permanently dropped to the builder user.
func (r *Repos) ModifySlice(ctx context.Context, c func(repos []types.Repo) ([]types.Repo, error), pull bool) error {
	// TODO: Consider rewriting this to use GetSafeConfigModifier or smth else
	// to avoid the side-effect of dropping privileges here.
	// This is may not be ideal, but refactoring is deferred.

	repos := r.cfg.Repos()
	newRepos, err := c(repos)
	if err != nil {
		return err
	}
	r.cfg.SetRepos(newRepos)
	_ = r.cfg.Save("system")

	err = r.sys.DropCapsToBuilderUser()
	if err != nil {
		return err
	}

	if pull {
		for i, repo := range newRepos {
			updatedRepo, err := r.rp.Pull(ctx, repo, &emptyNotfier{})
			if err != nil {
				return errors.WrapIntoI18nError(err, gotext.Get("Error pulling repositories"))
			}
			newRepos[i] = updatedRepo
		}
	}
	return nil
}

func (r *Repos) PullAll(ctx context.Context) error {
	repos := r.cfg.Repos()
	for i, repo := range repos {
		updatedRepo, err := r.pull(ctx, repo)
		if err != nil {
			return err
		}
		repos[i] = updatedRepo
	}
	return nil
}

type emptyNotfier struct{}

func (tn *emptyNotfier) Notify(ctx context.Context, event shared.NotifyEvent, data map[string]string) error {
	return nil
}

func (tn *emptyNotfier) NotifyWrite(ctx context.Context, event shared.NotifyWriterEvent, p []byte) (n int, err error) {
	return len(p), nil
}

var ErrPullRepoInterrupted = stdErrors.New("pullRepo interrupted")

func (r *Repos) pullTui(ctx context.Context, repo types.Repo) (types.Repo, error) {
	m := newPullModel(ctx, repo, r.rp, false)
	w := &progressViewportWriter{}
	m.writer = w
	p := tea.NewProgram(m,
		tea.WithInput(nil),
		tea.WithContext(ctx),
	)
	w.onLine = func(line string, isUpdate bool) {
		p.Send(progressViewportMsg{line: line, isUpdate: isUpdate})
	}
	if _, err := p.Run(); err != nil {
		if stdErrors.Is(err, tea.ErrInterrupted) {
			return m.repo, ErrPullRepoInterrupted
		}
		return m.repo, err
	}
	return m.repo, nil
}

func (r *Repos) pull(ctx context.Context, repo types.Repo) (types.Repo, error) {
	// TODO: different output based on term.IsTerminal(uintptr(syscall.Stdin))
	if term.IsTerminal(uintptr(syscall.Stdin)) {
		return r.pullTui(ctx, repo)
	}

	return r.rp.Pull(ctx, repo, &emptyNotfier{})
}
