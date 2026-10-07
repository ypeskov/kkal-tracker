---
name: prod-db-snapshot
description: Copy the live production database from the server into the local dev environment (consistent sqlite3 snapshot over SSH, read-only for prod). Use when the user wants fresh production data locally, e.g. to reproduce a bug.
---

# Live Production Database → Local

Takes a consistent snapshot of the current production database on the server and installs it as `data/app.db`.
For an older state, or when the server is unreachable, use `prod-db-from-backup` instead.

The production database is `<K8S_REPO_SERVER>/data/app.db` on `SSH_HOST` (both from `.deploy.env`), a hostPath volume
of the pod. Never write to it, never copy it with a plain `cp`/`scp` (inconsistent while the app writes).

## Steps
1. Download a snapshot (prints the local path, `tmp/kkal_snapshot_*.db.gz`):
   ```bash
   .agents/skills/prod-db-snapshot/scripts/fetch.sh
   ```
2. Install it locally. The script checks integrity, saves the current `data/app.db` as `data/app.db.bak_YYYYMMDD`,
   replaces it and runs `make migrate-up`:
   ```bash
   scripts/install-local-db.sh <path from step 1>
   ```
3. Delete the downloaded snapshot from `tmp/`
4. If the dev server is running, restart it: it keeps the old file open (the `dev-server` skill)
5. Report the number of users and the backup file name printed by the script

The snapshot contains real user data: keep it inside `data/` and `tmp/`, never commit or send it anywhere.
