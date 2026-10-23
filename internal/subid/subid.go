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

// Package subid reads subordinate id ranges from /etc/subuid and /etc/subgid.
package subid

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	UIDFile = "/etc/subuid"
	GIDFile = "/etc/subgid"
)

type Range struct {
	Owner string
	Start uint32
	Size  uint32
}

func (r Range) End() uint64 {
	return uint64(r.Start) + uint64(r.Size)
}

// Parse skips comments and malformed lines, as shadow does.
func Parse(r io.Reader) ([]Range, error) {
	var ranges []Range

	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 3 {
			slog.Debug("subid: skipping malformed line", "line", lineNum)
			continue
		}

		start, startErr := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 32)
		size, sizeErr := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 32)
		if startErr != nil || sizeErr != nil || size == 0 {
			slog.Debug("subid: skipping malformed line", "line", lineNum)
			continue
		}

		ranges = append(ranges, Range{Owner: parts[0], Start: uint32(start), Size: uint32(size)})
	}

	return ranges, scanner.Err()
}

func ReadFile(path string) ([]Range, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Parse(f)
}

// Owned reads the ranges of the given user; a missing file means no ranges.
func Owned(path, username, uid string) ([]Range, error) {
	ranges, err := ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return OwnedBy(ranges, username, uid), nil
}

// OwnedBy matches owners by name or numeric uid, since both forms are valid.
func OwnedBy(ranges []Range, username, uid string) []Range {
	var owned []Range
	for _, r := range ranges {
		if r.Owner == username || r.Owner == uid {
			owned = append(owned, r)
		}
	}
	return owned
}

// NextFree returns the start of a range of the given size placed after every
// existing range, so it cannot overlap any of them. The last id (2^32-1) is
// invalid on Linux, hence the bound.
func NextFree(ranges []Range, minStart, size uint64) (uint64, bool) {
	start := minStart
	for _, r := range ranges {
		start = max(start, r.End())
	}

	if start+size > math.MaxUint32 {
		return 0, false
	}
	return start, true
}

type IDMap struct {
	ContainerID uint32
	HostID      uint32
	Size        uint32
}

// Mappings maps container root to hostID and places the ranges right after
// it. Ranges that don't fit into the 32-bit id space are skipped.
func Mappings(hostID uint32, ranges []Range) []IDMap {
	maps := []IDMap{{ContainerID: 0, HostID: hostID, Size: 1}}
	var next uint32 = 1

	for _, r := range ranges {
		if r.End() > math.MaxUint32 || uint64(next)+uint64(r.Size) > math.MaxUint32 {
			slog.Debug("subid: skipping overflow mapping", "start", r.Start, "size", r.Size)
			continue
		}
		maps = append(maps, IDMap{ContainerID: next, HostID: r.Start, Size: r.Size})
		next += r.Size
	}
	return maps
}

// Format renders maps in the /proc/<pid>/uid_map format.
func Format(maps []IDMap) string {
	var b strings.Builder
	for _, m := range maps {
		fmt.Fprintf(&b, "%d %d %d\n", m.ContainerID, m.HostID, m.Size)
	}
	return b.String()
}
