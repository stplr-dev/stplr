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
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"strings"

	"github.com/goreleaser/nfpm/v2"
	"github.com/leonelquinteros/gotext"

	"go.stplr.dev/stplr/pkg/types"
)

type Fedora struct{}

const fedoraRpmDeps = "/usr/lib/rpm/rpmdeps"

func (o *Fedora) FindProvides(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	args := []string{
		"--define=_use_internal_dependency_generator 1",
		"--provides",
	}
	if len(skiplist) > 0 {
		args = append(args, fmt.Sprintf(
			"--define=__provides_exclude_from %s\"",
			strings.Join(skiplist, "|"),
		))
	}
	return rpmFindDependenciesFedora(
		ctx,
		pkgInfo,
		dirs,
		fedoraRpmDeps,
		args,
		func(dep string) {
			slog.Info(gotext.Get("Provided dependency found"), "dep", dep)
			pkgInfo.Provides = append(pkgInfo.Provides, dep)
		})
}

func (o *Fedora) FindRequires(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, skiplist, filter []string) error {
	args := []string{
		"--define=_use_internal_dependency_generator 1",
		"--requires",
	}
	if len(skiplist) > 0 {
		args = append(args, fmt.Sprintf(
			"--define=__requires_exclude_from %s",
			strings.Join(skiplist, "|"),
		))
	}
	return rpmFindDependenciesFedora(
		ctx,
		pkgInfo,
		dirs,
		fedoraRpmDeps,
		args,
		func(dep string) {
			slog.Info(gotext.Get("Required dependency found"), "dep", dep)
			pkgInfo.Depends = append(pkgInfo.Depends, dep)
		})
}

func (o *Fedora) BuildDepends(ctx context.Context) ([]string, error) {
	return []string{"rpm-build"}, nil
}

func rpmFindDependenciesFedora(ctx context.Context, pkgInfo *nfpm.Info, dirs types.Directories, command string, args []string, updateFunc func(string)) error {
	if _, err := exec.LookPath(command); err != nil {
		slog.Info(gotext.Get("Command not found on the system"), "command", command)
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

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = bytes.NewBufferString(strings.Join(paths, "\n") + "\n")
	cmd.Env = append(cmd.Env,
		"RPM_BUILD_ROOT="+dirs.PkgDir,
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		slog.Error(stderr.String())
		return err
	}
	slog.Debug(stderr.String())

	dependencies := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, dep := range dependencies {
		if dep != "" {
			updateFunc(dep)
		}
	}

	return nil
}
