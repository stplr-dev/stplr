// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "ALR - Any Linux Repository"
// created by the ALR Authors.
// It was later modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) 2025 The ALR Authors
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

package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/leonelquinteros/gotext"
	"github.com/urfave/cli/v2"

	"go.stplr.dev/stplr/internal/cliutils"
	appbuilder "go.stplr.dev/stplr/internal/cliutils/app_builder"
	"go.stplr.dev/stplr/internal/constants"
)

func GetUidGidBuilderUserString() (string, string, error) {
	u, err := user.Lookup(constants.BuilderUser)
	if err != nil {
		return "", "", err
	}

	return u.Uid, u.Gid, nil
}

func GetUidGidStaplerUser() (int, int, error) {
	strUid, strGid, err := GetUidGidBuilderUserString()
	if err != nil {
		return 0, 0, err
	}

	uid, err := strconv.Atoi(strUid)
	if err != nil {
		return 0, 0, err
	}
	gid, err := strconv.Atoi(strGid)
	if err != nil {
		return 0, 0, err
	}

	return uid, gid, nil
}

func dropCapsToBuilderUser() error {
	uid, gid, err := GetUidGidStaplerUser()
	if err != nil {
		return err
	}
	err = syscall.Setgid(gid)
	if err != nil {
		return err
	}
	err = syscall.Setuid(uid)
	if err != nil {
		return err
	}
	return EnsureIsBuilderUser()
}

func ExitIfCantDropGidToStapler() cli.ExitCoder {
	_, gid, err := GetUidGidStaplerUser()
	if err != nil {
		return cliutils.FormatCliExit(fmt.Sprintf("cannot get gid %s", constants.BuilderUser), err)
	}
	err = syscall.Setgid(gid)
	if err != nil {
		return cliutils.FormatCliExit(fmt.Sprintf("cannot get setgid %s", constants.BuilderUser), err)
	}
	return nil
}

// ExitIfCantDropCapsToBuilderUser attempts to drop capabilities to the already
// running user. Returns a cli.ExitCoder with an error if the operation fails.
// See also [ExitIfCantDropCapsToBuilderUserNoPrivs] for a version that also applies
// no-new-privs.
func ExitIfCantDropCapsToBuilderUser() cli.ExitCoder {
	err := dropCapsToBuilderUser()
	if err != nil {
		return cliutils.FormatCliExit(gotext.Get("Error on dropping capabilities"), err)
	}
	return nil
}

func ExitIfCantSetNoNewPrivs() cli.ExitCoder {
	if err := NoNewPrivs(); err != nil {
		return cliutils.FormatCliExit("error on NoNewPrivs", err)
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
	if IsNotRoot() {
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

func IsNotRoot() bool {
	return os.Getuid() != 0
}

func IsRoot() bool {
	return !IsNotRoot()
}

func EnsureIsBuilderUser() error {
	uid, gid, err := GetUidGidStaplerUser()
	if err != nil {
		return err
	}
	newUid := syscall.Getuid()
	if newUid != uid {
		return errors.New("uid don't matches requested")
	}
	newGid := syscall.Getgid()
	if newGid != gid {
		return errors.New("gid don't matches requested")
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

	return cliutils.FormatCliExit(
		gotext.Get("You need to be a %s member or root to perform this action", constants.PrivilegedGroup),
		nil,
	)
}

func EscalateToRootGid() error {
	return syscall.Setgid(0)
}

func EscalateToRootUid() error {
	return syscall.Setuid(0)
}

func EscalateToRoot() error {
	err := EscalateToRootUid()
	if err != nil {
		return err
	}
	err = EscalateToRootGid()
	if err != nil {
		return err
	}
	return nil
}

func RootNeededAction(f cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		deps, err := appbuilder.
			New(ctx.Context).
			WithConfig().
			Build()
		if err != nil {
			return err
		}
		defer deps.Defer()

		if IsNotRoot() {
			if !deps.Cfg.UseRootCmd() {
				return cli.Exit(gotext.Get("You need to be root to perform this action"), 1)
			}
			executable, err := os.Executable()
			if err != nil {
				return cliutils.FormatCliExit("failed to get executable path", err)
			}
			args := append([]string{executable}, os.Args[1:]...)

			// Looks like no risk of injection
			//gosec:disable G204
			cmd := exec.Command(deps.Cfg.RootCmd(), args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		return f(ctx)
	}
}

func ReadonlyAction(f cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		if IsNotRoot() {
			// TODO: relaunch in userns with HOME hide
		} else {
			if err := ExitIfCantDropCapsToBuilderUserNoPrivs(); err != nil {
				return err
			}
		}

		return f(ctx)
	}
}
