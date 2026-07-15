// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.stplr.dev/stplr/internal/copier"
)

// PreCopySourceOverrides copies local file paths from source overrides into
// a temporary location accessible to the build user. Non-local paths and URLs
// are passed through unchanged.
func PreCopySourceOverrides(ctx context.Context, c copier.CopierExecutor, overrides map[int]string) (map[int]string, func(), error) {
	noop := func() {}
	if len(overrides) == 0 {
		return overrides, noop, nil
	}

	result := make(map[int]string, len(overrides))
	var tempDirs []string
	cleanup := func() {
		for _, dir := range tempDirs {
			if err := os.RemoveAll(dir); err != nil {
				slog.Warn("failed to cleanup source override", "dir", dir, "err", err)
			}
		}
	}

	for idx, path := range overrides {
		if !strings.HasPrefix(path, "local:///") {
			result[idx] = path
			continue
		}

		u, err := url.Parse(path)
		if err != nil {
			cleanup()
			return nil, noop, fmt.Errorf("source override [%d] %q: %w", idx, path, err)
		}

		newPath, err := c.CopySourceFile(ctx, u.Path)
		if err != nil {
			cleanup()
			return nil, noop, fmt.Errorf("source override [%d] %q: %w", idx, path, err)
		}

		newURL := "local:///" + newPath
		if u.RawQuery != "" {
			newURL += "?" + u.RawQuery
		}
		result[idx] = newURL
		tempDirs = append(tempDirs, filepath.Dir(newPath))
	}

	return result, cleanup, nil
}
