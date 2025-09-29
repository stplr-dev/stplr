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

set -euo pipefail

NAME=${NAME:-stplr}
OUT=./out

declare -A DISTROS=(
  [deb]="debian"
  [rpm]="fedora altlinux"
  [apk]="alpine"
  [archlinux]="arch"
)

rm -rf "$OUT"
mkdir -p "$OUT"
cp -r ./packaging "$OUT"
cp "./$NAME" "$OUT/packaging/$NAME"

for fmt in "${!DISTROS[@]}"; do
  for distro in ${DISTROS[$fmt]}; do
    echo "===> Building $fmt for $distro..."
    (
      cd "$OUT"
      STPLR_PKG_FORMAT=$fmt STPLR_DISTRO=$distro \
        ../$NAME -i=false build -c -s ./packaging/Staplerfile --no-suffix
    )
  done
done