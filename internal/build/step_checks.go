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
	"errors"
	"fmt"
	"log/slog"

	"github.com/leonelquinteros/gotext"
)

type checksStep struct {
	e ChecksExecutor
}

func ChecksStep(e ChecksExecutor) *checksStep {
	return &checksStep{e: e}
}

func (s *checksStep) Run(ctx context.Context, state *BuildState) error {
	slog.Info(gotext.Get("Building package"), "name", state.BasePackage)

	for _, pkg := range state.Packages {
		cont, err := s.e.RunChecks(ctx, pkg, state.Input)
		if err != nil {
			return fmt.Errorf("RunChecks failed: %w", err)
		}
		if !cont {
			return errors.New("check step declined continuation")
		}
	}
	return nil
}
