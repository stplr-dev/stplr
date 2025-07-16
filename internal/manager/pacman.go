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
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// Pacman represents the Pacman package manager
type Pacman struct {
	CommonPackageManager
}

func NewPacman() *Pacman {
	return &Pacman{
		CommonPackageManager: CommonPackageManager{
			noConfirmArg: "--noconfirm",
		},
	}
}

func (*Pacman) Exists() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

func (*Pacman) Name() string {
	return "pacman"
}

func (*Pacman) Format() string {
	return "archlinux"
}

func (p *Pacman) Sync(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "pacman", "-Sy")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pacman: sync: %w", err)
	}
	return nil
}

func (p *Pacman) Install(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "pacman", "-S", "--needed")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pacman: install: %w", err)
	}
	return nil
}

func (p *Pacman) InstallLocal(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "pacman", "-U", "--needed")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pacman: installlocal: %w", err)
	}
	return nil
}

func (p *Pacman) Remove(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "pacman", "-R")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pacman: remove: %w", err)
	}
	return nil
}

func (p *Pacman) Upgrade(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	return p.Install(opts, pkgs...)
}

func (p *Pacman) UpgradeAll(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "pacman", "-Su")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pacman: upgradeall: %w", err)
	}
	return nil
}

func (p *Pacman) ListInstalled(opts *Opts) (map[string]string, error) {
	out := map[string]string{}
	cmd := exec.Command("pacman", "-Q")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		name, version, ok := strings.Cut(scanner.Text(), " ")
		if !ok {
			continue
		}
		out[name] = version
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (p *Pacman) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("pacman", "-Q", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Pacman returns exit code 1 if the package is not found
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("pacman: isinstalled: %w, output: %s", err, output)
	}
	return true, nil
}
