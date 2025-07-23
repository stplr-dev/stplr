// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file was originally part of the project "LURE - Linux User REpository",
// created by Elara Musayelyan.
// It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
// This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
//
// Copyright (C) Elara Musayelyan (LURE)
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

package cpu

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/sys/cpu"
)

// armVariant checks which variant of ARM stplr is running
// on, by using the same detection method as Go itself
func armVariant() string {
	armEnv := os.Getenv("STPLR_ARM_VARIANT")
	// ensure value has "arm" prefix, such as arm5 or arm6
	if strings.HasPrefix(armEnv, "arm") {
		return armEnv
	}

	switch {
	case cpu.ARM.HasVFPv3:
		return "arm7"
	case cpu.ARM.HasVFP:
		return "arm6"
	default:
		return "arm5"
	}
}

// Arch returns the canonical CPU architecture of the system
func Arch() string {
	arch := os.Getenv("STPLR_ARCH")
	if arch == "" {
		arch = runtime.GOARCH
	}
	if arch == "arm" {
		arch = armVariant()
	}
	return arch
}

func isCompatibleARM(target, arch string) bool {
	if !strings.HasPrefix(target, "arm") || !strings.HasPrefix(arch, "arm") {
		return false
	}

	targetVer, err1 := getARMVersion(target)
	archVer, err2 := getARMVersion(arch)
	if err1 != nil || err2 != nil {
		return false
	}

	return targetVer >= archVer
}

func IsCompatibleWith(target string, list []string) bool {
	if target == "all" || slices.Contains(list, "all") {
		return true
	}

	for _, arch := range list {
		if target == arch || isCompatibleARM(target, arch) {
			return true
		}
	}

	return false
}

// CompatibleArches returns a slice of compatible architectures for the given processor architecture
func CompatibleArches(arch string) ([]string, error) {
	if strings.HasPrefix(arch, "arm") {
		ver, err := getARMVersion(arch)
		if err != nil {
			return nil, err
		}

		if ver > 5 {
			var out []string
			for i := ver; i >= 5; i-- {
				out = append(out, "arm"+strconv.Itoa(i))
			}
			return out, nil
		}
	}

	return []string{arch}, nil
}

func getARMVersion(arch string) (int, error) {
	// Extract the version number from ARM architecture
	version := strings.TrimPrefix(arch, "arm")
	if version == "" {
		return 5, nil // Default to arm5 if version is not specified
	}
	return strconv.Atoi(version)
}
