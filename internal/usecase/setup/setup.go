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

package setup

import (
	"context"
	"fmt"
	"os/exec"
	"os/user"
	"strings"

	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/errors"
	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/subid"
)

const (
	subIDMinStart = 100000
	subIDCount    = 65536
)

type subIDKind struct {
	file string
	flag string
}

var subIDKinds = []subIDKind{
	{file: subid.UIDFile, flag: "--add-subuids"},
	{file: subid.GIDFile, flag: "--add-subgids"},
}

type useCase struct {
	out      output.Output
	username string
	kinds    []subIDKind
	addRange func(ctx context.Context, flag, username string, start, end uint64) error
	lookPath func(string) (string, error)
}

func New(out output.Output) *useCase {
	return &useCase{
		out:      out,
		username: constants.BuilderUser,
		kinds:    subIDKinds,
		addRange: usermodAddRange,
		lookPath: exec.LookPath,
	}
}

// Run is idempotent so it can be called from package scripts on every
// install and upgrade.
func (u *useCase) Run(ctx context.Context) error {
	usr, err := user.Lookup(u.username)
	if err != nil {
		return errors.WrapIntoI18nError(err, gotext.Get("User %s not found", u.username))
	}

	for _, kind := range u.kinds {
		if err := u.ensureSubIDs(ctx, kind, usr.Uid); err != nil {
			return errors.WrapIntoI18nError(err, gotext.Get("Failed to configure %s", kind.file))
		}
	}

	for _, bin := range []string{"newuidmap", "newgidmap"} {
		if _, err := u.lookPath(bin); err != nil {
			u.out.Warn("%s", gotext.Get("%s not found, builds run by a regular user can't change file ownership", bin))
		}
	}

	return nil
}

func (u *useCase) ensureSubIDs(ctx context.Context, kind subIDKind, uid string) error {
	ranges, err := subid.ReadFile(kind.file)
	if err != nil {
		return err
	}

	if len(subid.OwnedBy(ranges, u.username, uid)) > 0 {
		return nil
	}

	start, ok := subid.NextFree(ranges, subIDMinStart, subIDCount)
	if !ok {
		return fmt.Errorf("no free range of %d ids left", subIDCount)
	}

	end := start + subIDCount - 1
	if err := u.addRange(ctx, kind.flag, u.username, start, end); err != nil {
		return err
	}

	u.out.Info("%s", gotext.Get("Added range %d-%d to %s for %s", start, end, kind.file, u.username))
	return nil
}

func usermodAddRange(ctx context.Context, flag, username string, start, end uint64) error {
	cmd := exec.CommandContext(ctx, "usermod", flag, fmt.Sprintf("%d-%d", start, end), username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("usermod: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
