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

package manager

import (
	"bufio"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

// Epm represents the EPM package manager
type Epm struct {
	CommonPackageManager
}

func NewEpm() *Epm {
	return &Epm{
		CommonPackageManager: CommonPackageManager{
			noConfirmArg: "-y",
		},
	}
}

func (e *Epm) Exists() bool {
	_, err := exec.LookPath("epm")
	if err != nil {
		return false
	}

	return slices.Contains(SupportedPackageFormats(), e.Format())
}

func (*Epm) Name() string {
	return "epm"
}

func (*Epm) format() string {
	cmd := exec.Command("epm", "print", "info", "-p")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

func (e *Epm) Format() string {
	f := e.format()

	if slices.Contains([]string{"pkg.tar.zst", "pkg.tar.xz"}, f) {
		f = "archlinux"
	}

	return f
}

func (p *Epm) Sync(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "epm", "update")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("epm: update: %w", err)
	}
	return nil
}

func (p *Epm) Install(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "epm", "install")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("epm: install: %w", err)
	}
	return nil
}

func (p *Epm) InstallLocal(opts *Opts, pkgs ...string) error {
	return p.Install(opts, pkgs...)
}

func (p *Epm) Remove(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "epm", "remove")
	cmd.Args = append(cmd.Args, pkgs...)
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("epm: remove: %w", err)
	}
	return nil
}

func (p *Epm) Upgrade(opts *Opts, pkgs ...string) error {
	opts = ensureOpts(opts)
	return p.Install(opts, pkgs...)
}

func (p *Epm) UpgradeAll(opts *Opts) error {
	opts = ensureOpts(opts)
	cmd := p.getCmd(opts, "epm", "upgrade")
	setCmdEnv(cmd)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("epm: upgrade: %w", err)
	}
	return nil
}

func (p *Epm) ListInstalled(opts *Opts) (map[string]string, error) {
	out := map[string]string{}
	cmd := exec.Command("epm", "list-installed")

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

func (p *Epm) IsInstalled(pkg string) (bool, error) {
	cmd := exec.Command("epm", "query", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// epm returns exit code 1 if the package is not found
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("epm: query: %w, output: %s", err, output)
	}
	return true, nil
}
