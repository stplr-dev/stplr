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

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/pkg/staplerfile"
)

func ViewNonfree(
	ctx context.Context,
	input *BuildInput,
	vars *staplerfile.Package,
) error {
	if !input.opts.Interactive {
		return nil
	}

	if !vars.NonFree {
		return nil
	}

	var content string
	var err error

	msgFile := vars.NonFreeMsgFile.Resolved()
	msg := vars.NonFreeMsg.Resolved()

	switch {
	case msgFile != "":
		contentBytes, err := os.ReadFile(msgFile)
		if err != nil {
			return fmt.Errorf("failed to read nonfree message file: %w", err)
		}
		content = string(contentBytes)
	case msg != "":
		content = msg
	default:
		content = gotext.Get("This package contains non-free software that requires license acceptance.")
	}

	pager := NewNonfree(vars.Name, content, vars.NonFreeUrl.Resolved())
	accepted, err := pager.Run()
	if err != nil {
		return fmt.Errorf("failed to display EULA: %w", err)
	}

	if !accepted {
		return fmt.Errorf("license agreement was declined")
	}

	return nil
}
