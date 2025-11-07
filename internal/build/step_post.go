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

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/cliprompts"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/installer"
	"go.stplr.dev/stplr/internal/manager"
)

type postStep struct {
	installerExecutor installer.InstallerExecutor
}

func PostStep(installerExecutor installer.InstallerExecutor) *postStep {
	return &postStep{installerExecutor: installerExecutor}
}

func (s *postStep) Run(ctx context.Context, state *BuildState) error {
	err := s.removeBuildDeps(ctx, state.Input, state.InstalledBuildDeps)

	return err
}

func (s *postStep) removeBuildDeps(ctx context.Context, input interface {
	commonbuild.BuildOptsProvider
}, deps []string,
) error {
	if len(deps) > 0 {
		remove, err := cliprompts.YesNoPrompt(ctx, gotext.Get("Would you like to remove the build dependencies?"), input.BuildOpts().Interactive, false)
		if err != nil {
			return err
		}

		if remove {
			err = s.installerExecutor.Remove(
				ctx,
				deps,
				&manager.Opts{
					NoConfirm: !input.BuildOpts().Interactive,
				},
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
