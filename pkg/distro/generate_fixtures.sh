#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-or-later
#
# Stapler
# Copyright (C) 2025 The Stapler Authors
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

set -e

IMAGES=(
  "debian:12"
  "debian:13"
  "ubuntu:22.04"
  "ubuntu:24.04"
  "fedora:41"
  "fedora:42"
  "registry.altlinux.org/sisyphus/alt:latest"
  "registry.altlinux.org/p11/alt:latest"
)

mkdir -p tests-fixtures

for img in "${IMAGES[@]}"; do
  fname=$(echo "$img" | tr ':/' '-')
  podman run --rm "$img" cat /etc/os-release > "tests-fixtures/$fname"
done