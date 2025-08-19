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
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/logger"
)

var pluginMap = map[string]plugin.Plugin{
	"script-executor": &ScriptExecutorPlugin{},
	"installer":       &InstallerExecutorPlugin{},
	"repos":           &ReposExecutorPlugin{},
	"script-copier":   &ScriptCopierPlugin{},
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "STPLR_PLUGIN",
	MagicCookieValue: "-",
}

func setCommonCmdEnv(cmd *exec.Cmd) {
	cmd.Env = []string{
		fmt.Sprintf("HOME=%s", constants.SystemCachePath),
		fmt.Sprintf("LOGNAME=%s", constants.BuilderUser),
		fmt.Sprintf("USER=%s", constants.BuilderUser),
		"PATH=/usr/bin:/bin:/usr/local/bin",
	}
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "LANG=") ||
			strings.HasPrefix(env, "LANGUAGE=") ||
			strings.HasPrefix(env, "LC_") ||
			strings.HasPrefix(env, "STPLR_LOG_LEVEL=") {
			cmd.Env = append(cmd.Env, env)
		}
	}
}

func GetPluginServeCommonConfig() *plugin.ServeConfig {
	return &plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:        "plugin",
			Output:      os.Stderr,
			Level:       hclog.Trace,
			JSONFormat:  true,
			DisableTime: true,
		}),
	}
}

func GetSafeInstaller() (InstallerExecutor, func(), error) {
	return getSafeExecutor[InstallerExecutor]("_internal-installer", "installer")
}

func GetSafeScriptExecutor() (ScriptExecutor, func(), error) {
	return getSafeExecutor[ScriptExecutor]("_internal-safe-script-executor", "script-executor")
}

func GetSafeReposExecutor() (ReposExecutor, func(), error) {
	return getSafeExecutor[ReposExecutor]("_internal-repos", "repos")
}

func GetSafeScriptCopier() (ScriptCopier, func(), error) {
	return getSafeExecutor[ScriptCopier]("_internal-script-copier", "script-copier")
}

func getSafeExecutor[T any](subCommand, pluginName string) (T, func(), error) {
	var err error

	executable, err := os.Executable()
	if err != nil {
		var zero T
		return zero, nil, err
	}

	cmd := exec.Command(executable, subCommand)
	setCommonCmdEnv(cmd)

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             cmd,
		Logger:          logger.GetHCLoggerAdapter(),
		SkipHostEnv:     true,
		UnixSocketConfig: &plugin.UnixSocketConfig{
			Group: constants.BuilderGroup,
		},
		SyncStderr: os.Stderr,
	})
	rpcClient, err := client.Client()
	if err != nil {
		var zero T
		return zero, nil, err
	}

	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			client.Kill()
		})
	}

	defer func() {
		if err != nil {
			slog.Debug("close executor")
			cleanup()
		}
	}()

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	executor, ok := raw.(T)
	if !ok {
		var zero T
		err = fmt.Errorf("dispensed object is not a %T (got %T)", zero, raw)
		return zero, nil, err
	}

	return executor, cleanup, nil
}
