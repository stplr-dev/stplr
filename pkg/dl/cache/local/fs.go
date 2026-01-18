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

package local

import (
	"os"
	"path/filepath"

	"go.stplr.dev/stplr/pkg/dl/cache"
)

func handleCache(cacheEntry, dest string, t cache.Type) (bool, error) {
	switch t {
	case cache.TypeFile:
		err := os.Link(cacheEntry, dest)
		if err != nil {
			return false, err
		}
		return true, nil
	case cache.TypeDir:
		err := linkDir(cacheEntry, dest)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// Функция linkDir рекурсивно создает жесткие ссылки для файлов из каталога src в каталог dest
func linkDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		newPath := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(newPath, info.Mode())
		}

		return os.Link(path, newPath)
	})
}
