#!/usr/bin/env bash
# Installs a copy of the production database as the local data/app.db.
# Usage: scripts/install-local-db.sh <backup.db.gz | backup.db>
# The current data/app.db is kept as data/app.db.bak_YYYYMMDD[_HHMMSS]; pending migrations are applied afterwards.
set -euo pipefail

SRC="${1:-}"
if [ -z "$SRC" ] || [ ! -f "$SRC" ]; then
    echo "Usage: $(basename "$0") <backup.db.gz | backup.db>" >&2
    exit 1
fi

cd "$(dirname "$0")/.."
DB="data/app.db"
TMP="data/app.db.incoming"
mkdir -p data

case "$SRC" in
    *.gz) gunzip -c "$SRC" > "$TMP" ;;
    *)    cp "$SRC" "$TMP" ;;
esac

if [ "$(sqlite3 "$TMP" 'PRAGMA integrity_check;')" != "ok" ]; then
    rm -f "$TMP"
    echo "Integrity check failed for $SRC, local database left untouched" >&2
    exit 1
fi
USERS=$(sqlite3 "$TMP" 'SELECT count(*) FROM users;')

if [ -f "$DB" ]; then
    BAK="data/app.db.bak_$(date +%Y%m%d)"
    [ -e "$BAK" ] && BAK="${BAK}_$(date +%H%M%S)"
    cp "$DB" "$BAK"
    echo "Previous database saved as $BAK"
fi

# Journal files of the old database must not be applied to the new one
rm -f "${DB}-wal" "${DB}-shm" "${DB}-journal"
mv "$TMP" "$DB"
echo "Installed $SRC as $DB ($USERS users)"

make migrate-up
