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
	"log/slog"
	"strconv"
	"strings"
	"syscall"

	stdErrors "errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/app/output"
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

type Repos struct {
	cfg *config.ALRConfig
	db  *db.Database
	rp  PullExecutor
	out output.Output
}

func New(cfg *config.ALRConfig, db *db.Database, rp PullExecutor, out output.Output) *Repos {
	return &Repos{
		cfg: cfg,
		db:  db,
		rp:  rp,
		out: out,
	}
}

type simpleNotifier struct {
	out output.Output
}

func (tn *simpleNotifier) Notify(ctx context.Context, event shared.NotifyEvent, data map[string]string) error {
	switch event {
	case puller.EventTryPull:
		i, _ := strconv.Atoi(data["i"])
		url := data["url"]
		var msg string
		if i == 0 {
			msg = gotext.Get("Pull %s", url)
		} else {
			msg = gotext.Get("Trying mirror %d: %s", i, url)
		}
		tn.out.Info(msg)
	case puller.EventErrorPull:
		url := data["url"]
		errMsg := data["err"]
		if errMsg == "" {
			errMsg = "unknown error"
		}
		tn.out.Error("Failed to pull from %s: %v", url, strings.TrimSpace(errMsg))

	default:
		tn.out.Warn("Unknown notify event: %v, data: %v", event, data)
	}
	return nil
}

func (tn *simpleNotifier) NotifyWrite(ctx context.Context, event shared.NotifyWriterEvent, p []byte) (n int, err error) {
	return len(p), nil
}

// Modify updates the configured repositories using the provided callback.
// The updated list is saved into the system configuration.
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

func (r *Repos) ModifySlice(ctx context.Context, c func(repos []types.Repo) ([]types.Repo, error), pull bool) error {
	repos := r.cfg.Repos()
	newRepos, err := c(repos)
	if err != nil {
		return err
	}
	r.cfg.SetRepos(newRepos)
	_ = r.cfg.Save("system")

	if pull {
		if err := r.pullRepos(ctx, newRepos); err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error pulling repositories"))
		}
	}

	return nil
}

func (r *Repos) pullRepos(ctx context.Context, repos []types.Repo) error {
	for i, repo := range repos {
		if repo.Disabled {
			err := r.deleteRepoPkgs(ctx, repo.Name)
			if err != nil {
				slog.Warn("failed to remove packages from disabled repo", "err", err)
			}
			continue
		}
		updatedRepo, err := r.Pull(ctx, repo)
		if err != nil {
			return err
		}
		repos[i] = updatedRepo
	}
	return nil
}

func (r *Repos) PullAll(ctx context.Context) error {
	return r.pullRepos(ctx, r.cfg.Repos())
}

type rereadOption func(*rereadConfig)

type rereadConfig struct {
	deleteFailed bool
}

func WithDeleteFailed() rereadOption {
	return func(cfg *rereadConfig) {
		cfg.deleteFailed = true
	}
}

func (r *Repos) deleteRepoPkgs(ctx context.Context, name string) error {
	if err := r.db.DeletePkgs(ctx, "repository = ?", name); err != nil {
		return fmt.Errorf("failed to delete repo packages %q: %w", name, err)
	}
	return nil
}

func (r *Repos) RereadAll(ctx context.Context, opts ...rereadOption) error {
	cfg := &rereadConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	repos := r.cfg.Repos()
	for i, repo := range repos {
		if repo.Disabled {
			if delErr := r.deleteRepoPkgs(ctx, repo.Name); delErr != nil {
				return delErr
			}
			continue
		}
		updatedRepo, err := r.rp.Read(ctx, repo, &simpleNotifier{r.out})
		if err != nil {
			if cfg.deleteFailed {
				if delErr := r.deleteRepoPkgs(ctx, repo.Name); delErr != nil {
					return delErr
				}
				continue
			}
			return err
		}
		repos[i] = updatedRepo
	}
	return nil
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
			return m.state.repo, ErrPullRepoInterrupted
		}
		return m.state.repo, err
	}
	return m.state.repo, nil
}

func (r *Repos) Pull(ctx context.Context, repo types.Repo) (types.Repo, error) {
	if term.IsTerminal(uintptr(syscall.Stdin)) {
		return r.pullTui(ctx, repo)
	}
	repo, err := r.rp.Pull(ctx, repo, &simpleNotifier{r.out})
	if err != nil {
		return repo, err
	}
	r.out.Info(gotext.Get("Repository pulled successfully!"))
	return repo, err
}
