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

package sys

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"go.stplr.dev/stplr/internal/constants"
)

type Sys struct{}

func (s Sys) IsRoot() bool {
	return os.Getuid() == 0
}

func (s Sys) Getuid() int {
	return os.Getuid()
}

func (s Sys) Getgid() int {
	return os.Getgid()
}

func mustInt(s string) int {
	if i, err := strconv.Atoi(s); err != nil {
		panic(err)
	} else {
		return i
	}
}

func (s Sys) getBuilderUser() (*user.User, error) {
	return user.Lookup(constants.BuilderUser)
}

func (s Sys) DropCapsToBuilderUser() error {
	u, err := s.getBuilderUser()
	if err != nil {
		return fmt.Errorf("failed get builder user: %w", err)
	}

	err = syscall.Setgid(mustInt(u.Gid))
	if err != nil {
		return err
	}
	err = syscall.Setuid(mustInt(u.Uid))
	if err != nil {
		return err
	}

	u, err = s.getBuilderUser()
	if err != nil {
		return fmt.Errorf("failed get builder user after drop: %w", err)
	}

	uid := s.Getuid()
	gid := s.Getgid()

	if uid != mustInt(u.Uid) || gid != mustInt(u.Gid) {
		return fmt.Errorf(
			"failed to drop caps to builder user: %s:%s != %d:%d",
			u.Uid, u.Gid,
			uid, gid,
		)
	}

	return nil
}
