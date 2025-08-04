// DO NOT EDIT MANUALLY. This file is generated.

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
	"net/rpc"

	"context"
	"github.com/hashicorp/go-plugin"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

type InstallerExecutorPlugin struct {
	Impl InstallerExecutor
}

type InstallerExecutorRPCServer struct {
	Impl InstallerExecutor
}

type InstallerExecutorRPC struct {
	client *rpc.Client
}

func (p *InstallerExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &InstallerExecutorRPC{client: c}, nil
}

func (p *InstallerExecutorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &InstallerExecutorRPCServer{Impl: p.Impl}, nil
}

type ScriptExecutorPlugin struct {
	Impl ScriptExecutor
}

type ScriptExecutorRPCServer struct {
	Impl ScriptExecutor
}

type ScriptExecutorRPC struct {
	client *rpc.Client
}

func (p *ScriptExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ScriptExecutorRPC{client: c}, nil
}

func (p *ScriptExecutorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ScriptExecutorRPCServer{Impl: p.Impl}, nil
}

type ReposExecutorPlugin struct {
	Impl ReposExecutor
}

type ReposExecutorRPCServer struct {
	Impl ReposExecutor
}

type ReposExecutorRPC struct {
	client *rpc.Client
}

func (p *ReposExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ReposExecutorRPC{client: c}, nil
}

func (p *ReposExecutorPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ReposExecutorRPCServer{Impl: p.Impl}, nil
}

type ScriptReaderPlugin struct {
	Impl ScriptReader
}

type ScriptReaderRPCServer struct {
	Impl ScriptReader
}

type ScriptReaderRPC struct {
	client *rpc.Client
}

func (p *ScriptReaderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ScriptReaderRPC{client: c}, nil
}

func (p *ScriptReaderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ScriptReaderRPCServer{Impl: p.Impl}, nil
}

type PackagesParserPlugin struct {
	Impl PackagesParser
}

type PackagesParserRPCServer struct {
	Impl PackagesParser
}

type PackagesParserRPC struct {
	client *rpc.Client
}

func (p *PackagesParserPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PackagesParserRPC{client: c}, nil
}

func (p *PackagesParserPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PackagesParserRPCServer{Impl: p.Impl}, nil
}

type InstallerExecutorInstallLocalArgs struct {
	Paths []string
	Opts  *manager.Opts
}

type InstallerExecutorInstallLocalResp struct {
}

func (s *InstallerExecutorRPC) InstallLocal(ctx context.Context, paths []string, opts *manager.Opts) error {
	var resp *InstallerExecutorInstallLocalResp
	err := s.client.Call("Plugin.InstallLocal", &InstallerExecutorInstallLocalArgs{
		Paths: paths,
		Opts:  opts,
	}, &resp)
	if err != nil {
		return err
	}
	return nil
}

func (s *InstallerExecutorRPCServer) InstallLocal(args *InstallerExecutorInstallLocalArgs, resp *InstallerExecutorInstallLocalResp) error {
	err := s.Impl.InstallLocal(context.Background(), args.Paths, args.Opts)
	if err != nil {
		return err
	}
	*resp = InstallerExecutorInstallLocalResp{}
	return nil
}

type InstallerExecutorInstallArgs struct {
	Pkgs []string
	Opts *manager.Opts
}

type InstallerExecutorInstallResp struct {
}

func (s *InstallerExecutorRPC) Install(ctx context.Context, pkgs []string, opts *manager.Opts) error {
	var resp *InstallerExecutorInstallResp
	err := s.client.Call("Plugin.Install", &InstallerExecutorInstallArgs{
		Pkgs: pkgs,
		Opts: opts,
	}, &resp)
	if err != nil {
		return err
	}
	return nil
}

func (s *InstallerExecutorRPCServer) Install(args *InstallerExecutorInstallArgs, resp *InstallerExecutorInstallResp) error {
	err := s.Impl.Install(context.Background(), args.Pkgs, args.Opts)
	if err != nil {
		return err
	}
	*resp = InstallerExecutorInstallResp{}
	return nil
}

type InstallerExecutorRemoveArgs struct {
	Pkgs []string
	Opts *manager.Opts
}

type InstallerExecutorRemoveResp struct {
}

