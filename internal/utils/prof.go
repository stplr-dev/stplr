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

package utils

import (
	"runtime"
	"runtime/debug"
)

func ForceGC() {
	runtime.GC()
	debug.FreeOSMemory()
}

/*
func WriteHeapProfile(path string) error {
	runtime.GC()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return pprof.Lookup("heap").WriteTo(f, 2)
}

func writeProfile(path, name string, debug int) error {
	p := pprof.Lookup(name)
	if p == nil {
		return fmt.Errorf("pprof profile %q not found", name)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return p.WriteTo(f, debug)
}

func WriteProf(dir, prefix string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	ts := time.Now().Format("20060102_150405")

	// HEAP:
	runtime.GC()
	if err := writeProfile(
		filepath.Join(dir, fmt.Sprintf("%s_heap_%s.pprof", prefix, ts)),
		"heap", 0,
	); err != nil {
		return err
	}

	// ALLOCS
	if err := writeProfile(
		filepath.Join(dir, fmt.Sprintf("%s_allocs_%s.pprof", prefix, ts)),
		"allocs", 0,
	); err != nil {
		return err
	}

	// goroutines
	if err := writeProfile(
		filepath.Join(dir, fmt.Sprintf("%s_goroutine_%s.pprof", prefix, ts)),
		"goroutine", 0,
	); err != nil {
		return err
	}

	return nil
}
*/
