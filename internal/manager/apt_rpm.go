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

package manager

import (
	"fmt"
	"os/exec"
	"strings"
)

// APTRpm represents the APT-RPM package manager
type APTRpm struct {
	CommonPackageManager
	CommonRPM
}

func NewAPTRpm() *APTRpm {
	return &APTRpm{
		CommonPackageManager: CommonPackageManager{
			noConfirmArg: "-y",
		},
	}
}

func (*APTRpm) Name() string {
	return "apt-rpm"
}

func (*APTRpm) Format() string {
	return "rpm"
}

func (*APTRpm) Exists() bool {
	cmd := exec.Command("apt-config", "dump")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "RPM")
}

func (a *APTRpm) Sync(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := a.getCmd(opts, "apt-get", "update")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("apt-get: sync: %w", err)
	}
	return nil
}

func (a *APTRpm) Install(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := a.getCmd(opts, "apt-get", "install", "-o", "APT::Install::Virtual=true")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	cmd.Stdout = cmd.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("apt-get: install: %w", err)
	}
	return nil
}

func (a *APTRpm) InstallLocal(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	return a.Install(opts, pkgs...)
}

func (a *APTRpm) Remove(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := a.getCmd(opts, "apt-get", "remove")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("apt-get: remove: %w", err)
	}
	return nil
}

func (a *APTRpm) Upgrade(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	return a.Install(opts, pkgs...)
}

func (a *APTRpm) UpgradeAll(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := a.getCmd(opts, "apt-get", "dist-upgrade")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("apt-get: upgradeall: %w", err)
	}
	return nil
}
