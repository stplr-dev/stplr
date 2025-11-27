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
	"errors"
	"io"
	"os"
	"regexp"
	"slices"
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

	ReleaseID string // Major version (RHEL-like), codename (Ubuntu-like) or other
}

var parsed *OSRelease

// OSReleaseName returns a struct parsed from the system's os-release
// file. It checks /etc/os-release as well as /usr/lib/os-release.
// The first time it's called, it'll parse the os-release file.
// Subsequent calls will return the same value.
func ParseOSRelease(ctx context.Context) (*OSRelease, error) {
	// NOTE: The use of the global `parsed` variable is not a clean architectural choice,
	// but it has been this way historically. Keeping it as is for now to avoid breaking behavior.
	if parsed != nil {
		return parsed, nil
	}
	paths := []string{"/usr/lib/os-release", "/etc/os-release"}
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}

		out, err := parseOSReleaseFromFile(ctx, f)
		if err == nil {
			parsed = out
			return out, nil
		}
	}
	return nil, errors.New("couldn't open or find os-release file")
}

func parseOSReleaseFromFile(ctx context.Context, f io.Reader) (*OSRelease, error) {
	file, err := syntax.NewParser().Parse(f, "/usr/lib/os-release")
	if err != nil {
		return nil, err
	}

	runner, err := newShellRunner()
	if err != nil {
		return nil, err
	}

	err = runner.Run(ctx, file)
	if err != nil {
		return nil, err
	}

	return parseOSReleaseFromRunner(runner)
}

func newShellRunner() (*interp.Runner, error) {
	// Create new shell interpreter with nop open, exec, readdir, and stat handlers
	// as well as no environment variables in order to prevent vulnerabilities
	return interp.New(
		interp.OpenHandler(handlers.NopOpen),
		interp.ExecHandler(handlers.NopExec),
		interp.ReadDirHandler2(handlers.NopReadDir),
		interp.StatHandler(handlers.NopStat),
		interp.Env(expand.ListEnviron()),
		interp.Dir("/"),
	)
}

func parseOSReleaseFromRunner(runner *interp.Runner) (*OSRelease, error) {
	vars := runner.Vars
	out := &OSRelease{
		Name:             vars["NAME"].Str,
		PrettyName:       vars["PRETTY_NAME"].Str,
		ID:               vars["ID"].Str,
		VersionID:        vars["VERSION_ID"].Str,
		ANSIColor:        vars["ANSI_COLOR"].Str,
		HomeURL:          vars["HOME_URL"].Str,
		DocumentationURL: vars["DOCUMENTATION_URL"].Str,
		SupportURL:       vars["SUPPORT_URL"].Str,
		BugReportURL:     vars["BUG_REPORT_URL"].Str,
		Logo:             vars["LOGO"].Str,
		PlatformID:       vars["PLATFORM_ID"].Str,
	}

	distroUpdated := false
	if distID, ok := os.LookupEnv("STPLR_DISTRO"); ok {
		out.ID = distID
		distroUpdated = true
	}

	if distLike, ok := os.LookupEnv("STPLR_DISTRO_LIKE"); ok {
		out.Like = strings.Split(distLike, " ")
	} else if vars["ID_LIKE"].IsSet() && !distroUpdated {
		out.Like = strings.Split(vars["ID_LIKE"].Str, " ")
	}

	setReleaseID(runner, out)

	return out, nil
}

func IsIdEqualOrLike(info *OSRelease, id string) bool {
	return info.ID == id || slices.Contains(info.Like, id)
}

func parseRHELPlatfrom(platform string) string {
	re := regexp.MustCompile(`\d+`)
	return re.FindString(platform)
}

func setReleaseID(runner *interp.Runner, info *OSRelease) {
	switch {
	case IsIdEqualOrLike(info, "altlinux"):
		info.ReleaseID = runner.Vars["ALT_BRANCH_ID"].Str

	case IsIdEqualOrLike(info, "rhel"), IsIdEqualOrLike(info, "fedora"):
		info.ReleaseID = parseRHELPlatfrom(runner.Vars["PLATFORM_ID"].Str)

	case IsIdEqualOrLike(info, "debian"), IsIdEqualOrLike(info, "ubuntu"):
		info.ReleaseID = runner.Vars["VERSION_CODENAME"].Str
	}

	if info.ReleaseID == "" {
		if ver := runner.Vars["VERSION_ID"].Str; ver != "" {
			if i := strings.Index(ver, "."); i > 0 {
				info.ReleaseID = ver[:i]
			} else {
				info.ReleaseID = ver
			}
		}
	}

	re := regexp.MustCompile(`[^A-Za-z0-9_]`)
	info.ReleaseID = re.ReplaceAllString(info.ReleaseID, "_")
}
