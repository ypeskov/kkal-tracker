---
name: restore-prod-db
description: Restore the latest production SQLite database backup to the local development environment
disable-model-invocation: true
---

# Restore Production DB to Local Dev

Restore the latest production SQLite database backup to the local development environment.

## Prerequisites

- **Stop the development server** (Air / `make dev`) before running — SQLite locks the database file and replacing it under a running process will cause errors.

## Steps

1. **Find the latest backup**
   - Go to `/Users/ypeskov/Google Drive/My Drive/services/kkal-tracker/backups/`
   - Find the most recent `.db.gz` file (format: `kkal_tracker_backup_YYYYMMDD_HHMMSS.db.gz`). Sort by modification date (`ls -t`), not by name

2. **Copy to project data directory**
   - Copy the backup file to `data/` in the project root:
     ```bash
     cp "/Users/ypeskov/Google Drive/My Drive/services/kkal-tracker/backups/<latest>.db.gz" data/
     ```

3. **Remove the current local database**
   - ```bash
     rm -f data/app.db
     ```

4. **Extract the backup**
   - ```bash
     gunzip data/<latest>.db.gz
     ```
   - This produces `data/kkal_tracker_backup_YYYYMMDD_HHMMSS.db`

5. **Rename to app.db**
   - ```bash
     mv data/kkal_tracker_backup_*.db data/app.db
     ```

6. **Verify**
   - Run: `sqlite3 data/app.db "SELECT count(*) FROM users;"`
   - Confirm the query returns a number (database is valid and has data)

7. **Run pending migrations**
   - Local dev may have newer migrations than production:
     ```bash
     make migrate-up
     ```

## Configuration

| Parameter | Value |
|-----------|-------|
| Backup location | `/Users/ypeskov/Google Drive/My Drive/services/kkal-tracker/backups/` |
| Backup format | `kkal_tracker_backup_YYYYMMDD_HHMMSS.db.gz` (gzipped SQLite) |
| Local DB path | `data/app.db` |
| Database type | SQLite |
