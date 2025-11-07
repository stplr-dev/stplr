// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
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

package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/internal/config/savers"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/copier"
	"go.stplr.dev/stplr/internal/installer"
	"go.stplr.dev/stplr/internal/logger"
	"go.stplr.dev/stplr/internal/scripter"
	"go.stplr.dev/stplr/internal/service/repos"
	"go.stplr.dev/stplr/internal/utils"
)

const (
	ProviderSubcommand     = "_plugin_provider"
	RootProviderSubcommand = "_root_plugin_provider"
)

const (
	pullerPluginKey       = "puller"
	installerPluginKey    = "installer"
	copierPluginKey       = "copier"
	scripterPluginKey     = "scripter"
	systemConfigWriterKey = "system-config-writer"
)

var builderPluginMap = map[string]plugin.Plugin{
	pullerPluginKey:       &repos.PullExecutorPlugin{},
	installerPluginKey:    &installer.InstallerExecutorPlugin{},
	copierPluginKey:       &copier.CopierExecutorPlugin{},
	scripterPluginKey:     &scripter.ScriptExecutorPlugin{},
	systemConfigWriterKey: &savers.SystemConfigWriterExecutorPlugin{},
}

func PluginMap(r repos.PullExecutor, s scripter.ScriptExecutor) map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		pullerPluginKey: &repos.PullExecutorPlugin{
			Impl: r,
		},
		scripterPluginKey: &scripter.ScriptExecutorPlugin{
			Impl: s,
		},
	}
}

func RootPluginMap(
	p installer.InstallerExecutor,
	c copier.CopierExecutor,
	s savers.SystemConfigWriterExecutor,
) map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		installerPluginKey: &installer.InstallerExecutorPlugin{
			Impl: p,
		},
		copierPluginKey: &copier.CopierExecutorPlugin{
			Impl: c,
		},
		systemConfigWriterKey: &savers.SystemConfigWriterExecutorPlugin{
			Impl: s,
		},
	}
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "STPLR_PLUGIN",
	MagicCookieValue: "-",
}

func prepareSocketDirPath() error {
	err := os.MkdirAll(constants.SocketDirPath, 0o777)
	if err != nil {
		return err
	}

	info, err := os.Stat(constants.SocketDirPath)
	if err != nil {
		return err
	}

	requiredMode := os.FileMode(0o777 | os.ModeSticky)

	if info.Mode().Perm() != requiredMode.Perm() || info.Mode()&os.ModeSticky == 0 {
		if err := os.Chmod(constants.SocketDirPath, requiredMode); err != nil {
			return err
		}
	}

	return nil
}

type Provider struct {
	output output.Output

	c      *plugin.Client
	client plugin.ClientProtocol
}

func (p *Provider) Cleanup() error {
	if p.c != nil {
		p.c.Kill()
	}
	return nil
}

func (p *Provider) setupConnection(subcommand string) error {
	var err error

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	if err := prepareSocketDirPath(); err != nil {
		return fmt.Errorf("failed to prepare socket dir: %w", err)
	}

	unixSocketConfig := &plugin.UnixSocketConfig{}

	args := []string{subcommand}
	cmd := exec.Command(executable, args...)

	isBuildUser, err := utils.IsBuilderUser()
	if err != nil {
		return err
	}

	if utils.IsRoot() || isBuildUser {
		unixSocketConfig.Group = constants.BuilderGroup
		setCommonCmdEnv(cmd)
	} else {
		passCommonEnv(cmd)
	}

	// HACK
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", plugin.EnvUnixSocketDir, constants.SocketDirPath))

	p.c = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  HandshakeConfig,
		Plugins:          builderPluginMap,
		Cmd:              cmd,
		Logger:           logger.GetHCLoggerAdapter(p.output),
		SkipHostEnv:      true,
		UnixSocketConfig: unixSocketConfig,
		SyncStderr:       os.Stderr,
	})
	p.client, err = p.c.Client()
	if err != nil {
		return err
	}

	return nil
}

func NewProvider(out output.Output) *Provider {
	return &Provider{
		output: out,
	}
}

func (p *Provider) SetupConnection() error {
	return p.setupConnection(ProviderSubcommand)
}

func (p *Provider) SetupRootConnection() error {
	return p.setupConnection(RootProviderSubcommand)
}

func Serve(plugins map[string]plugin.Plugin) error {
	logger.SetupForGoPlugin()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:        "plugin",
		Output:      os.Stderr,
		Level:       hclog.Debug,
		JSONFormat:  true,
		DisableTime: true,
	})
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         plugins,
		Logger:          logger,
	})
	return nil
}

func Get[T any](ctx context.Context, provider *Provider, name string) (T, error) {
	var zero T
	raw, err := provider.client.Dispense(ctx, name)
	if err != nil {
		return zero, err
	}
	v, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("dispensed object is not a %T (got %T)", zero, raw)
	}
	return v, nil
}

func GetPuller(ctx context.Context, provider *Provider) (repos.PullExecutor, error) {
	return Get[repos.PullExecutor](ctx, provider, pullerPluginKey)
}

func GetInstaller(ctx context.Context, provider *Provider) (installer.InstallerExecutor, error) {
	return Get[installer.InstallerExecutor](ctx, provider, installerPluginKey)
}

func GetCopier(ctx context.Context, provider *Provider) (copier.CopierExecutor, error) {
	return Get[copier.CopierExecutor](ctx, provider, copierPluginKey)
}

func GetScripter(ctx context.Context, provider *Provider) (scripter.ScriptExecutor, error) {
	return Get[scripter.ScriptExecutor](ctx, provider, scripterPluginKey)
}

func GetSystemConfigWriter(ctx context.Context, provider *Provider) (savers.SystemConfigWriterExecutor, error) {
	return Get[savers.SystemConfigWriterExecutor](ctx, provider, systemConfigWriterKey)
}
