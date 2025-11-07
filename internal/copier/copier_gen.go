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

package copier

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/keegancsmith/rpc"
	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type CopierExecutorPlugin struct {
	Impl CopierExecutor
}
type CopierExecutorRPCServer struct {
	Impl   CopierExecutor
	broker *plugin.MuxBroker
}
type CopierExecutorRPC struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

func (p *CopierExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &CopierExecutorRPC{
		broker: b,
		client: c,
	}, nil
}
func (p *CopierExecutorPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &CopierExecutorRPCServer{
		Impl:   p.Impl,
		broker: b,
	}, nil
}

type CopierExecutorCopyArgs struct {
	F    *staplerfile.ScriptFile
	Info *distro.OSRelease
}
type CopierExecutorCopyResp struct {
	Result0 string
}

func (s *CopierExecutorRPC) Copy(ctx context.Context, f *staplerfile.ScriptFile, info *distro.OSRelease) (string, error) {
	var resp *CopierExecutorCopyResp
	err := s.client.Call(ctx, "Plugin.Copy", &CopierExecutorCopyArgs{F: f, Info: info}, &resp)
	if err != nil {
		return "", err
	}
	return resp.Result0, nil
}
func (s *CopierExecutorRPCServer) Copy(ctx context.Context, args *CopierExecutorCopyArgs, resp *CopierExecutorCopyResp) error {
	var err error
	result0, err := s.Impl.Copy(ctx, args.F, args.Info)
	if err != nil {
		return err
	}
	*resp = CopierExecutorCopyResp{Result0: result0}
	return nil
}

type CopierExecutorCopyOutArgs struct {
	Pkgs []commonbuild.BuiltDep
}
type CopierExecutorCopyOutResp struct{}

func (s *CopierExecutorRPC) CopyOut(ctx context.Context, pkgs []commonbuild.BuiltDep) error {
	var resp *CopierExecutorCopyOutResp
	err := s.client.Call(ctx, "Plugin.CopyOut", &CopierExecutorCopyOutArgs{Pkgs: pkgs}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *CopierExecutorRPCServer) CopyOut(ctx context.Context, args *CopierExecutorCopyOutArgs, resp *CopierExecutorCopyOutResp) error {
	var err error
	err = s.Impl.CopyOut(ctx, args.Pkgs)
	if err != nil {
		return err
	}
	*resp = CopierExecutorCopyOutResp{}
	return nil
}
