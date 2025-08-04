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
	"log/slog"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/cliutils"
	"go.stplr.dev/stplr/internal/cpu"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type ChecksRunner struct {
	mgr manager.Manager
}

func NewChecksRunner(mgr manager.Manager) *ChecksRunner {
	return &ChecksRunner{mgr: mgr}
}

func (r *ChecksRunner) RunChecks(ctx context.Context, pkg *staplerfile.Package, input *BuildInput) (bool, error) {
	if !cpu.IsCompatibleWith(cpu.Arch(), pkg.Architectures) {
		cont, err := cliutils.YesNoPrompt(
			ctx,
			gotext.Get("Your system's CPU architecture doesn't match this package. Do you want to build anyway?"),
			input.Opts.Interactive,
			true,
		)
		if err != nil {
			return false, err
		}
		if !cont {
			return false, nil
		}
	}

	installed, err := r.mgr.ListInstalled(nil)
	if err != nil {
		return false, err
	}

	filename, err := pkgFileName(input, pkg)
	if err != nil {
		return false, err
	}

	if instVer, ok := installed[filename]; ok {
		slog.Warn(gotext.Get("This package is already installed"),
			"name", pkg.Name,
			"version", instVer,
		)
	}

	return true, nil
}

type ChecksExecutor interface {
	RunChecks(ctx context.Context, pkg *staplerfile.Package, input *BuildInput) (bool, error)
}
