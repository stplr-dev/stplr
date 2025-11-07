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

package scripter

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/keegancsmith/rpc"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type ScriptExecutorPlugin struct {
	Impl ScriptExecutor
}
type ScriptExecutorRPCServer struct {
	Impl   ScriptExecutor
	broker *plugin.MuxBroker
}
type ScriptExecutorRPC struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

func (p *ScriptExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ScriptExecutorRPC{
		broker: b,
		client: c,
	}, nil
}
func (p *ScriptExecutorPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &ScriptExecutorRPCServer{
		Impl:   p.Impl,
		broker: b,
	}, nil
}

type ScriptExecutorExecuteSecondPassArgs struct {
	Input          *commonbuild.BuildInput
	Sf             *staplerfile.ScriptFile
	VarsOfPackages []*staplerfile.Package
	RepoDeps       []string
	BuiltDeps      []*commonbuild.BuiltDep
	BasePkg        string
}
type ScriptExecutorExecuteSecondPassResp struct {
	Result0 []*commonbuild.BuiltDep
}

func (s *ScriptExecutorRPC) ExecuteSecondPass(ctx context.Context, input *commonbuild.BuildInput, sf *staplerfile.ScriptFile, varsOfPackages []*staplerfile.Package, repoDeps []string, builtDeps []*commonbuild.BuiltDep, basePkg string) ([]*commonbuild.BuiltDep, error) {
	var resp *ScriptExecutorExecuteSecondPassResp
	err := s.client.Call(ctx, "Plugin.ExecuteSecondPass", &ScriptExecutorExecuteSecondPassArgs{Input: input, Sf: sf, VarsOfPackages: varsOfPackages, RepoDeps: repoDeps, BuiltDeps: builtDeps, BasePkg: basePkg}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result0, nil
}
func (s *ScriptExecutorRPCServer) ExecuteSecondPass(ctx context.Context, args *ScriptExecutorExecuteSecondPassArgs, resp *ScriptExecutorExecuteSecondPassResp) error {
	var err error
	result0, err := s.Impl.ExecuteSecondPass(ctx, args.Input, args.Sf, args.VarsOfPackages, args.RepoDeps, args.BuiltDeps, args.BasePkg)
	if err != nil {
		return err
	}
	*resp = ScriptExecutorExecuteSecondPassResp{Result0: result0}
	return nil
}

type ScriptExecutorParsePackagesArgs struct {
	File     *staplerfile.ScriptFile
	Packages []string
	Info     distro.OSRelease
}
type ScriptExecutorParsePackagesResp struct {
	Result0 string
	Result1 []*staplerfile.Package
}

func (s *ScriptExecutorRPC) ParsePackages(ctx context.Context, file *staplerfile.ScriptFile, packages []string, info distro.OSRelease) (string, []*staplerfile.Package, error) {
	var resp *ScriptExecutorParsePackagesResp
	err := s.client.Call(ctx, "Plugin.ParsePackages", &ScriptExecutorParsePackagesArgs{File: file, Packages: packages, Info: info}, &resp)
	if err != nil {
		return "", nil, err
	}
	return resp.Result0, resp.Result1, nil
}
func (s *ScriptExecutorRPCServer) ParsePackages(ctx context.Context, args *ScriptExecutorParsePackagesArgs, resp *ScriptExecutorParsePackagesResp) error {
	var err error
	result0, result1, err := s.Impl.ParsePackages(ctx, args.File, args.Packages, args.Info)
	if err != nil {
		return err
	}
	*resp = ScriptExecutorParsePackagesResp{
		Result0: result0,
		Result1: result1,
	}
	return nil
}

type ScriptExecutorPrepareDirsArgs struct {
	Input   *commonbuild.BuildInput
	BasePkg string
}
type ScriptExecutorPrepareDirsResp struct{}

func (s *ScriptExecutorRPC) PrepareDirs(ctx context.Context, input *commonbuild.BuildInput, basePkg string) error {
	var resp *ScriptExecutorPrepareDirsResp
	err := s.client.Call(ctx, "Plugin.PrepareDirs", &ScriptExecutorPrepareDirsArgs{Input: input, BasePkg: basePkg}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *ScriptExecutorRPCServer) PrepareDirs(ctx context.Context, args *ScriptExecutorPrepareDirsArgs, resp *ScriptExecutorPrepareDirsResp) error {
	var err error
	err = s.Impl.PrepareDirs(ctx, args.Input, args.BasePkg)
	if err != nil {
		return err
	}
	*resp = ScriptExecutorPrepareDirsResp{}
	return nil
}

type ScriptExecutorReadArgs struct {
	Path string
}
type ScriptExecutorReadResp struct {
	Result0 *staplerfile.ScriptFile
}

func (s *ScriptExecutorRPC) Read(ctx context.Context, path string) (*staplerfile.ScriptFile, error) {
	var resp *ScriptExecutorReadResp
	err := s.client.Call(ctx, "Plugin.Read", &ScriptExecutorReadArgs{Path: path}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result0, nil
}
func (s *ScriptExecutorRPCServer) Read(ctx context.Context, args *ScriptExecutorReadArgs, resp *ScriptExecutorReadResp) error {
	var err error
	result0, err := s.Impl.Read(ctx, args.Path)
	if err != nil {
		return err
	}
	*resp = ScriptExecutorReadResp{Result0: result0}
	return nil
}
