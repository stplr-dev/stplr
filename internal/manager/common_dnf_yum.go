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

package manager

import "fmt"

type commonDNFYUM struct {
	CommonPackageManager
	CommonRPM
	binary string
}

func (m *commonDNFYUM) Sync(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := m.getCmd(opts, m.binary, "upgrade")
	setCmdEnv(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: sync: %w", m.binary, err)
	}
	return nil
}

func (m *commonDNFYUM) Install(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := m.getCmd(opts, m.binary, "install", "--allowerasing")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: install: %w", m.binary, err)
	}
	return nil
}

func (m *commonDNFYUM) InstallLocal(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	return m.Install(opts, pkgs...)
}

func (m *commonDNFYUM) Remove(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := m.getCmd(opts, m.binary, "remove")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: remove: %w", m.binary, err)
	}
	return nil
}

func (m *commonDNFYUM) Upgrade(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := m.getCmd(opts, m.binary, "upgrade")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: upgrade: %w", m.binary, err)
	}
	return nil
}

func (m *commonDNFYUM) UpgradeAll(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := m.getCmd(opts, m.binary, "upgrade")
	setCmdEnv(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: upgradeall: %w", m.binary, err)
	}
	return nil
}
