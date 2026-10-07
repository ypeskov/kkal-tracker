#!/usr/bin/env bash
# Takes a consistent snapshot of the live production database on the server and downloads it.
# Prints the local path of the downloaded .db.gz. Read-only for production.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
source "$ROOT/.deploy.env"
: "${SSH_HOST:?SSH_HOST is not set in .deploy.env}"
: "${K8S_REPO_SERVER:?K8S_REPO_SERVER is not set in .deploy.env}"

STAMP="$(date +%Y%m%d_%H%M%S)"
REMOTE="/tmp/kkal_snapshot_${STAMP}.db"
LOCAL="$ROOT/tmp/kkal_snapshot_${STAMP}.db.gz"
mkdir -p "$ROOT/tmp"

# sqlite3 .backup is safe while the application writes; a plain file copy is not
ssh "$SSH_HOST" "sqlite3 '${K8S_REPO_SERVER}/data/app.db' \".backup '${REMOTE}'\" && gzip -f '${REMOTE}'" >&2
scp -q "${SSH_HOST}:${REMOTE}.gz" "$LOCAL"
ssh "$SSH_HOST" "rm -f '${REMOTE}.gz'"

echo "$LOCAL"
