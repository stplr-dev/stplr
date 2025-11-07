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

package puller

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/keegancsmith/rpc"
	"go.stplr.dev/stplr/internal/plugins/shared"
	"go.stplr.dev/stplr/pkg/types"
)

type PullExecutorPlugin struct {
	Impl PullExecutor
}
type PullExecutorRPCServer struct {
	Impl   PullExecutor
	broker *plugin.MuxBroker
}
type PullExecutorRPC struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

func (p *PullExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PullExecutorRPC{
		broker: b,
		client: c,
	}, nil
}
func (p *PullExecutorPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &PullExecutorRPCServer{
		Impl:   p.Impl,
		broker: b,
	}, nil
}

type PullReporterPlugin struct {
	Impl PullReporter
}
type PullReporterRPCServer struct {
	Impl   PullReporter
	broker *plugin.MuxBroker
}
type PullReporterRPC struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

func (p *PullReporterPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PullReporterRPC{
		broker: b,
		client: c,
	}, nil
}
func (p *PullReporterPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &PullReporterRPCServer{
		Impl:   p.Impl,
		broker: b,
	}, nil
}

type PullExecutorPullArgs struct {
	Repo   types.Repo
	Report uint32
}
type PullExecutorPullResp struct {
	Result0 types.Repo
}

func (s *PullExecutorRPC) Pull(ctx context.Context, repo types.Repo, report PullReporter) (types.Repo, error) {
	var resp *PullExecutorPullResp
	serverReport := &PullReporterRPCServer{report, s.broker}
	brokerIdReport := s.broker.NextId()
	go s.broker.AcceptAndServe(brokerIdReport, serverReport)
	err := s.client.Call(ctx, "Plugin.Pull", &PullExecutorPullArgs{Repo: repo, Report: brokerIdReport}, &resp)
	if err != nil {
		return types.Repo{}, err
	}
	return resp.Result0, nil
}
func (s *PullExecutorRPCServer) Pull(ctx context.Context, args *PullExecutorPullArgs, resp *PullExecutorPullResp) error {
	var err error
	connReport, err := s.broker.Dial(args.Report)
	if err != nil {
		return err
	}
	clientReport := rpc.NewClient(connReport)
	rpcReport := &PullReporterRPC{clientReport, s.broker}
	result0, err := s.Impl.Pull(ctx, args.Repo, rpcReport)
	if err != nil {
		return err
	}
	*resp = PullExecutorPullResp{Result0: result0}
	return nil
}

type PullReporterNotifyArgs struct {
	Event shared.NotifyEvent
	Data  map[string]string
}
type PullReporterNotifyResp struct{}

func (s *PullReporterRPC) Notify(ctx context.Context, event shared.NotifyEvent, data map[string]string) error {
	var resp *PullReporterNotifyResp
	err := s.client.Call(ctx, "Plugin.Notify", &PullReporterNotifyArgs{Event: event, Data: data}, &resp)
	if err != nil {
		return err
	}
	return nil
}
func (s *PullReporterRPCServer) Notify(ctx context.Context, args *PullReporterNotifyArgs, resp *PullReporterNotifyResp) error {
	var err error
	err = s.Impl.Notify(ctx, args.Event, args.Data)
	if err != nil {
		return err
	}
	*resp = PullReporterNotifyResp{}
	return nil
}

type PullReporterNotifyWriteArgs struct {
	Event shared.NotifyWriterEvent
	P     []byte
}
type PullReporterNotifyWriteResp struct {
	N int
}

func (s *PullReporterRPC) NotifyWrite(ctx context.Context, event shared.NotifyWriterEvent, p []byte) (int, error) {
	var resp *PullReporterNotifyWriteResp
	err := s.client.Call(ctx, "Plugin.NotifyWrite", &PullReporterNotifyWriteArgs{Event: event, P: p}, &resp)
	if err != nil {
		return 0, err
	}
	return resp.N, nil
}
func (s *PullReporterRPCServer) NotifyWrite(ctx context.Context, args *PullReporterNotifyWriteArgs, resp *PullReporterNotifyWriteResp) error {
	var err error
	n, err := s.Impl.NotifyWrite(ctx, args.Event, args.P)
	if err != nil {
		return err
	}
	*resp = PullReporterNotifyWriteResp{N: n}
	return nil
}
