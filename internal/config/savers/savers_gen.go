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
