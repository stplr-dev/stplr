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

package build

import (
	"context"

	"go.stplr.dev/stplr/internal/manager"
)

func NewInstaller(mgr manager.Manager, needRootCmd bool, rootCmd string) *Installer {
	return &Installer{
		mgr:         mgr,
		rootCmd:     rootCmd,
		needRootCmd: needRootCmd,
	}
}

type Installer struct {
	mgr         manager.Manager
	rootCmd     string
	needRootCmd bool
}

func (i *Installer) modifyOpts(opts *manager.Opts) *manager.Opts {
	if opts == nil {
		opts = &manager.Opts{}
	}
	if i.needRootCmd {
		opts.AsRoot = true
		if opts.RootCmd == "" {
			opts.RootCmd = i.rootCmd
		}
	}
	return opts
}

func (i *Installer) InstallLocal(ctx context.Context, paths []string, opts *manager.Opts) error {
	return i.mgr.InstallLocal(i.modifyOpts(opts), paths...)
}

func (i *Installer) Install(ctx context.Context, pkgs []string, opts *manager.Opts) error {
	return i.mgr.Install(i.modifyOpts(opts), pkgs...)
}

func (i *Installer) Remove(ctx context.Context, pkgs []string, opts *manager.Opts) error {
	return i.mgr.Remove(i.modifyOpts(opts), pkgs...)
}

func (i *Installer) RemoveAlreadyInstalled(ctx context.Context, pkgs []string) ([]string, error) {
	filteredPackages := []string{}

	for _, dep := range pkgs {
		installed, err := i.mgr.IsInstalled(dep)
		if err != nil {
			return nil, err
		}
		if installed {
			continue
		}
		filteredPackages = append(filteredPackages, dep)
	}

	return filteredPackages, nil
}
