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
	repoDirWriterKey      = "repo-dir-writer"
)

var builderPluginMap = map[string]plugin.Plugin{
	pullerPluginKey:       &repos.PullExecutorPlugin{},
	installerPluginKey:    &installer.InstallerExecutorPlugin{},
	copierPluginKey:       &copier.CopierExecutorPlugin{},
	scripterPluginKey:     &scripter.ScriptExecutorPlugin{},
	systemConfigWriterKey: &savers.SystemConfigWriterExecutorPlugin{},
	repoDirWriterKey:      &savers.RepoDirWriterExecutorPlugin{},
}

type pluginInitFunc func() (plugin.PluginSet, error)

func PluginMap(r repos.PullExecutor, s scripter.ScriptExecutor) pluginInitFunc {
	return func() (plugin.PluginSet, error) {
		return plugin.PluginSet{
			pullerPluginKey: &repos.PullExecutorPlugin{
				Impl: r,
			},
			scripterPluginKey: &scripter.ScriptExecutorPlugin{
				Impl: s,
			},
		}, nil
	}
}

func RootPluginMap(
	p installer.InstallerExecutor,
	c copier.CopierExecutor,
	s savers.SystemConfigWriterExecutor,
	r savers.RepoDirWriterExecutor,
) pluginInitFunc {
	return func() (plugin.PluginSet, error) {
		return plugin.PluginSet{
			installerPluginKey: &installer.InstallerExecutorPlugin{
				Impl: p,
			},
			copierPluginKey: &copier.CopierExecutorPlugin{
				Impl: c,
			},
			systemConfigWriterKey: &savers.SystemConfigWriterExecutorPlugin{
				Impl: s,
			},
			repoDirWriterKey: &savers.RepoDirWriterExecutorPlugin{
				Impl: r,
			},
		}, nil
	}
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "STPLR_PLUGIN",
	MagicCookieValue: "stplr-plugin-handshake-v1-adef82d6e4622d8cdba1098291845fe9", // echo "stplr-plugin-handshake-v1-$(openssl rand -hex 16)"
}

// ensureSocketBaseDir creates the shared base directory.
//
// The base must remain world-writable + sticky because different UIDs (regular
// user, builder, root) may each host their own per-invocation subdirs in here.
func ensureSocketBaseDir() error {
	if err := os.MkdirAll(constants.SocketDirPath, 0o777); err != nil {
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

// prepareSocketDir creates a per-invocation subdirectory under the shared base.
func prepareSocketDir(privileged bool) (string, error) {
	if err := ensureSocketBaseDir(); err != nil {
		return "", err
	}
	dir, err := os.MkdirTemp(constants.SocketDirPath, "sock-")
	if err != nil {
		return "", err
	}
	if privileged {
		_, gid, err := utils.GetUidGidStaplerUser()
		if err != nil {
			_ = os.RemoveAll(dir)
			return "", fmt.Errorf("lookup %s group: %w", constants.BuilderGroup, err)
		}
		if err := os.Chown(dir, -1, gid); err != nil {
			_ = os.RemoveAll(dir)
			return "", err
		}
		if err := os.Chmod(dir, 0o770); err != nil {
			_ = os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}

type Provider struct {
	output output.Output

	c         *plugin.Client
	client    plugin.ClientProtocol
	socketDir string
}

func (p *Provider) Cleanup() error {
	if p.c != nil {
		p.c.Kill()
	}
	if p.socketDir != "" {
		_ = os.RemoveAll(p.socketDir)
		p.socketDir = ""
	}
	return nil
}

func (p *Provider) setupConnection(subcommand string) error {
	var err error

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	isBuildUser, err := utils.IsBuilderUser()
	if err != nil {
		return err
	}
	privileged := utils.IsRoot() || isBuildUser

	socketDir, err := prepareSocketDir(privileged)
	if err != nil {
		return fmt.Errorf("failed to prepare socket dir: %w", err)
	}
	p.socketDir = socketDir

	unixSocketConfig := &plugin.UnixSocketConfig{}

	args := []string{subcommand}
	cmd := exec.Command(executable, args...)

	if privileged {
		// Socket file itself is created by the child; group must be set so
		// the parent (which may have demoted from root to builder by now) can
		// connect to a socket created by a root-uid child.
		unixSocketConfig.Group = constants.BuilderGroup
		setCommonCmdEnv(cmd)
	} else {
		passCommonEnv(cmd)
	}

	// HACK: tell go-plugin where to put the unix socket.
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", plugin.EnvUnixSocketDir, socketDir))

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

func commonServe(cfg *plugin.ServeConfig) error {
	logger.SetupForGoPlugin()
	logger := hclog.New(&hclog.LoggerOptions{
		Name:        "plugin",
		Output:      os.Stderr,
		Level:       hclog.Debug,
		JSONFormat:  true,
		DisableTime: true,
	})
	cfg.HandshakeConfig = HandshakeConfig
	cfg.Logger = logger
	plugin.Serve(cfg)
	return nil
}

func Serve(initFunc pluginInitFunc) error {
	return commonServe(&plugin.ServeConfig{
		InitFunc: initFunc,
	})
}

func ServeError(err error) error {
	return commonServe(&plugin.ServeConfig{
		InitFunc: func() (plugin.PluginSet, error) {
			return nil, err
		},
	})
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

func GetRepoDirWriter(ctx context.Context, provider *Provider) (savers.RepoDirWriterExecutor, error) {
	return Get[savers.RepoDirWriterExecutor](ctx, provider, repoDirWriterKey)
}
