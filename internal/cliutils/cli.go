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

package cliutils

import (
	"fmt"
	"os/user"
	"syscall"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v3"

	"go.stplr.dev/stplr/internal/utils"

	"go.stplr.dev/stplr/internal/constants"
)

func ExitIfCantDropGidToStapler() cli.ExitCoder {
	_, gid, err := utils.GetUidGidStaplerUser()
	if err != nil {
		return FormatCliExit(fmt.Sprintf("cannot get gid %s", constants.BuilderUser), err)
	}
	err = syscall.Setgid(gid)
	if err != nil {
		return FormatCliExit(fmt.Sprintf("cannot get setgid %s", constants.BuilderUser), err)
	}
	return nil
}

// ExitIfCantDropCapsToBuilderUser attempts to drop capabilities to the already
// running user. Returns a cli.ExitCoder with an error if the operation fails.
// See also [ExitIfCantDropCapsToBuilderUserNoPrivs] for a version that also applies
// no-new-privs.
func ExitIfCantDropCapsToBuilderUser() cli.ExitCoder {
	err := utils.DropCapsToBuilderUser()
	if err != nil {
		return FormatCliExit(gotext.Get("Error on dropping capabilities"), err)
	}
	return nil
}

func ExitIfCantSetNoNewPrivs() cli.ExitCoder {
	if err := utils.NoNewPrivs(); err != nil {
		return FormatCliExit("error on NoNewPrivs", err)
	}

	return nil
}

// ExitIfCantDropCapsToBuilderUserNoPrivs combines [ExitIfCantDropCapsToBuilderUser] with [ExitIfCantSetNoNewPrivs]
func ExitIfCantDropCapsToBuilderUserNoPrivs() cli.ExitCoder {
	if err := ExitIfCantDropCapsToBuilderUser(); err != nil {
		return err
	}

	if err := ExitIfCantSetNoNewPrivs(); err != nil {
		return err
	}

	return nil
}

func ExitIfRootCantDropCapsNoPrivs() cli.ExitCoder {
	if utils.IsNotRoot() {
		return nil
	}

	if err := ExitIfCantDropCapsToBuilderUser(); err != nil {
		return err
	}

	if err := ExitIfCantSetNoNewPrivs(); err != nil {
		return err
	}

	return nil
}

func EnsureIsPrivilegedGroupMemberOrRoot() error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	if currentUser.Uid == "0" {
		return nil
	}

	group, err := user.LookupGroup(constants.PrivilegedGroup)
	if err != nil {
		return err
	}

	groups, err := currentUser.GroupIds()
	if err != nil {
		return err
	}

	for _, gid := range groups {
		if gid == group.Gid {
			return nil
		}
	}

	return FormatCliExit(
		gotext.Get("You need to be a %s member or root to perform this action", constants.PrivilegedGroup),
		nil,
	)
}
