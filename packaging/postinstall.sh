#!/bin/sh
set -e

if ! id "stapler-builder" >/dev/null 2>&1; then
    useradd \
        --system \
        --home-dir /var/cache/stplr \
        --shell /usr/sbin/nologin \
        --no-create-home \
        stapler-builder
fi

chown -R stapler-builder:stapler-builder /var/cache/stplr
chmod 755 /var/cache/stplr
