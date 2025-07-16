// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package manager

import (
	"fmt"
	"os/exec"
)

// Zypper represents the Zypper package manager
type Zypper struct {
	CommonPackageManager
	CommonRPM
}

func NewZypper() *YUM {
	return &YUM{
		CommonPackageManager: CommonPackageManager{
			noConfirmArg: "-y",
		},
	}
}

func (*Zypper) Exists() bool {
	_, err := exec.LookPath("zypper")
	return err == nil
}

func (*Zypper) Name() string {
	return "zypper"
}

func (*Zypper) Format() string {
	return "rpm"
}

func (z *Zypper) Sync(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := z.getCmd(opts, "zypper", "refresh")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("zypper: sync: %w", err)
	}
	return nil
}

func (z *Zypper) Install(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := z.getCmd(opts, "zypper", "install", "-y")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("zypper: install: %w", err)
	}
	return nil
}

func (z *Zypper) InstallLocal(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	return z.Install(opts, pkgs...)
}

func (z *Zypper) Remove(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := z.getCmd(opts, "zypper", "remove", "-y")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("zypper: remove: %w", err)
	}
	return nil
}

func (z *Zypper) Upgrade(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := z.getCmd(opts, "zypper", "update", "-y")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("zypper: upgrade: %w", err)
	}
	return nil
}

func (z *Zypper) UpgradeAll(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := z.getCmd(opts, "zypper", "update", "-y")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("zypper: upgradeall: %w", err)
	}
	return nil
}
