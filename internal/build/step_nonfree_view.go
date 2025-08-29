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
	"os"
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/tui/pager"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type NonFreeViewer struct{ cfg Config }

func NewNonFreeViewer(cfg Config) *NonFreeViewer {
	return &NonFreeViewer{cfg: cfg}
}

func resolveUnderBase(baseDir, rel string) (string, error) {
	joined := filepath.Join(baseDir, rel)

	evaluated, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate symlinks: %w", err)
	}

	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}

	absTarget, err := filepath.Abs(evaluated)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	if !strings.HasPrefix(absTarget+string(os.PathSeparator), absBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path: %s is outside of base %s", absTarget, absBase)
	}

	return absTarget, nil
}

func (v *NonFreeViewer) ViewNonfree(
	ctx context.Context,
	pkg *staplerfile.Package,
	scriptPath string,
	interactive bool,
) error {
	if !interactive {
		return nil
	}

	if !pkg.NonFree {
		return nil
	}

	var content string
	var err error

	msgFile := pkg.NonFreeMsgFile.Resolved()
	msg := pkg.NonFreeMsg.Resolved()

	switch {
	case msgFile != "":
		resolvedFile, err := resolveUnderBase(filepath.Dir(scriptPath), msgFile)
		if err != nil {
			return fmt.Errorf("invalid nonfree message file path: %w", err)
		}
		contentBytes, err := os.ReadFile(resolvedFile)
		if err != nil {
			return fmt.Errorf("failed to read nonfree message file: %w", err)
		}
		content = string(contentBytes)
	case msg != "":
		content = msg
	default:
		content = gotext.Get("This package contains non-free software that requires license acceptance.")
	}

	pager := pager.NewNonfree(pkg.Name, content, pkg.NonFreeUrl.Resolved())
	accepted, err := pager.Run()
	if err != nil {
		return fmt.Errorf("failed to display nonfree: %w", err)
	}

	if !accepted {
		return fmt.Errorf("license agreement was declined")
	}

	return nil
}

type NonFreeViewerExecutor interface {
	ViewNonfree(ctx context.Context, pkg *staplerfile.Package, scriptPath string, interactive bool) error
}

type nonfreeViewStep struct {
	e NonFreeViewerExecutor
}

func NonfreeViewStep(e NonFreeViewerExecutor) *nonfreeViewStep {
	return &nonfreeViewStep{e: e}
}

func (s *nonfreeViewStep) Run(ctx context.Context, state *BuildState) error {
	for _, pkg := range state.Packages {
		if err := s.e.ViewNonfree(
			ctx,
			pkg,
			state.ScriptFile.Path(),
			state.Input.Opts.Interactive,
		); err != nil {
			return err
		}
	}
	return nil
}
