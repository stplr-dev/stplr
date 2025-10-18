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

package rpm

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/goreleaser/nfpm/v2"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/internal/app/output"
	"go.stplr.dev/stplr/pkg/types"
)

type ALTLinux struct{}

func NewALTLinux() *ALTLinux {
	return &ALTLinux{}
}

func (o *ALTLinux) FindProvides(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	return o.rpmFindDependenciesALTLinux(ctx, out, pkgInfo, dirs, "/usr/lib/rpm/find-provides", []string{"RPM_FINDPROV_SKIPLIST=" + strings.Join(skiplist, "\n")}, func(dep string) {
		// slog.Info(gotext.Get("Provided dependency found"), "dep", dep)
		out.Info(gotext.Get("Provided dependency found: %s", dep))
		pkgInfo.Provides = append(pkgInfo.Provides, dep)
	})
}

func (o *ALTLinux) FindRequires(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	return o.rpmFindDependenciesALTLinux(ctx, out, pkgInfo, dirs, "/usr/lib/rpm/find-requires", []string{"RPM_FINDREQ_SKIPLIST=" + strings.Join(skiplist, "\n")}, func(dep string) {
		// slog.Info(gotext.Get("Required dependency found"), "dep", dep)
		out.Info(gotext.Get("Required dependency found: %s", dep))
		pkgInfo.Depends = append(pkgInfo.Depends, dep)
	})
}

func (o *ALTLinux) BuildDepends(ctx context.Context) ([]string, error) {
	return []string{"rpm-build"}, nil
}

func (o *ALTLinux) rpmFindDependenciesALTLinux(ctx context.Context, out output.Output, pkgInfo *nfpm.Info, dirs types.Directories, command string, envs []string, updateFunc func(string)) error {
	if _, err := exec.LookPath(command); err != nil {
		out.Warn(gotext.Get("Command %q not found on the system", command))
		// slog.Info(gotext.Get("Command %q not found on the system"), "command", command)
		return nil
	}

	var paths []string
	for _, content := range pkgInfo.Contents {
		if content.Type != "dir" {
			paths = append(paths,
				path.Join(dirs.PkgDir, content.Destination),
			)
		}
	}

	if len(paths) == 0 {
		return nil
	}

	cmd := exec.CommandContext(ctx, command)
	cmd.Stdin = bytes.NewBufferString(strings.Join(paths, "\n") + "\n")
	cmd.Env = append(cmd.Env,
		"RPM_BUILD_ROOT="+dirs.PkgDir,
		"RPM_BUILD_DIR="+dirs.SrcDir,
		"RPM_FINDPROV_METHOD=",
		"RPM_FINDREQ_METHOD=",
		"RPM_DATADIR=",
		"RPM_SUBPACKAGE_NAME=",
		"HOME="+os.Getenv("HOME"),
	)
	cmd.Env = append(cmd.Env, envs...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		filteredStderr := []string{}
		for _, line := range strings.Split(stderr.String(), "\n") {
			if !strings.HasPrefix(line, "lib.req: WARNING") {
				filteredStderr = append(filteredStderr, line)
			}
		}
		filteredStderrStr := strings.Join(filteredStderr, "\n")
		slog.Error(filteredStderrStr)
		return err
	}

	dependencies := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	for _, dep := range dependencies {
		if dep != "" {
			updateFunc(dep)
		}
	}

	return nil
}
