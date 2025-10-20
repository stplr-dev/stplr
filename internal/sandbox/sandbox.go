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

package sandbox

import (
	"errors"
	"fmt"
	"os"
	"path"
	"syscall"

	"golang.org/x/sys/unix"

	"go.stplr.dev/stplr/internal/constants"
)

func sandboxDirs(keep, hide []string) error {
	tmpKeep := make([]string, len(keep))

	for i, dir := range keep {
		tmpKeep[i] = path.Join(os.TempDir(), fmt.Sprintf(".stplr-build-src-%d", i))

		if err := os.MkdirAll(tmpKeep[i], 0o755); err != nil {
			return fmt.Errorf("failed to create tmpDir: %w", err)
		}

		if err := unix.Mount(dir, tmpKeep[i], "", unix.MS_BIND, ""); err != nil {
			return fmt.Errorf("failed to backup dir: %w", err)
		}
	}

	for _, dir := range hide {
		flags := uintptr(unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC)
		if err := unix.Mount("tmpfs", dir, "tmpfs", flags, "size=64M"); err != nil {
			return fmt.Errorf("hide %s: %w", dir, err)
		}
	}

	for i, dir := range keep {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create dir: %w", err)
		}
		if err := unix.Mount(tmpKeep[i], dir, "", unix.MS_BIND, ""); err != nil {
			return fmt.Errorf("failed to restore dir: %w", err)
		}
		if err := unix.Unmount(tmpKeep[i], 0); err != nil {
			return fmt.Errorf("failed to unmount tmpKeep: %w", err)
		}
	}

	flags := uintptr(unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC)
	if err := unix.Mount("tmpfs", "/tmp", "tmpfs", flags, "size=64M"); err != nil {
		return fmt.Errorf("hide %s: %w", "/tmp", err)
	}

	return nil
}

func hideExecutable() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	if err := unix.Mount("/dev/null", execPath, "", unix.MS_RDONLY|unix.MS_BIND, ""); err != nil {
		if errors.Is(err, syscall.ENOENT) {
			return nil
		}
		return fmt.Errorf("failed to hide %s: %w", execPath, err)
	}

	return nil
}

func sandboxSocket() error {
	err := os.MkdirAll(constants.SocketDirPath, 0o755)
	if err != nil {
		return err
	}

	flags := unix.MS_NODEV | unix.MS_NOSUID
	if err := unix.Mount("tmpfs", constants.SocketDirPath, "tmpfs", uintptr(flags), "size=10M"); err != nil {
		return fmt.Errorf("hide socket dir: %w", err)
	}

	return nil
}

func hideProc() error {
	err := unix.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return fmt.Errorf("failed to hide /proc: %w", err)
	}

	if err := unix.Mount("", "/proc", "", unix.MS_REMOUNT|unix.MS_RDONLY|unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("protect proc: %w", err)
	}

	return nil
}

func Setup(srcDir, pkgDir, homeDir string) error {
	err := unix.Mount("", "/", "/", unix.MS_PRIVATE|unix.MS_REC, "")
	if err != nil {
		return fmt.Errorf("failed to isolate root mounts: %w", err)
	}

	if err := sandboxDirs(
		[]string{srcDir, pkgDir},
		[]string{
			constants.SystemCachePath,
			homeDir,
			"/run",
			"/var/run",
			"/var/log",
			"/dev/shm",
		},
	); err != nil {
		return err
	}

	if err := sandboxSocket(); err != nil {
		return err
	}

	if err := hideExecutable(); err != nil {
		return err
	}

	if err := hideProc(); err != nil {
		return err
	}

	return nil
}
