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

func sandboxSystemCacheDir(srcDir, pkgDir string) error {
	tmpSrc := path.Join(os.TempDir(), "stplr-build-src")
	tmpPkg := path.Join(os.TempDir(), "stplr-build-pkg")

	if err := os.MkdirAll(tmpSrc, 0o755); err != nil {
		return fmt.Errorf("failed to create tmpSrc: %w", err)
	}

	if err := os.MkdirAll(tmpPkg, 0o755); err != nil {
		return fmt.Errorf("failed to create tmpPkg: %w", err)
	}

	if err := unix.Mount(srcDir, tmpSrc, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to backup srcDir: %w", err)
	}

	if err := unix.Mount(pkgDir, tmpPkg, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to backup pkgDir: %w", err)
	}

	if err := unix.Mount("tmpfs", constants.SystemCachePath, "tmpfs", 0, ""); err != nil {
		return fmt.Errorf("failed to hide SystemCachePath: %w", err)
	}

	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return fmt.Errorf("failed to create srcDir: %w", err)
	}

	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return fmt.Errorf("failed to create pkgDir: %w", err)
	}

	if err := unix.Mount(tmpSrc, srcDir, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to restore srcDir: %w", err)
	}

	if err := unix.Mount(tmpPkg, pkgDir, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("failed to restore pkgDir: %w", err)
	}

	if err := unix.Unmount(tmpSrc, 0); err != nil {
		return fmt.Errorf("failed to unmount tmpPkg: %w", err)
	}

	if err := unix.Unmount(tmpPkg, 0); err != nil {
		return fmt.Errorf("failed to unmount tmpPkg: %w", err)
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

func Setup(srcDir, pkgDir, tmpDir string) error {
	if err := sandboxSystemCacheDir(srcDir, pkgDir); err != nil {
		return err
	}

	if err := sandboxSocket(); err != nil {
		return err
	}

	if err := hideExecutable(); err != nil {
		return err
	}

	return nil
}
