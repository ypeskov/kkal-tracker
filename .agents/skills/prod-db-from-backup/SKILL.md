---
name: prod-db-from-backup
description: Restore a nightly production database backup (latest or a given day of roughly the last week) into the local dev environment, from the server's backup directory or from Google Drive. Use when the user wants production data as of a past night, or the live server is unreachable.
---

# Production Backup → Local

Nightly backups (`kkal_tracker_backup_YYYYMMDD_HHMMSS.db.gz`, UTC time in the name, kept for about a week) exist in two places:
- on the server: `<K8S_REPO_SERVER>/data/backups/` on `SSH_HOST` (both from `.deploy.env`) — the primary source
- on Google Drive: `services/kkal-tracker/backups` — when the server is unreachable

For the current state of production use `prod-db-snapshot` instead.

## Steps
1. Get the backup file into `tmp/`:
   - from the server (latest, or the given day):
     ```bash
     .agents/skills/prod-db-from-backup/scripts/fetch.sh [YYYYMMDD]
     ```
   - or from Google Drive, if a Google Drive connector is available: search for files named
     `kkal_tracker_backup_*.db.gz`, take the latest (or the requested day) and save it to `tmp/`. Some connectors
     return the content base64-encoded: decode it (`base64 -d`) before use
2. Install it locally. The script checks integrity, saves the current `data/app.db` as `data/app.db.bak_YYYYMMDD`,
   replaces it and runs `make migrate-up`:
   ```bash
   scripts/install-local-db.sh tmp/<backup>.db.gz
   ```
3. Delete the downloaded backup from `tmp/`
4. If the dev server is running, restart it: it keeps the old file open (the `dev-server` skill)
5. Report which backup was installed, the number of users and the name of the saved previous database

The backup contains real user data: keep it inside `data/` and `tmp/`, never commit or send it anywhere.
