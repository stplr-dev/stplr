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

package savers

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/keegancsmith/rpc"
)

type SystemConfigWriterExecutorPlugin struct {
	Impl SystemConfigWriterExecutor
}
type SystemConfigWriterExecutorRPCServer struct {
	Impl   SystemConfigWriterExecutor
	broker *plugin.MuxBroker
}
type SystemConfigWriterExecutorRPC struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

func (p *SystemConfigWriterExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &SystemConfigWriterExecutorRPC{
		broker: b,
		client: c,
	}, nil
}
func (p *SystemConfigWriterExecutorPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &SystemConfigWriterExecutorRPCServer{
		Impl:   p.Impl,
		broker: b,
	}, nil
}

type RepoDirWriterExecutorPlugin struct {
	Impl RepoDirWriterExecutor
}
type RepoDirWriterExecutorRPCServer struct {
	Impl   RepoDirWriterExecutor
	broker *plugin.MuxBroker
}
type RepoDirWriterExecutorRPC struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

func (p *RepoDirWriterExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RepoDirWriterExecutorRPC{
		broker: b,
		client: c,
	}, nil
}
func (p *RepoDirWriterExecutorPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &RepoDirWriterExecutorRPCServer{
		Impl:   p.Impl,
		broker: b,
	}, nil
}

type SystemConfigWriterExecutorWriteArgs struct {
	P []byte
}
type SystemConfigWriterExecutorWriteResp struct {
	N int
}

func (s *SystemConfigWriterExecutorRPC) Write(p []byte) (int, error) {
	var resp *SystemConfigWriterExecutorWriteResp
	ctx := context.Background()
	err := s.client.Call(ctx, "Plugin.Write", &SystemConfigWriterExecutorWriteArgs{P: p}, &resp)
	if err != nil {
		return 0, err
	}
	return resp.N, nil
}
func (s *SystemConfigWriterExecutorRPCServer) Write(ctx context.Context, args *SystemConfigWriterExecutorWriteArgs, resp *SystemConfigWriterExecutorWriteResp) error {
	var err error
	n, err := s.Impl.Write(args.P)
	if err != nil {
		return err
	}
	*resp = SystemConfigWriterExecutorWriteResp{N: n}
	return nil
}

type RepoDirWriterExecutorRemoveOverrideArgs struct {
	Name string
}
type RepoDirWriterExecutorRemoveOverrideResp struct{}

func (s *RepoDirWriterExecutorRPC) RemoveOverride(ctx context.Context, name string) error {
	var resp *RepoDirWriterExecutorRemoveOverrideResp
	err := s.client.Call(ctx, "Plugin.RemoveOverride", &RepoDirWriterExecutorRemoveOverrideArgs{Name: name}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *RepoDirWriterExecutorRPCServer) RemoveOverride(ctx context.Context, args *RepoDirWriterExecutorRemoveOverrideArgs, resp *RepoDirWriterExecutorRemoveOverrideResp) error {
	var err error
	err = s.Impl.RemoveOverride(ctx, args.Name)
	if err != nil {
		return err
	}
	*resp = RepoDirWriterExecutorRemoveOverrideResp{}
	return nil
}

type RepoDirWriterExecutorRemoveUserRepoArgs struct {
	Name string
}
type RepoDirWriterExecutorRemoveUserRepoResp struct{}

func (s *RepoDirWriterExecutorRPC) RemoveUserRepo(ctx context.Context, name string) error {
	var resp *RepoDirWriterExecutorRemoveUserRepoResp
	err := s.client.Call(ctx, "Plugin.RemoveUserRepo", &RepoDirWriterExecutorRemoveUserRepoArgs{Name: name}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *RepoDirWriterExecutorRPCServer) RemoveUserRepo(ctx context.Context, args *RepoDirWriterExecutorRemoveUserRepoArgs, resp *RepoDirWriterExecutorRemoveUserRepoResp) error {
	var err error
	err = s.Impl.RemoveUserRepo(ctx, args.Name)
	if err != nil {
		return err
	}
	*resp = RepoDirWriterExecutorRemoveUserRepoResp{}
	return nil
}

type RepoDirWriterExecutorWriteOverrideArgs struct {
	Name string
	Data []byte
}
type RepoDirWriterExecutorWriteOverrideResp struct{}

func (s *RepoDirWriterExecutorRPC) WriteOverride(ctx context.Context, name string, data []byte) error {
	var resp *RepoDirWriterExecutorWriteOverrideResp
	err := s.client.Call(ctx, "Plugin.WriteOverride", &RepoDirWriterExecutorWriteOverrideArgs{Name: name, Data: data}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *RepoDirWriterExecutorRPCServer) WriteOverride(ctx context.Context, args *RepoDirWriterExecutorWriteOverrideArgs, resp *RepoDirWriterExecutorWriteOverrideResp) error {
	var err error
	err = s.Impl.WriteOverride(ctx, args.Name, args.Data)
	if err != nil {
		return err
	}
	*resp = RepoDirWriterExecutorWriteOverrideResp{}
	return nil
}

type RepoDirWriterExecutorWriteUserRepoArgs struct {
	Name string
	Data []byte
}
type RepoDirWriterExecutorWriteUserRepoResp struct{}

func (s *RepoDirWriterExecutorRPC) WriteUserRepo(ctx context.Context, name string, data []byte) error {
	var resp *RepoDirWriterExecutorWriteUserRepoResp
	err := s.client.Call(ctx, "Plugin.WriteUserRepo", &RepoDirWriterExecutorWriteUserRepoArgs{Name: name, Data: data}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *RepoDirWriterExecutorRPCServer) WriteUserRepo(ctx context.Context, args *RepoDirWriterExecutorWriteUserRepoArgs, resp *RepoDirWriterExecutorWriteUserRepoResp) error {
	var err error
	err = s.Impl.WriteUserRepo(ctx, args.Name, args.Data)
	if err != nil {
		return err
	}
	*resp = RepoDirWriterExecutorWriteUserRepoResp{}
	return nil
}
