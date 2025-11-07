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

package copier

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/constants"
	"go.stplr.dev/stplr/internal/osutils"
	"go.stplr.dev/stplr/internal/utils"
	"go.stplr.dev/stplr/pkg/distro"
	"go.stplr.dev/stplr/pkg/staplerfile"
)

type Copier struct {
	uid int
	gid int
	wd  string

	isRoot bool

	cachePath string
}

func New(uid, gid int, wd string) (*Copier, error) {
	var err error
	cachePath := ""
	if utils.IsRoot() {
		cachePath = constants.SystemCachePath
	} else {
		cachePath, err = os.UserCacheDir()
		if err != nil {
			return nil, err
		}
	}

	return &Copier{
		uid: uid,
		gid: gid,
		wd:  wd,

		isRoot:    utils.IsRoot(),
		cachePath: cachePath,
	}, nil
}

func (e *Copier) Copy(ctx context.Context, f *staplerfile.ScriptFile, info *distro.OSRelease) (string, error) {
	externalFiles, err := f.ExternalFiles(context.Background(), info)
	if err != nil {
		return "", err
	}

	scriptPath := f.Path()
	scriptFile := filepath.Base(scriptPath)
	scriptDir := filepath.Dir(scriptPath)

	tmpdir, err := e.tmpdir()
	if err != nil {
		return "", err
	}

	keep := false
	defer e.cleanupTempDir(tmpdir, &keep)

	newScriptPath := filepath.Join(tmpdir, scriptFile)

	if err := e.copy(scriptPath, newScriptPath); err != nil {
		return "", err
	}
	for _, ef := range externalFiles {
		if err := e.processExternalFile(ef, scriptDir, tmpdir); err != nil {
			return "", err
		}
	}
	err = e.chownToBuilder(tmpdir)
	if err != nil {
		return "", err
	}

	keep = true
	return newScriptPath, nil
}

func (e *Copier) CopyOut(ctx context.Context, pkgs []commonbuild.BuiltDep) error {
	for _, pkg := range pkgs {
		name := filepath.Base(pkg.Path)
		if err := e.copyOut(ctx, pkg.Path, filepath.Join(e.wd, name), 0, 0); err != nil {
			return err
		}
	}
	return nil
}

func (e *Copier) copyOut(_ context.Context, from, dest string, uid, gid int) error {
	if err := osutils.Move(from, dest); err != nil {
		return err
	}
	if err := e.chown(dest, uid, gid); err != nil {
		return err
	}
	return nil
}

func (e *Copier) processExternalFile(ef, scriptDir, tmpdir string) error {
	src := filepath.Join(scriptDir, ef)
	ok, err := isPathWithinBase(scriptDir, src)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("external file %q is outside script directory", ef)
	}
	return e.copy(src, filepath.Join(tmpdir, ef))
}

func (e *Copier) copy(srcfile, destfile string) error {
	if err := os.MkdirAll(filepath.Dir(destfile), 0o755); err != nil {
		return err
	}
	if err := osutils.CopyFile(srcfile, destfile); err != nil {
		return err
	}
	return nil
}

func generateID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (e *Copier) tmpdir() (string, error) {
	id := generateID()
	build := filepath.Join(e.cachePath, "build")
	tmp := filepath.Join(build, id)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return "", err
	}
	if err := e.chownToBuilder(build); err != nil {
		return "", err
	}
	return tmp, nil
}

func (e *Copier) cleanupTempDir(tmpdir string, keep *bool) {
	if r := recover(); r != nil || !*keep {
		if derr := os.RemoveAll(tmpdir); derr != nil {
			slog.Warn("failed to remove tmpdir %q: %v", tmpdir, derr)
		}
		if r != nil {
			panic(r)
		}
	}
}

func (e *Copier) chownToBuilder(path string) error {
	uid, gid, err := utils.GetUidGidStaplerUser()
	if err != nil {
		return err
	}
	return e.chown(path, uid, gid)
}

func (e *Copier) chown(path string, uid, gid int) error {
	if !e.isRoot {
		return nil
	}
	uid, gid, err := utils.GetUidGidStaplerUser()
	if err != nil {
		return err
	}
	walkFn := func(p string, info os.FileInfo, err error) error {
		return os.Chown(p, uid, gid)
	}
	return filepath.Walk(path, walkFn)
}

func isPathWithinBase(basePath, targetPath string) (bool, error) {
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false, fmt.Errorf("cannot resolve base path: %w", err)
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false, fmt.Errorf("cannot resolve target path: %w", err)
	}

	absBase, err = filepath.EvalSymlinks(absBase)
	if err != nil {
		return false, err
	}
	absTarget, err = filepath.EvalSymlinks(absTarget)
	if err != nil {
		return false, err
	}

	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false, fmt.Errorf("cannot compute relative path: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}
