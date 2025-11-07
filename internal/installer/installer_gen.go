// DO NOT EDIT MANUALLY. This file is generated.
//
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

package installer

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/keegancsmith/rpc"
	"go.stplr.dev/stplr/internal/manager"
)

type InstallerExecutorPlugin struct {
	Impl InstallerExecutor
}
type InstallerExecutorRPCServer struct {
	Impl   InstallerExecutor
	broker *plugin.MuxBroker
}
type InstallerExecutorRPC struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

func (p *InstallerExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &InstallerExecutorRPC{
		broker: b,
		client: c,
	}, nil
}
func (p *InstallerExecutorPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &InstallerExecutorRPCServer{
		Impl:   p.Impl,
		broker: b,
	}, nil
}

type InstallerExecutorInstallArgs struct {
	Pkgs []string
	Opts *manager.Opts
}
type InstallerExecutorInstallResp struct{}

func (s *InstallerExecutorRPC) Install(ctx context.Context, pkgs []string, opts *manager.Opts) error {
	var resp *InstallerExecutorInstallResp
	err := s.client.Call(ctx, "Plugin.Install", &InstallerExecutorInstallArgs{Pkgs: pkgs, Opts: opts}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *InstallerExecutorRPCServer) Install(ctx context.Context, args *InstallerExecutorInstallArgs, resp *InstallerExecutorInstallResp) error {
	var err error
	err = s.Impl.Install(ctx, args.Pkgs, args.Opts)
	if err != nil {
		return err
	}
	*resp = InstallerExecutorInstallResp{}
	return nil
}

type InstallerExecutorInstallLocalArgs struct {
	Paths []string
	Opts  *manager.Opts
}
type InstallerExecutorInstallLocalResp struct{}

func (s *InstallerExecutorRPC) InstallLocal(ctx context.Context, paths []string, opts *manager.Opts) error {
	var resp *InstallerExecutorInstallLocalResp
	err := s.client.Call(ctx, "Plugin.InstallLocal", &InstallerExecutorInstallLocalArgs{Paths: paths, Opts: opts}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *InstallerExecutorRPCServer) InstallLocal(ctx context.Context, args *InstallerExecutorInstallLocalArgs, resp *InstallerExecutorInstallLocalResp) error {
	var err error
	err = s.Impl.InstallLocal(ctx, args.Paths, args.Opts)
	if err != nil {
		return err
	}
	*resp = InstallerExecutorInstallLocalResp{}
	return nil
}

type InstallerExecutorRemoveArgs struct {
	Pkgs []string
	Opts *manager.Opts
}
type InstallerExecutorRemoveResp struct{}

func (s *InstallerExecutorRPC) Remove(ctx context.Context, pkgs []string, opts *manager.Opts) error {
	var resp *InstallerExecutorRemoveResp
	err := s.client.Call(ctx, "Plugin.Remove", &InstallerExecutorRemoveArgs{Pkgs: pkgs, Opts: opts}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *InstallerExecutorRPCServer) Remove(ctx context.Context, args *InstallerExecutorRemoveArgs, resp *InstallerExecutorRemoveResp) error {
	var err error
	err = s.Impl.Remove(ctx, args.Pkgs, args.Opts)
	if err != nil {
		return err
	}
	*resp = InstallerExecutorRemoveResp{}
	return nil
}

type InstallerExecutorRemoveAlreadyInstalledArgs struct {
	Pkgs []string
}
type InstallerExecutorRemoveAlreadyInstalledResp struct {
	Result0 []string
}

func (s *InstallerExecutorRPC) RemoveAlreadyInstalled(ctx context.Context, pkgs []string) ([]string, error) {
	var resp *InstallerExecutorRemoveAlreadyInstalledResp
	err := s.client.Call(ctx, "Plugin.RemoveAlreadyInstalled", &InstallerExecutorRemoveAlreadyInstalledArgs{Pkgs: pkgs}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result0, nil
}
func (s *InstallerExecutorRPCServer) RemoveAlreadyInstalled(ctx context.Context, args *InstallerExecutorRemoveAlreadyInstalledArgs, resp *InstallerExecutorRemoveAlreadyInstalledResp) error {
	var err error
	result0, err := s.Impl.RemoveAlreadyInstalled(ctx, args.Pkgs)
	if err != nil {
		return err
	}
	*resp = InstallerExecutorRemoveAlreadyInstalledResp{Result0: result0}
	return nil
}
