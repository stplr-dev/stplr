// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package distro

import (
	"context"
	"os"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"go.stplr.dev/stplr/internal/shutils/handlers"
)

// OSRelease contains information from an os-release file
type OSRelease struct {
	Name             string
	PrettyName       string
	ID               string
	Like             []string
	VersionID        string
	ANSIColor        string
	HomeURL          string
	DocumentationURL string
	SupportURL       string
	BugReportURL     string
	Logo             string
	PlatformID       string
}

var parsed *OSRelease

// OSReleaseName returns a struct parsed from the system's os-release
// file. It checks /etc/os-release as well as /usr/lib/os-release.
// The first time it's called, it'll parse the os-release file.
// Subsequent calls will return the same value.
func ParseOSRelease(ctx context.Context) (*OSRelease, error) {
	if parsed != nil {
		return parsed, nil
	}

	fl, err := os.Open("/usr/lib/os-release")
	if err != nil {
		fl, err = os.Open("/etc/os-release")
		if err != nil {
			return nil, err
		}
	}

	file, err := syntax.NewParser().Parse(fl, "/usr/lib/os-release")
	if err != nil {
		return nil, err
	}

	fl.Close()

	// Create new shell interpreter with nop open, exec, readdir, and stat handlers
	// as well as no environment variables in order to prevent vulnerabilities
	// caused by changing the os-release file.
	runner, err := interp.New(
		interp.OpenHandler(handlers.NopOpen),
		interp.ExecHandler(handlers.NopExec),
		interp.ReadDirHandler2(handlers.NopReadDir),
		interp.StatHandler(handlers.NopStat),
		interp.Env(expand.ListEnviron()),
		interp.Dir("/"),
	)
	if err != nil {
		return nil, err
	}

	err = runner.Run(ctx, file)
	if err != nil {
		return nil, err
	}

	out := &OSRelease{
		Name:             runner.Vars["NAME"].Str,
		PrettyName:       runner.Vars["PRETTY_NAME"].Str,
		ID:               runner.Vars["ID"].Str,
		VersionID:        runner.Vars["VERSION_ID"].Str,
		ANSIColor:        runner.Vars["ANSI_COLOR"].Str,
		HomeURL:          runner.Vars["HOME_URL"].Str,
		DocumentationURL: runner.Vars["DOCUMENTATION_URL"].Str,
		SupportURL:       runner.Vars["SUPPORT_URL"].Str,
		BugReportURL:     runner.Vars["BUG_REPORT_URL"].Str,
		Logo:             runner.Vars["LOGO"].Str,
		PlatformID:       runner.Vars["PLATFORM_ID"].Str,
	}

	distroUpdated := false
	if distID, ok := os.LookupEnv("STPLR_DISTRO"); ok {
		out.ID = distID
	}

	if distLike, ok := os.LookupEnv("STPLR_DISTRO_LIKE"); ok {
		out.Like = strings.Split(distLike, " ")
	} else if runner.Vars["ID_LIKE"].IsSet() && !distroUpdated {
		out.Like = strings.Split(runner.Vars["ID_LIKE"].Str, " ")
	}

	parsed = out
	return out, nil
}
