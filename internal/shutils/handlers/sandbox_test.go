// SPDX-License-Identifier: GPL-3.0-or-later
//
// Stapler
// Copyright (C) 2026 The Stapler Authors
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

package handlers

import (
	"errors"
	"testing"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMounts(t *testing.T) {
	tests := []struct {
		name         string
		homeDir      string
		srcDir       string
		pkgDir       string
		isolatedProc bool
		wantProcType string
		wantProcSrc  string
	}{
		{
			name:         "isolated proc mount",
			homeDir:      "/home/user",
			srcDir:       "/tmp/src",
			pkgDir:       "/tmp/pkg",
			isolatedProc: true,
			wantProcType: "proc",
			wantProcSrc:  "proc",
		},
		{
			name:         "bind mount proc fallback",
			homeDir:      "/home/user",
			srcDir:       "/tmp/src",
			pkgDir:       "/tmp/pkg",
			isolatedProc: false,
			wantProcType: "bind",
			wantProcSrc:  "/proc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mounts := buildMounts(tt.homeDir, tt.srcDir, tt.pkgDir, tt.isolatedProc)

			require.NotEmpty(t, mounts, "buildMounts should return non-empty slice")

			procMount := findMount(mounts, "/proc")
			require.NotNil(t, procMount, "no /proc mount found in mounts")

			assert.Equal(t, tt.wantProcType, procMount.Type, "proc mount type mismatch")

			assert.Equal(t, tt.wantProcSrc, procMount.Source, "proc mount source mismatch")

			if !tt.isolatedProc {
				assert.Contains(t, procMount.Options, "rbind", "bind proc mount should have 'rbind' option")
				assert.Contains(t, procMount.Options, "ro", "bind proc mount should have 'ro' option")
			}

			devMount := findMount(mounts, "/dev")
			require.NotNil(t, devMount, "no /dev mount found in mounts")
			assert.Equal(t, "tmpfs", devMount.Type, "dev mount should be tmpfs")
		})
	}
}

func TestIsMountError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "operation not permitted",
			err:  errors.New("operation not permitted"),
			want: true,
		},
		{
			name: "OPERATION NOT PERMITTED (uppercase)",
			err:  errors.New("OPERATION NOT PERMITTED"),
			want: true,
		},
		{
			name: "permission denied",
			err:  errors.New("permission denied"),
			want: true,
		},
		{
			name: "Permission Denied (mixed case)",
			err:  errors.New("Permission Denied"),
			want: true,
		},
		{
			name: "mount proc error",
			err:  errors.New("error mounting proc to /proc"),
			want: true,
		},
		{
			name: "mount error without proc",
			err:  errors.New("error mounting tmpfs"),
			want: false,
		},
		{
			name: "proc error without mount",
			err:  errors.New("proc filesystem error"),
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "nested container error",
			err:  errors.New("unable to start container process: error during container init: error mounting \"proc\" to rootfs at \"/proc\": mount src=proc, dst=/proc, dstFd=/proc/thread-self/fd/11: operation not permitted"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMountError(tt.err)
			assert.Equal(t, tt.want, got, "isMountError() result mismatch")
		})
	}
}

func TestBuildMountsStructure(t *testing.T) {
	homeDir := "/home/testuser"
	srcDir := "/tmp/test-src"
	pkgDir := "/tmp/test-pkg"

	t.Run("check essential mounts present", func(t *testing.T) {
		mounts := buildMounts(homeDir, srcDir, pkgDir, true)

		essentialMounts := []string{"/proc", "/dev", homeDir, srcDir, pkgDir}

		for _, essential := range essentialMounts {
			assert.NotNil(t, findMount(mounts, essential),
				"essential mount %q should be present", essential)
		}
	})

	t.Run("isolated proc has no options", func(t *testing.T) {
		mounts := buildMounts(homeDir, srcDir, pkgDir, true)

		procMount := findMount(mounts, "/proc")
		require.NotNil(t, procMount)

		if procMount.Type == "proc" {
			assert.Empty(t, procMount.Options,
				"isolated proc mount should have no options")
		}
	})

	t.Run("bind proc is readonly", func(t *testing.T) {
		mounts := buildMounts(homeDir, srcDir, pkgDir, false)

		procMount := findMount(mounts, "/proc")
		require.NotNil(t, procMount)

		if procMount.Type == "bind" {
			assert.Contains(t, procMount.Options, "ro",
				"bind proc mount should be readonly")
		}
	})

	t.Run("dev mount has correct options", func(t *testing.T) {
		mounts := buildMounts(homeDir, srcDir, pkgDir, true)

		devMount := findMount(mounts, "/dev")
		require.NotNil(t, devMount)

		assert.Equal(t, "tmpfs", devMount.Type)
		assert.Contains(t, devMount.Options, "nosuid")
		assert.Contains(t, devMount.Options, "strictatime")
		assert.Contains(t, devMount.Options, "mode=755")
		assert.Contains(t, devMount.Options, "size=65536k")
	})
}

func TestBuildMountsConsistency(t *testing.T) {
	homeDir := "/home/testuser"
	srcDir := "/tmp/test-src"
	pkgDir := "/tmp/test-pkg"

	t.Run("isolated and bind produce same number of mounts", func(t *testing.T) {
		isolatedMounts := buildMounts(homeDir, srcDir, pkgDir, true)
		bindMounts := buildMounts(homeDir, srcDir, pkgDir, false)

		assert.Equal(t, len(isolatedMounts), len(bindMounts),
			"both mount modes should produce same number of mounts")
	})

	t.Run("all mounts have destination", func(t *testing.T) {
		mounts := buildMounts(homeDir, srcDir, pkgDir, true)

		for i, mount := range mounts {
			assert.NotEmpty(t, mount.Destination,
				"mount at index %d should have destination", i)
		}
	})

	t.Run("no duplicate destinations", func(t *testing.T) {
		mounts := buildMounts(homeDir, srcDir, pkgDir, true)

		destinations := make(map[string]bool)
		for _, mount := range mounts {
			assert.False(t, destinations[mount.Destination],
				"duplicate mount destination: %s", mount.Destination)
			destinations[mount.Destination] = true
		}
	})
}

// Helper function
func findMount(mounts []specs.Mount, destination string) *specs.Mount {
	for i := range mounts {
		if mounts[i].Destination == destination {
			return &mounts[i]
		}
	}
	return nil
}
