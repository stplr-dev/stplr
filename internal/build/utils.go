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

package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	// Импортируем пакеты для поддержки различных форматов пакетов (APK, DEB, RPM и ARCH).

	_ "github.com/goreleaser/nfpm/v2/apk"
	_ "github.com/goreleaser/nfpm/v2/arch"
	_ "github.com/goreleaser/nfpm/v2/deb"
	_ "github.com/goreleaser/nfpm/v2/rpm"

	"github.com/goreleaser/nfpm/v2"
	"github.com/goreleaser/nfpm/v2/files"

	"go.stplr.dev/stplr/internal/cpu"
	"go.stplr.dev/stplr/internal/manager"
	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/pkg/distro"
	alrsh "go.stplr.dev/stplr/pkg/staplerfile"
	"go.stplr.dev/stplr/pkg/types"
)

// Функция prepareDirs подготавливает директории для сборки.
func prepareDirs(dirs types.Directories) error {
	err := os.RemoveAll(dirs.BaseDir) // Удаляем базовую директорию, если она существует
	if err != nil {
		return err
	}
	err = os.MkdirAll(dirs.SrcDir, 0o755) // Создаем директорию для источников
	if err != nil {
		return err
	}
	return os.MkdirAll(dirs.PkgDir, 0o755) // Создаем директорию для пакетов
}

// Функция buildContents создает секцию содержимого пакета, которая содержит файлы,
// которые будут включены в конечный пакет.
func buildContents(vars *alrsh.Package, dirs types.Directories, preferedContents *[]string) ([]*files.Content, error) {
	contents := []*files.Content{}

	processPath := func(path, trimmed string, prefered bool) error {
		fi, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if !prefered {
				_, err = f.Readdirnames(1)
				if err != io.EOF {
					return nil
				}
			}

			contents = append(contents, &files.Content{
				Source:      path,
				Destination: trimmed,
				Type:        "dir",
				FileInfo: &files.ContentFileInfo{
					MTime: fi.ModTime(),
				},
			})
			return nil
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			link = strings.TrimPrefix(link, dirs.PkgDir)

			contents = append(contents, &files.Content{
				Source:      link,
				Destination: trimmed,
				Type:        "symlink",
				FileInfo: &files.ContentFileInfo{
					MTime: fi.ModTime(),
					Mode:  fi.Mode(),
				},
			})
			return nil
		}

		fileContent := &files.Content{
			Source:      path,
			Destination: trimmed,
			FileInfo: &files.ContentFileInfo{
				MTime: fi.ModTime(),
				Mode:  fi.Mode(),
				Size:  fi.Size(),
			},
		}

		if slices.Contains(vars.Backup.Resolved(), trimmed) {
			fileContent.Type = "config|noreplace"
		}

		contents = append(contents, fileContent)
		return nil
	}

	if preferedContents != nil {
		for _, trimmed := range *preferedContents {
			path := filepath.Join(dirs.PkgDir, trimmed)
			if err := processPath(path, trimmed, true); err != nil {
				return nil, err
			}
		}
	} else {
		err := filepath.Walk(dirs.PkgDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			trimmed := strings.TrimPrefix(path, dirs.PkgDir)
			return processPath(path, trimmed, false)
		})
		if err != nil {
			return nil, err
		}
	}

	return contents, nil
}

func normalizeContents(contents []*files.Content) {
	for _, content := range contents {
		content.Destination = filepath.Join("/", content.Destination)
	}
}

var RegexpALRPackageName = regexp.MustCompile(`^(?P<package>[^+]+)\+stplr-(?P<repo>.+)$`)

func getBasePkgInfo(vars *alrsh.Package, input interface {
	RepositoryProvider
	OsInfoProvider
},
) *nfpm.Info {
	return &nfpm.Info{
		Name:    fmt.Sprintf("%s+stplr-%s", vars.Name, input.Repository()),
		Arch:    cpu.Arch(),
		Version: vars.Version,
		Release: overrides.ReleasePlatformSpecific(vars.Release, input.OSRelease()),
		Epoch:   strconv.FormatUint(uint64(vars.Epoch), 10),
	}
}

// Функция getPkgFormat возвращает формат пакета из менеджера пакетов,
// или ALR_PKG_FORMAT, если он установлен.
func GetPkgFormat(mgr manager.Manager) string {
	pkgFormat := mgr.Format()
	if format, ok := os.LookupEnv("ALR_PKG_FORMAT"); ok {
		pkgFormat = format
	}
	return pkgFormat
}

// Функция createBuildEnvVars создает переменные окружения, которые будут установлены
// в скрипте сборки при его выполнении.
func createBuildEnvVars(info *distro.OSRelease, dirs types.Directories) []string {
	env := os.Environ()

	env = append(
		env,
		"DISTRO_NAME="+info.Name,
		"DISTRO_PRETTY_NAME="+info.PrettyName,
		"DISTRO_ID="+info.ID,
		"DISTRO_VERSION_ID="+info.VersionID,
		"DISTRO_ID_LIKE="+strings.Join(info.Like, " "),
		"ARCH="+cpu.Arch(),
		"NCPU="+strconv.Itoa(runtime.NumCPU()),
	)

	if dirs.ScriptDir != "" {
		env = append(env, "scriptdir="+dirs.ScriptDir)
	}

	if dirs.PkgDir != "" {
		env = append(env, "pkgdir="+dirs.PkgDir)
	}

	if dirs.SrcDir != "" {
		env = append(env, "srcdir="+dirs.SrcDir)
	}

	return env
}

// Функция setScripts добавляет скрипты-перехватчики к метаданным пакета.
func setScripts(vars *alrsh.Package, info *nfpm.Info, scriptDir string) {
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

/*
// Функция setVersion изменяет переменную версии в скрипте runner.
// Она используется для установки версии на вывод функции version().
func setVersion(ctx context.Context, r *interp.Runner, to string) error {
	fl, err := syntax.NewParser().Parse(strings.NewReader("version='"+to+"'"), "")
	if err != nil {
		return err
	}
	return r.Run(ctx, fl)
}
*/

// Функция packageNames возвращает имена всех предоставленных пакетов.
/*
func packageNames(pkgs []db.Package) []string {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	return names
}
*/

// Функция removeDuplicates убирает любые дубликаты из предоставленного среза.
func removeDuplicates[T comparable](slice []T) []T {
	seen := map[T]struct{}{}
	result := []T{}

	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

func removeDuplicatesSources(sources, checksums []string) ([]string, []string) {
	seen := map[string]string{}
	keys := make([]string, 0)
	for i, s := range sources {
		if val, ok := seen[s]; !ok || strings.EqualFold(val, "SKIP") {
			if !ok {
				keys = append(keys, s)
			}
			seen[s] = checksums[i]
		}
	}

	newSources := make([]string, len(keys))
	newChecksums := make([]string, len(keys))
	for i, k := range keys {
		newSources[i] = k
		newChecksums[i] = seen[k]
	}
	return newSources, newChecksums
}
