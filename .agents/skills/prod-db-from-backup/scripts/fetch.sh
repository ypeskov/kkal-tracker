#!/usr/bin/env bash
# Downloads a nightly backup of the production database from the server's backup directory.
# Usage: fetch.sh [YYYYMMDD]   (default: the latest backup). Prints the local path of the .db.gz.
set -euo pipefail

ROOT="$(git rev-parse --show-toplevel)"
source "$ROOT/.deploy.env"
: "${SSH_HOST:?SSH_HOST is not set in .deploy.env}"
: "${K8S_REPO_SERVER:?K8S_REPO_SERVER is not set in .deploy.env}"

DAY="${1:-}"
DIR="${K8S_REPO_SERVER}/data/backups"
NAME=$(ssh "$SSH_HOST" "ls -1 '${DIR}' | grep '^kkal_tracker_backup_${DAY}.*\.db\.gz$' | sort | tail -1")
if [ -z "$NAME" ]; then
    echo "No backup${DAY:+ for $DAY} in ${SSH_HOST}:${DIR}. Available:" >&2
    ssh "$SSH_HOST" "ls -1 '${DIR}'" >&2
    exit 1
fi

mkdir -p "$ROOT/tmp"
scp -q "${SSH_HOST}:${DIR}/${NAME}" "$ROOT/tmp/${NAME}"
echo "$ROOT/tmp/${NAME}"
