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

package savers

import (
	"fmt"
	"os"

	"go.stplr.dev/stplr/internal/constants"
)

type SystemConfigWriter struct{}

func (s *SystemConfigWriter) Write(b []byte) (int, error) {
	file, err := os.Create(constants.SystemConfigPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	n, err := file.Write(b)
	if err != nil {
		return 0, fmt.Errorf("failed to write config: %w", err)
	}

	if err := file.Sync(); err != nil {
		return 0, fmt.Errorf("failed to sync config: %w", err)
	}

	return n, nil
}

var _ SystemConfigWriterExecutor = (*SystemConfigWriter)(nil)
