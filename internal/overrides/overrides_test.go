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

package overrides_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"go.stplr.dev/stplr/internal/overrides"
	"go.stplr.dev/stplr/pkg/distro"
)

var info = &distro.OSRelease{
	ID:        "centos",
	Like:      []string{"rhel", "fedora"},
	ReleaseID: "9",
}

func TestResolve(t *testing.T) {
	names, err := overrides.Resolve(info, nil)
	assert.NoError(t, err)

	expected := []string{
		"amd64_centos_9_en",
		"centos_9_en",
		"amd64_rhel_9_en",
		"rhel_9_en",
		"amd64_fedora_9_en",
		"fedora_9_en",
		"amd64_centos_en",
		"centos_en",
		"amd64_rhel_en",
		"rhel_en",
		"amd64_fedora_en",
		"fedora_en",
		"amd64_en",
		"en",
		"amd64_centos_9",
		"centos_9",
		"amd64_rhel_9",
		"rhel_9",
		"amd64_fedora_9",
		"fedora_9",
		"amd64_centos",
		"centos",
		"amd64_rhel",
		"rhel",
		"amd64_fedora",
		"fedora",
		"amd64",
		"",
	}

	assert.Equal(t, expected, names)
}

func TestResolveName(t *testing.T) {
	names, err := overrides.Resolve(info, &overrides.Opts{
		Name:        "deps",
		Overrides:   true,
		LikeDistros: true,
	})
	assert.NoError(t, err)

	expected := []string{
		"deps_amd64_centos_9",
		"deps_centos_9",
		"deps_amd64_rhel_9",
		"deps_rhel_9",
		"deps_amd64_fedora_9",
		"deps_fedora_9",
		"deps_amd64_centos",
		"deps_centos",
		"deps_amd64_rhel",
		"deps_rhel",
		"deps_amd64_fedora",
		"deps_fedora",
		"deps_amd64",
		"deps",
	}

	assert.Equal(t, expected, names)
}

func TestResolveArch(t *testing.T) {
	os.Setenv("STPLR_ARCH", "arm7")
	defer os.Setenv("STPLR_ARCH", "")

	names, err := overrides.Resolve(info, &overrides.Opts{
		Name:        "deps",
		Overrides:   true,
		LikeDistros: true,
	})
	assert.NoError(t, err)

	expected := []string{
		"deps_arm7_centos_9",
		"deps_arm6_centos_9",
		"deps_arm5_centos_9",
		"deps_centos_9",
		"deps_arm7_rhel_9",
		"deps_arm6_rhel_9",
		"deps_arm5_rhel_9",
		"deps_rhel_9",
		"deps_arm7_fedora_9",
		"deps_arm6_fedora_9",
		"deps_arm5_fedora_9",
		"deps_fedora_9",
		"deps_arm7_centos",
		"deps_arm6_centos",
		"deps_arm5_centos",
		"deps_centos",
		"deps_arm7_rhel",
		"deps_arm6_rhel",
		"deps_arm5_rhel",
		"deps_rhel",
		"deps_arm7_fedora",
		"deps_arm6_fedora",
		"deps_arm5_fedora",
		"deps_fedora",
		"deps_arm7",
		"deps_arm6",
		"deps_arm5",
		"deps",
	}

	assert.Equal(t, expected, names)
}

func TestResolveNoLikeDistros(t *testing.T) {
	names, err := overrides.Resolve(info, &overrides.Opts{
		Overrides:   true,
		LikeDistros: false,
	})
	assert.NoError(t, err)

	expected := []string{
		"amd64_centos_9",
		"centos_9",
		"amd64_centos",
		"centos",
		"amd64",
		"",
	}

	assert.Equal(t, names, expected)
}

func TestResolveNoOverrides(t *testing.T) {
	names, err := overrides.Resolve(info, &overrides.Opts{
		Name:        "deps",
		Overrides:   false,
		LikeDistros: false,
	})
	assert.NoError(t, err)

	expected := []string{"deps"}

	assert.Equal(t, expected, names)
}

func TestResolveLangs(t *testing.T) {
	names, err := overrides.Resolve(info, &overrides.Opts{
		Overrides:    true,
		Languages:    []string{"ru_RU", "en", "en_US"},
		LanguageTags: []language.Tag{language.BritishEnglish},
	})
	assert.NoError(t, err)

	expected := []string{
		"amd64_centos_9_en",
		"centos_9_en",
		"amd64_centos_en",
		"centos_en",
		"amd64_en",
		"en",
		"amd64_centos_9_ru",
		"centos_9_ru",
		"amd64_centos_ru",
		"centos_ru",
		"amd64_ru",
		"ru",
		"amd64_centos_9",
		"centos_9",
		"amd64_centos",
		"centos",
		"amd64",
		"",
	}

	assert.Equal(t, expected, names)
}

func TestReleasePlatformSpecific(t *testing.T) {
	type testCase struct {
		info     *distro.OSRelease
		expected string
	}

	for _, tc := range []testCase{
		{
			info: &distro.OSRelease{
				ID:         "centos",
				Like:       []string{"rhel", "fedora"},
				PlatformID: "platform:el8",
			},
			expected: "1.el8",
		},
		{
			info: &distro.OSRelease{
				ID:         "fedora",
				PlatformID: "platform:f42",
			},
			expected: "1.f42",
		},
		{
			info: &distro.OSRelease{
				ID: "altlinux",
			},
			expected: "alt1",
		},
		{
			info: &distro.OSRelease{
				ID:   "alt-atomic",
				Like: []string{"altlinux"},
			},
			expected: "alt1",
		},
		{
			info: &distro.OSRelease{
				ID: "ubuntu",
			},
			expected: "1",
		},
	} {
		assert.Equal(t, tc.expected, overrides.ReleasePlatformSpecific(1, tc.info))
		release, err := overrides.ParseReleasePlatformSpecific(tc.expected, tc.info)
		assert.NoError(t, err)
		assert.Equal(t, 1, release)
	}
}
