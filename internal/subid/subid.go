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

// Package subid reads id ranges from /etc/subuid, /etc/subgid,
// /proc/<pid>/uid_map and /proc/<pid>/gid_map.
package subid

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"sort"
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

// ParseMap parses a /proc/<pid>/uid_map or /proc/<pid>/gid_map file and
// returns the ranges of ids valid in the namespace it describes (the left
// column).
func ParseMap(r io.Reader) ([]Range, error) {
	var ranges []Range

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 3 {
			slog.Debug("subid: skipping malformed map line", "line", line)
			continue
		}

		first, firstErr := strconv.ParseUint(parts[0], 10, 32)
		size, sizeErr := strconv.ParseUint(parts[2], 10, 32)
		if firstErr != nil || sizeErr != nil || size == 0 {
			slog.Debug("subid: skipping malformed map line", "line", line)
			continue
		}

		ranges = append(ranges, Range{Start: uint32(first), Size: uint32(size)})
	}

	return ranges, scanner.Err()
}

// ValidRanges reads a map file (e.g. /proc/self/uid_map) and returns the ids
// that are valid in the described namespace. A nested namespace can only map
// to these ids.
func ValidRanges(path string) ([]Range, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseMap(f)
}

// ClipTo returns the parts of ranges that lie within valid.
func ClipTo(ranges, valid []Range) []Range {
	var clipped []Range
	for _, r := range ranges {
		for _, v := range valid {
			start := max(uint64(r.Start), uint64(v.Start))
			end := min(r.End(), v.End())
			if start < end {
				//gosec:disable G115 -- id ranges never exceed 32-bit bounds
				clipped = append(clipped, Range{Owner: r.Owner, Start: uint32(start), Size: uint32(end - start)})
			}
		}
	}
	return clipped
}

// ExcludeID removes id from ranges, splitting ranges as needed.
func ExcludeID(ranges []Range, id uint32) []Range {
	var out []Range
	for _, r := range ranges {
		start := uint64(r.Start)
		end := r.End()
		if uint64(id) < start || uint64(id) >= end {
			out = append(out, r)
			continue
		}
		if start < uint64(id) {
			//gosec:disable G115 -- id ranges never exceed 32-bit bounds
			out = append(out, Range{Owner: r.Owner, Start: r.Start, Size: uint32(uint64(id) - start)})
		}
		if uint64(id)+1 < end {
			//gosec:disable G115 -- id ranges never exceed 32-bit bounds
			out = append(out, Range{Owner: r.Owner, Start: id + 1, Size: uint32(end - uint64(id) - 1)})
		}
	}
	return out
}

// CoversAll reports whether ranges cover every id Linux allows (2^32-1 is
// never a valid id).
func CoversAll(ranges []Range) bool {
	sorted := make([]Range, len(ranges))
	copy(sorted, ranges)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Start < sorted[j].Start })

	var next uint64
	for _, r := range sorted {
		if uint64(r.Start) > next {
			return false
		}
		next = max(next, r.End())
	}
	return next >= math.MaxUint32
}

// Subordinate returns configured clipped to valid, excluding id. If no ranges
// remain, synthesize selects valid minus id unless valid covers every id.
// synthesized reports whether the fallback was used.
func Subordinate(id uint32, configured, valid []Range, synthesize bool) (ranges []Range, synthesized bool) {
	ranges = ExcludeID(ClipTo(configured, valid), id)
	if len(ranges) > 0 {
		return ranges, false
	}
	if !synthesize || CoversAll(valid) {
		return nil, false
	}
	if ranges = ExcludeID(valid, id); len(ranges) == 0 {
		return nil, false
	}
	return ranges, true
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
