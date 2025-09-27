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
	"fmt"
	"os"
	"path"

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
		if err := unix.Mount("tmpfs", dir, "tmpfs", 0, ""); err != nil {
			return fmt.Errorf("failed to hide dir: %w", err)
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

	return nil
}

func hideExecutable() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	if err := unix.Mount("/dev/null", execPath, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to hide %s: %w", execPath, err)
	}

	return nil
}

func sandboxSocket() error {
	err := os.MkdirAll(constants.SocketDirPath, 0o755)
	if err != nil {
		return err
	}

	if err := unix.Mount("tmpfs", constants.SocketDirPath, "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("failed to hide SocketDirPath %w", err)
	}

	return nil
}

func hideProc() error {
	err := unix.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		return fmt.Errorf("failed to hide /proc: %w", err)
	}

	return nil
}

func Setup(srcDir, pkgDir, homeDir string) error {
	if err := sandboxDirs([]string{srcDir, pkgDir}, []string{constants.SystemCachePath, homeDir}); err != nil {
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
