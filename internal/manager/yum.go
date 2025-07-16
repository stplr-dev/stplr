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

// YUM represents the YUM package manager
type YUM struct {
	CommonPackageManager
	CommonRPM
}

func NewYUM() *YUM {
	return &YUM{
		CommonPackageManager: CommonPackageManager{
			noConfirmArg: "-y",
		},
	}
}

func (*YUM) Exists() bool {
	_, err := exec.LookPath("yum")
	return err == nil
}

func (*YUM) Name() string {
	return "yum"
}

func (*YUM) Format() string {
	return "rpm"
}

func (y *YUM) Sync(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := y.getCmd(opts, "yum", "upgrade")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("yum: sync: %w", err)
	}
	return nil
}

func (y *YUM) Install(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := y.getCmd(opts, "yum", "install", "--allowerasing")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("yum: install: %w", err)
	}
	return nil
}

func (y *YUM) InstallLocal(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	return y.Install(opts, pkgs...)
}

func (y *YUM) Remove(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := y.getCmd(opts, "yum", "remove")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("yum: remove: %w", err)
	}
	return nil
}

func (y *YUM) Upgrade(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := y.getCmd(opts, "yum", "upgrade")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("yum: upgrade: %w", err)
	}
	return nil
}

func (y *YUM) UpgradeAll(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := y.getCmd(opts, "yum", "upgrade")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("yum: upgradeall: %w", err)
	}
	return nil
}
