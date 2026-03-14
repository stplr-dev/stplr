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

package manager

import "os/exec"

// YUM represents the YUM package manager
type YUM struct {
	commonDNFYUM
}

func NewYUM() *YUM {
	return &YUM{
		commonDNFYUM: commonDNFYUM{
			CommonPackageManager: CommonPackageManager{
				noConfirmArg: "-y",
			},
			binary: "yum",
		},
	}
}

func (*YUM) Exists() bool {
	_, err := exec.LookPath("yum")
	return err == nil
}

func (*YUM) Name() string {
	return "yum"
}

func (*YUM) Format() string {
	return "rpm"
}
