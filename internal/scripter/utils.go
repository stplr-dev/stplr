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

package scripter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/arch"
	_ "github.com/goreleaser/nfpm/v2/deb"

	// rpm packager with modifications from our fork
	_ "github.com/goreleaser/nfpm/v2/rpm-lowmem"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"

	"go.stplr.dev/stplr/internal/commonbuild"
	"go.stplr.dev/stplr/internal/cpu"
	"go.stplr.dev/stplr/pkg/overrides"
	"go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

func prepareDirs(dirs types.Directories) error {
	err := os.RemoveAll(dirs.BaseDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dirs.SrcDir, 0o755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dirs.PkgDir, 0o755)
	if err != nil {
		return err
	}
	return os.MkdirAll(dirs.HomeDir, 0o755)
}

func isDirEmpty(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}

func createDirContent(fi os.FileInfo, path, trimmed string, prefered bool) (*files.Content, error) {
	if !prefered && !isDirEmpty(path) {
		return nil, nil
	}

	return &files.Content{
		Source:      path,
		Destination: trimmed,
		Type:        "dir",
		FileInfo: &files.ContentFileInfo{
			MTime: fi.ModTime(),
		},
	}, nil
}

func createSymlinkContent(fi os.FileInfo, path, trimmed, pkgDir string) (*files.Content, error) {
	link, err := os.Readlink(path)
	if err != nil {
		return nil, err
	}
	link = strings.TrimPrefix(link, pkgDir)

	return &files.Content{
		Source:      link,
		Destination: trimmed,
		Type:        "symlink",
		FileInfo: &files.ContentFileInfo{
			MTime: fi.ModTime(),
			Mode:  fi.Mode(),
		},
	}, nil
}

func createFileContent(fi os.FileInfo, path, trimmed string, configFiles []string) (*files.Content, error) {
	content := &files.Content{
		Source:      path,
		Destination: trimmed,
		FileInfo: &files.ContentFileInfo{
			MTime: fi.ModTime(),
			Mode:  fi.Mode(),
			Size:  fi.Size(),
		},
	}

	if slices.Contains(configFiles, trimmed) {
		content.Type = "config|noreplace"
	}

	return content, nil
}

func processPreferredContents(preferedContents []string, pkgDir string, configFiles []string) ([]*files.Content, error) {
	contents := []*files.Content{}
	for _, trimmed := range preferedContents {
		path := filepath.Join(pkgDir, trimmed)
		content, err := processPath(path, trimmed, true, pkgDir, configFiles)
		if err != nil {
			return nil, err
		}
		if content != nil {
			contents = append(contents, content)
		}
	}
	return contents, nil
}

func processAllContents(pkgDir string, configFiles []string) ([]*files.Content, error) {
	contents := []*files.Content{}
	err := filepath.Walk(pkgDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		trimmed := strings.TrimPrefix(path, pkgDir)
		content, err := processPath(path, trimmed, true, pkgDir, configFiles)
		if err != nil {
			return err
		}
		if content != nil {
			contents = append(contents, content)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return contents, nil
}

func processPath(path, trimmed string, prefered bool, pkgDir string, configFiles []string) (*files.Content, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return createDirContent(fi, path, trimmed, prefered)
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		return createSymlinkContent(fi, path, trimmed, pkgDir)
	}

	return createFileContent(fi, path, trimmed, configFiles)
}

// Функция buildContents создает секцию содержимого пакета, которая содержит файлы,
// которые будут включены в конечный пакет.
func buildContents(dirs types.Directories, preferedContents *[]string, configFiles []string) ([]*files.Content, error) {
	if preferedContents != nil {
		return processPreferredContents(*preferedContents, dirs.PkgDir, configFiles)
	}

	return processAllContents(dirs.PkgDir, configFiles)
}

type getBasePkgInfoInput interface {
	commonbuild.OSReleaser
	commonbuild.RepositoryGetter
	commonbuild.BuildOptsProvider
}

func FormatName(name, repo string) string {
	return fmt.Sprintf("%s+stplr-%s", name, repo)
}

func GetBasePkgInfo(pkg *staplerfile.Package, input getBasePkgInfoInput) *nfpm.Info {
	var name string
	if input.BuildOpts().NoSuffix {
		name = pkg.Name
	} else {
		name = FormatName(pkg.Name, input.Repository())
	}
	return &nfpm.Info{
		Name:    name,
		Arch:    cpu.Arch(),
		Version: pkg.Version,
		Release: overrides.ReleasePlatformSpecific(pkg.Release, input.OSRelease()),
		Epoch:   strconv.FormatUint(uint64(pkg.Epoch), 10),
	}
}

// Функция setScripts добавляет скрипты-перехватчики к метаданным пакета.
func setScripts(vars *staplerfile.Package, info *nfpm.Info, scriptDir string) {
	if vars.Scripts.Resolved().PreInstall != "" {
		info.Scripts.PreInstall = filepath.Join(scriptDir, vars.Scripts.Resolved().PreInstall)
	}

	if vars.Scripts.Resolved().PostInstall != "" {
		info.Scripts.PostInstall = filepath.Join(scriptDir, vars.Scripts.Resolved().PostInstall)
	}

	if vars.Scripts.Resolved().PreRemove != "" {
		info.Scripts.PreRemove = filepath.Join(scriptDir, vars.Scripts.Resolved().PreRemove)
	}

	if vars.Scripts.Resolved().PostRemove != "" {
		info.Scripts.PostRemove = filepath.Join(scriptDir, vars.Scripts.Resolved().PostRemove)
	}

	if vars.Scripts.Resolved().PreUpgrade != "" {
		info.ArchLinux.Scripts.PreUpgrade = filepath.Join(scriptDir, vars.Scripts.Resolved().PreUpgrade)
		info.APK.Scripts.PreUpgrade = filepath.Join(scriptDir, vars.Scripts.Resolved().PreUpgrade)
	}

	if vars.Scripts.Resolved().PostUpgrade != "" {
		info.ArchLinux.Scripts.PostUpgrade = filepath.Join(scriptDir, vars.Scripts.Resolved().PostUpgrade)
		info.APK.Scripts.PostUpgrade = filepath.Join(scriptDir, vars.Scripts.Resolved().PostUpgrade)
	}

	if vars.Scripts.Resolved().PreTrans != "" {
		info.RPM.Scripts.PreTrans = filepath.Join(scriptDir, vars.Scripts.Resolved().PreTrans)
	}

	if vars.Scripts.Resolved().PostTrans != "" {
		info.RPM.Scripts.PostTrans = filepath.Join(scriptDir, vars.Scripts.Resolved().PostTrans)
	}
}

func Map[T, R any](items []T, f func(T) R) []R {
	res := make([]R, len(items))
	for i, item := range items {
		res[i] = f(item)
	}
	return res
}

func GetBuiltName(deps []*commonbuild.BuiltDep) []string {
	return Map(deps, func(dep *commonbuild.BuiltDep) string {
		return dep.Name
	})
}
