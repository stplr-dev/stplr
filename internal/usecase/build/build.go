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

package build

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/app/output"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
)

type system interface {
	IsRoot() bool
	Getuid() int
	Getgid() int
}

type useCase struct {
	sys system
	out output.Output
}

func New(sys system, out output.Output) *useCase {
	return &useCase{sys, out}
}

type Options struct {
	Script      string
	Subpackage  string
	Package     string
	Clean       bool
	Interactive bool
	NoSuffix    bool
}

func (u *useCase) Run(ctx context.Context, opts Options) error {
	if opts.Script == "" && opts.Package == "" {
		return errors.NewI18nError(gotext.Get("Nothing to build"))
	}

	state, err := u.prepareState(ctx, opts)
	if err != nil {
		return err
	}
	defer state.Cleanup()

	steps := u.getSteps(ctx, state, opts)
	for _, s := range steps {
		slog.Debug("execute step", "step", fmt.Sprintf("%T", s))
		err := s.Execute(ctx, state)
		if err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Error building package"))
		}
	}

	u.out.Info(gotext.Get("Done"))

	return nil
}

func (u *useCase) getSteps(ctx context.Context, state *stepState, opts Options) []step {
	var steps []step

	steps = append(steps, &checkStep{})

	if opts.Package != "" {
		steps = append(steps, &prepareDbArgs{state.input.deps.Repos})
	} else {
		steps = append(steps, &prepareScriptArgs{})
	}

	copyOut := &copyOutStep{}
	if u.sys.IsRoot() {
		steps = append(steps, &setupCopier{})
		if opts.Script != "" {
			steps = append(steps, &copyScript{fsys: os.DirFS("/")})
		}
		copyOut.copy = copyOutViaCopier
	} else {
		steps = append(steps, &modifyCfgPaths{})
		copyOut.copy = copyOutViaOsutils
	}

	steps = append(steps,
		&prepareInstallerAndScripterStep{},
		&buildStep{},
		copyOut,
	)

	return steps
}

func (u *useCase) prepareState(ctx context.Context, opts Options) (*stepState, error) {
	state := &stepState{}
	input := &state.input
	runtime := &state.runtime

	input.opts = opts
	input.origUid = u.sys.Getuid()
	input.origGid = u.sys.Getgid()

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	input.wd = wd

	deps, err := appbuilder.
		New(ctx).
		WithConfig().
		WithDB().
		WithReposNoPull().
		WithDistroInfo().
		WithManager().
		Build()
	if err != nil {
		return nil, err
	}
	runtime.cleanups = append(runtime.cleanups, deps.Defer)
	input.deps = deps

	return state, nil
}
