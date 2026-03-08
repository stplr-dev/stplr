#!/bin/bash
# SPDX-License-Identifier: GPL-3.0-or-later
#
# Stapler
# Copyright (C) 2026 The Stapler Authors
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


# create-fixtures.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="$SCRIPT_DIR/fixtures"

echo "Building Docker image..."
docker build -t archive-generator .

echo "Creating fixtures directory..."
rm -rf "$FIXTURES_DIR"
mkdir -p "$FIXTURES_DIR"

echo "Exporting fixtures..."
CONTAINER_ID=$(docker create archive-generator true)
docker cp "$CONTAINER_ID:/fixtures/." "$FIXTURES_DIR/"
docker rm "$CONTAINER_ID"

echo "Done! Fixtures are in $FIXTURES_DIR"
ls -lh "$FIXTURES_DIR"