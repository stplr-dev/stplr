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
	"os"
	"os/user"
	"strconv"
	"syscall"

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

func DropCapsToBuilderUser() error {
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

func IsNotRoot() bool {
	return os.Getuid() != 0
}

func IsRoot() bool {
	return !IsNotRoot()
}

func IsBuilderUser() (bool, error) {
	uid, gid, err := GetUidGidStaplerUser()
	if err != nil {
		return false, err
	}
	newUid := syscall.Getuid()
	newGid := syscall.Getgid()
	if newUid != uid || newGid != gid {
		return false, nil
	}

	return true, nil
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
