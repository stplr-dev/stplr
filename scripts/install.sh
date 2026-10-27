#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-or-later
#
# This file was originally part of the project "LURE - Linux User REpository",
# created by Elara Musayelyan.
# It was later modified as part of "ALR - Any Linux Repository" by the ALR Authors.
# This version has been further modified as part of "Stapler" by Maxim Slipenko and other Stapler Authors.
#
# Copyright (C) Elara Musayelyan (LURE)
# Copyright (C) 2025 The ALR Authors
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

UNAME_M=$(uname -m)
UNAME_S=$(uname -s)

if [ "$UNAME_S" != "Linux" ]; then
    echo "Operation system $UNAME_S is not supported!"
    exit 1
fi

case "$UNAME_M" in
    arm64) ARCH="linux-aarch64" ;;
    *) ARCH="linux-$UNAME_M" ;;
esac

REPO_URL="https://altlinux.space/stapler/stplr"
LATEST_RELEASE_API="https://altlinux.space/api/v1/repos/stapler/stplr/releases/latest"

get_latest_version() {
    curl -fsSL "$LATEST_RELEASE_API" \
        | grep -o '"tag_name":"v[^"]*"' \
        | head -n1 \
        | sed -E 's/.*"v([^"]+)"/\1/'
}

VERSION="${STPLR_VERSION:-}"
if [ -z "$VERSION" ]; then
    VERSION="$(get_latest_version)" || true
    if [ -z "$VERSION" ]; then
        echo "❌ Error: Failed to determine the latest stplr version."
        echo "Set STPLR_VERSION to install a specific version."
        exit 1
    fi
fi
VERSION="${VERSION#v}"

BASE_URL="$REPO_URL/releases/download/v$VERSION"
TAR_NAME="stplr-$VERSION-$ARCH.tar.gz"

PREFIX="/usr/local"
BIN_DIR="$PREFIX/bin"
BASH_COMPLETION_DIR="$PREFIX/share/bash-completion/completions"
ZSH_COMPLETION_DIR="$PREFIX/share/zsh/site-functions"
STPLR_BIN="$BIN_DIR/stplr"

SYSTEM_USER="stapler-builder"
ROOT_DIRS=("/var/cache/stplr" "/etc/stplr")

run_as_root() {
    if [ "$EUID" -eq 0 ]; then
        # Already root, run directly
        "$@"
    elif command -v pkexec >/dev/null 2>&1; then
        echo "🔐 Running with pkexec: $*"
        pkexec "$@"
    elif command -v sudo >/dev/null 2>&1; then
        echo "🔐 Running with sudo: $*"
        sudo "$@"
    else
        echo "❌ Error: Root privileges required but neither pkexec nor sudo found."
        echo "Please run as root or install sudo/pkexec."
        exit 1
    fi
}

create_root_script() {
    local script_content="$1"
    local temp_script=$(mktemp)
    
    cat > "$temp_script" << 'EOF'
#!/bin/bash
set -euo pipefail
EOF
    
    echo "$script_content" >> "$temp_script"
    chmod +x "$temp_script"
    echo "$temp_script"
}

echo "📦 Installing stplr v$VERSION..."

TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

echo "⬇️  Downloading stplr v$VERSION..."
HTTP_CODE=$(curl -sSL -w '%{http_code}' "$BASE_URL/$TAR_NAME" -o "$TMP_DIR/stplr.tar.gz")
case "$HTTP_CODE" in
    200) ;;
    404)
        echo "❌ Error: stplr v$VERSION is not available for $UNAME_M."
        echo "See $REPO_URL/releases for supported architectures."
        exit 1
        ;;
    *)
        echo "❌ Error: Failed to download $TAR_NAME (HTTP $HTTP_CODE)."
        exit 1
        ;;
esac
tar -xzf "$TMP_DIR/stplr.tar.gz" -C "$TMP_DIR"

ROOT_SCRIPT_CONTENT="
# Install binary
echo '📁 Installing binary...'
install -Dm755 '$TMP_DIR/stplr' '$STPLR_BIN'
echo '✅ Binary installed to $STPLR_BIN'

# Install shell completions if available
if [ -d '$TMP_DIR/completion' ]; then
    if [ -f '$TMP_DIR/completion/stplr' ]; then
        mkdir -p '$BASH_COMPLETION_DIR'
        install -Dm644 '$TMP_DIR/completion/stplr' '$BASH_COMPLETION_DIR/stplr'
        echo '🔧 Bash completion installed'
    fi
    if [ -f '$TMP_DIR/completion/_stplr' ]; then
        mkdir -p '$ZSH_COMPLETION_DIR'
        install -Dm644 '$TMP_DIR/completion/_stplr' '$ZSH_COMPLETION_DIR/_stplr'
        echo '🔧 Zsh completion installed'
    fi
fi

if [ "$VERSION" = "0.0.26" ]; then
    if ! command -v setcap >/dev/null 2>&1; then
        echo '❌ Error: setcap command not found. Please install libcap2-bin package.'
        exit 1
    fi

    echo '🔐 Setting capabilities...'
    if ! setcap cap_setuid,cap_setgid+ep '$STPLR_BIN'; then
        echo '❌ Error: Failed to set capabilities on $STPLR_BIN'
        echo 'This is required for stplr v0.0.26 to function properly.'
        exit 1
    fi
    echo '✅ Capabilities set successfully'
fi

# Create system user
if ! id '$SYSTEM_USER' >/dev/null 2>&1; then
    echo '👤 Creating system user \"$SYSTEM_USER\"...'
    useradd -r -s /usr/sbin/nologin '$SYSTEM_USER'
    echo '✅ System user created'
else
    echo '👤 System user \"$SYSTEM_USER\" already exists'
fi

# Create required directories
echo '📁 Creating required directories...'
for dir in ${ROOT_DIRS[@]}; do
    install -d -o '$SYSTEM_USER' -g '$SYSTEM_USER' -m 755 \"\$dir\"
    echo \"✅ Created \$dir owned by $SYSTEM_USER\"
done
"

ROOT_SCRIPT=$(create_root_script "$ROOT_SCRIPT_CONTENT")
run_as_root bash "$ROOT_SCRIPT"
rm -f "$ROOT_SCRIPT"

echo ""
echo "🎉 stplr v$VERSION installed successfully!"
echo "👉 Run 'stplr --help' to get started."