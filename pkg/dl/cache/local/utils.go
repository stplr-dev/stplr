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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"go.stplr.dev/stplr/pkg/dl/cache"
)

func hashUrl(url string) string {
	h := sha256.New()
	_, err := io.WriteString(h, url)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func toCacheId(id int64) cache.CacheID {
	return cache.CacheID(fmt.Sprintf("%d", id))
}

func fromCacheId(cacheId cache.CacheID) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(string(cacheId), "%d", &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