func (s *InstallerExecutorRPC) Remove(ctx context.Context, pkgs []string, opts *manager.Opts) error {
	var resp *InstallerExecutorRemoveResp
	err := s.client.Call("Plugin.Remove", &InstallerExecutorRemoveArgs{
		Pkgs: pkgs,
		Opts: opts,
	}, &resp)
	if err != nil {
		return err
	}
	return nil
}

func (s *InstallerExecutorRPCServer) Remove(args *InstallerExecutorRemoveArgs, resp *InstallerExecutorRemoveResp) error {
	err := s.Impl.Remove(context.Background(), args.Pkgs, args.Opts)
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
	err := s.client.Call("Plugin.RemoveAlreadyInstalled", &InstallerExecutorRemoveAlreadyInstalledArgs{
		Pkgs: pkgs,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result0, nil
}

func (s *InstallerExecutorRPCServer) RemoveAlreadyInstalled(args *InstallerExecutorRemoveAlreadyInstalledArgs, resp *InstallerExecutorRemoveAlreadyInstalledResp) error {
	result0, err := s.Impl.RemoveAlreadyInstalled(context.Background(), args.Pkgs)
	if err != nil {
		return err
	}
	*resp = InstallerExecutorRemoveAlreadyInstalledResp{
		Result0: result0,
	}
	return nil
}

type ScriptExecutorReadArgs struct {
	ScriptPath string
}

type ScriptExecutorReadResp struct {
	Result0 *staplerfile.ScriptFile
}

func (s *ScriptExecutorRPC) Read(ctx context.Context, scriptPath string) (*staplerfile.ScriptFile, error) {
	var resp *ScriptExecutorReadResp
	err := s.client.Call("Plugin.Read", &ScriptExecutorReadArgs{
		ScriptPath: scriptPath,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result0, nil
}

func (s *ScriptExecutorRPCServer) Read(args *ScriptExecutorReadArgs, resp *ScriptExecutorReadResp) error {
	result0, err := s.Impl.Read(context.Background(), args.ScriptPath)
	if err != nil {
		return err
	}
	*resp = ScriptExecutorReadResp{
		Result0: result0,
	}
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
	err := s.client.Call("Plugin.ParsePackages", &ScriptExecutorParsePackagesArgs{
		File:     file,
		Packages: packages,
		Info:     info,
	}, &resp)
	if err != nil {
		return "", nil, err
	}
	return resp.Result0, resp.Result1, nil
}

func (s *ScriptExecutorRPCServer) ParsePackages(args *ScriptExecutorParsePackagesArgs, resp *ScriptExecutorParsePackagesResp) error {
	result0, result1, err := s.Impl.ParsePackages(context.Background(), args.File, args.Packages, args.Info)
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
	Input   *BuildInput
	BasePkg string
}

type ScriptExecutorPrepareDirsResp struct {
}

func (s *ScriptExecutorRPC) PrepareDirs(ctx context.Context, input *BuildInput, basePkg string) error {
	var resp *ScriptExecutorPrepareDirsResp
	err := s.client.Call("Plugin.PrepareDirs", &ScriptExecutorPrepareDirsArgs{
		Input:   input,
		BasePkg: basePkg,
	}, &resp)
	if err != nil {
		return err
	}
	return nil
}

func (s *ScriptExecutorRPCServer) PrepareDirs(args *ScriptExecutorPrepareDirsArgs, resp *ScriptExecutorPrepareDirsResp) error {
	err := s.Impl.PrepareDirs(context.Background(), args.Input, args.BasePkg)
	if err != nil {
		return err
	}
	*resp = ScriptExecutorPrepareDirsResp{}
	return nil
}

type ScriptExecutorExecuteSecondPassArgs struct {
	Input          *BuildInput
	Sf             *staplerfile.ScriptFile
	VarsOfPackages []*staplerfile.Package
	RepoDeps       []string
	BuiltDeps      []*BuiltDep
	BasePkg        string
}

type ScriptExecutorExecuteSecondPassResp struct {
	Result0 []*BuiltDep
}

func (s *ScriptExecutorRPC) ExecuteSecondPass(ctx context.Context, input *BuildInput, sf *staplerfile.ScriptFile, varsOfPackages []*staplerfile.Package, repoDeps []string, builtDeps []*BuiltDep, basePkg string) ([]*BuiltDep, error) {
	var resp *ScriptExecutorExecuteSecondPassResp
	err := s.client.Call("Plugin.ExecuteSecondPass", &ScriptExecutorExecuteSecondPassArgs{
		Input:          input,
		Sf:             sf,
		VarsOfPackages: varsOfPackages,
		RepoDeps:       repoDeps,
		BuiltDeps:      builtDeps,
		BasePkg:        basePkg,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result0, nil
}

func (s *ScriptExecutorRPCServer) ExecuteSecondPass(args *ScriptExecutorExecuteSecondPassArgs, resp *ScriptExecutorExecuteSecondPassResp) error {
	result0, err := s.Impl.ExecuteSecondPass(context.Background(), args.Input, args.Sf, args.VarsOfPackages, args.RepoDeps, args.BuiltDeps, args.BasePkg)
	if err != nil {
		return err
	}
	*resp = ScriptExecutorExecuteSecondPassResp{
		Result0: result0,
	}
	return nil
}

type ReposExecutorPullOneAndUpdateFromConfigArgs struct {
	Repo *types.Repo
}

type ReposExecutorPullOneAndUpdateFromConfigResp struct {
	Result0 types.Repo
}

func (s *ReposExecutorRPC) PullOneAndUpdateFromConfig(ctx context.Context, repo *types.Repo) (types.Repo, error) {
	var resp *ReposExecutorPullOneAndUpdateFromConfigResp
	err := s.client.Call("Plugin.PullOneAndUpdateFromConfig", &ReposExecutorPullOneAndUpdateFromConfigArgs{
		Repo: repo,
	}, &resp)
	if err != nil {
		return types.Repo{}, err
	}
	return resp.Result0, nil
}

func (s *ReposExecutorRPCServer) PullOneAndUpdateFromConfig(args *ReposExecutorPullOneAndUpdateFromConfigArgs, resp *ReposExecutorPullOneAndUpdateFromConfigResp) error {
	result0, err := s.Impl.PullOneAndUpdateFromConfig(context.Background(), args.Repo)
	if err != nil {
		return err
	}
	*resp = ReposExecutorPullOneAndUpdateFromConfigResp{
		Result0: result0,
	}
	return nil
}

type ScriptReaderReadArgs struct {
	Path string
}

type ScriptReaderReadResp struct {
	Result0 *staplerfile.ScriptFile
}

func (s *ScriptReaderRPC) Read(ctx context.Context, path string) (*staplerfile.ScriptFile, error) {
	var resp *ScriptReaderReadResp
	err := s.client.Call("Plugin.Read", &ScriptReaderReadArgs{
		Path: path,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result0, nil
}

func (s *ScriptReaderRPCServer) Read(args *ScriptReaderReadArgs, resp *ScriptReaderReadResp) error {
	result0, err := s.Impl.Read(context.Background(), args.Path)
	if err != nil {
		return err
	}
	*resp = ScriptReaderReadResp{
		Result0: result0,
	}
	return nil
}

type PackagesParserParsePackagesArgs struct {
	File     *staplerfile.ScriptFile
	Packages []string
	Info     distro.OSRelease
}

type PackagesParserParsePackagesResp struct {
	Result0 string
	Result1 []*staplerfile.Package
}

func (s *PackagesParserRPC) ParsePackages(ctx context.Context, file *staplerfile.ScriptFile, packages []string, info distro.OSRelease) (string, []*staplerfile.Package, error) {
	var resp *PackagesParserParsePackagesResp
	err := s.client.Call("Plugin.ParsePackages", &PackagesParserParsePackagesArgs{
		File:     file,
		Packages: packages,
		Info:     info,
	}, &resp)
	if err != nil {
		return "", nil, err
	}
	return resp.Result0, resp.Result1, nil
}

func (s *PackagesParserRPCServer) ParsePackages(args *PackagesParserParsePackagesArgs, resp *PackagesParserParsePackagesResp) error {
	result0, result1, err := s.Impl.ParsePackages(context.Background(), args.File, args.Packages, args.Info)
	if err != nil {
		return err
	}
	*resp = PackagesParserParsePackagesResp{
		Result0: result0,
		Result1: result1,
	}
	return nil
}
