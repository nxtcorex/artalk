#!/bin/sh
set -eu

database=/data/ip2region.xdb

# Respect existing files, including user-provided symlinks.
if [ ! -e "$database" ] && [ ! -L "$database" ]; then
    if ! mkdir -p /data; then
        echo 'Artalk bundle: cannot create /data for the ip2region database' >&2
        exit 1
    fi
    if ! seed_file=$(mktemp /data/.ip2region.xdb.XXXXXX); then
        echo 'Artalk bundle: cannot create a temporary ip2region database in /data' >&2
        exit 1
    fi
    trap 'rm -f "$seed_file"' EXIT
    trap 'exit 1' HUP INT TERM
    if ! cp /usr/share/artalk/ip2region.xdb "$seed_file" \
        || ! chmod 644 "$seed_file" \
        || ! mv -nT "$seed_file" "$database"; then
        echo 'Artalk bundle: failed to initialize /data/ip2region.xdb' >&2
        exit 1
    fi
    rm -f "$seed_file"
    trap - EXIT HUP INT TERM
fi

exec /entrypoint.sh "$@"
