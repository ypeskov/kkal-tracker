# Acceptance Criteria: SQLite: enforce foreign keys and wait on locks (busy timeout)

Verified at BA acceptance on 2026-10-17, against the code and the results in `status.md` (developer, reviews, QA).
The unchecked items are the DevOps/production steps.

## Backend

### Connections (B1, B2, B3)
- [x] An app connection from `database.New` reports `PRAGMA foreign_keys` = 1 and `PRAGMA busy_timeout` = 5000.
      This holds on each of at least 3 connections held open at the same time, not only the first one.
      (`TestAppConnectionsEnforceForeignKeysAndWaitOnLocks`)
- [x] The pragmas are set through the `modernc.org/sqlite` `_pragma` DSN parameters, built in `internal/database`
      (`dsn()` in `database.go`). No one-time `db.Exec("PRAGMA …")` is used to configure the pool.
- [x] `DATABASE_PATH` is still a plain file path. Additionally, a path containing `?` and an empty path are rejected.
- [x] The timeout value is a named constant (`database.BusyTimeout`).
- [x] Outside `internal/database`, no Go code calls `sql.Open("sqlite", …)`.
- [x] `cmd/web`, `cmd/seed`, the user creation in `scripts/create_user.go` and the test environments use the app
      connection.
- [x] The `cmd/migrate` connection reports `foreign_keys` = 0 and `busy_timeout` = 5000.
      (`database.NewMigrator`, `TestMigrationConnectionsHaveForeignKeysOff`)
- [x] `scripts/create_user.go` and the test environment (`internal/server/smoke_test.go`) migrate through the same
      helper as `cmd/migrate` (`database.MigrateUp`). No test calls `goose.Up` directly on an app connection.
- [x] After `migrate up` / `migrate down`, `PRAGMA foreign_key_check` runs:
  - [x] Database with an orphan (inserted with foreign keys off after the last migration): `up` returns an error and
        `cmd/migrate` exits non-zero. The message names the child table and the parent table and contains no row
        values. (`TestMigrationsFailOnForeignKeyViolations`; QA: `up`/`down` exit 1 on copies)
  - [x] Clean database: no error, exit code 0.
- [x] Pool and transaction settings are unchanged:
  - no `SetMaxOpenConns(1)`
  - no `_txlock=immediate`
  - nothing sets `journal_mode`: a fresh database created by the app reports `delete`
    (`TestJournalModeStaysDelete`)

### Enforcement (B9.2-B9.4)
- [x] For each of the 8 declared foreign keys, an `INSERT` with a missing parent fails with a foreign key
      constraint error. Examples: `calorie_entries.user_id = 999999`, `global_ingredient_names.ingredient_id =
      999999`, `user_ingredients.global_ingredient_id = 999999`. (`TestForeignKeysAreEnforced`)
- [x] Deleting user A removes all of A's rows from `calorie_entries`, `weight_history`, `activation_tokens`,
      `api_keys`, `user_preferences` and `user_ingredients`. User B's rows in the same tables keep their counts and
      values. (`TestDeletingUserCascadesToTheUsersRows`)
- [x] Deleting global ingredient G (`TestDeletingGlobalIngredientRemovesNamesAndUnlinksUserIngredients`):
  - removes G's `global_ingredient_names` rows
  - sets `global_ingredient_id` to NULL on the `user_ingredients` rows that referenced G
  - keeps those `user_ingredients` rows (same count, name and values)

### Busy timeout (B9.5)
- [x] An independent `*sql.DB` on the same file holds a write lock (`BEGIN IMMEDIATE` + a write) and releases it
      after about 200 ms. A write through the app pool started meanwhile succeeds, with no `SQLITE_BUSY` /
      "database is locked", after waiting at least the hold time.
      (`TestAppWriteWaitsForLockHeldByAnotherConnection`; QA: the migrator waits ~1.75 s on a 2 s lock and fails
      after 5.02 s on a 7 s lock)

### Cleanup migration (B4, B9.6)
- [x] There is a new goose SQL migration in `migrations/`, with a version greater than `20261017065036`, that
      changes no schema: `20261017110038_cleanup_orphaned_rows.sql`.
- [x] Up uses "parent is missing" conditions only, with no hard-coded ids:
  - deletes orphans of `calorie_entries`, `weight_history`, `activation_tokens`, `api_keys`, `user_preferences`
    and `user_ingredients` (by `user_id`)
  - deletes orphans of `global_ingredient_names` (by `ingredient_id`)
  - sets a dangling `user_ingredients.global_ingredient_id` to NULL
- [x] Test: migrate to the version before the cleanup, then insert valid rows plus one orphan per foreign key (8),
      then migrate up. (`TestCleanupMigrationRemovesOrphans`)
  - [x] All 7 orphan rows of the deleting kind are gone.
  - [x] The user ingredient with the dangling `global_ingredient_id` still exists, with NULL there.
  - [x] All valid rows are unchanged.
  - [x] `PRAGMA foreign_key_check` is empty.
- [x] The cleanup statements on an already clean database change 0 rows.
- [x] Down has no statements and a comment that the rows cannot be restored. `migrate down` of this version
      succeeds (`EMPTY`), and `up` succeeds again afterwards.

### Seeds and Makefile (B5)
- [x] Fresh database: `make migrate-up && make seed` succeeds with foreign keys on, and `PRAGMA foreign_key_check`
      is empty afterwards. (developer and QA, scratch databases)
- [x] `cmd/seed/sql/02_user_data.sql` refers to the example user by email. No hard-coded `user_id` 1 is left.
- [x] `make seed-clean` on a copy of a database with users. Accepted deviation: it now runs `cmd/seed -clean` on an
      app connection in one transaction, instead of the `sqlite3` CLI, so that it uses `DATABASE_PATH`. The effect
      B5 requires is unchanged. QA and the developer checked it on a production copy with 4 users:
  - [x] succeeds
  - [x] leaves `PRAGMA foreign_key_check` empty
  - [x] every `global_ingredient_names.ingredient_id` exists in `global_ingredients` (50 ingredients with ids 1..50,
        150 names, 0 without an ingredient)
  - [x] the users' `user_ingredients` rows still exist (363 kept); their `global_ingredient_id` is NULL or valid
- [x] `README.md` is updated: the `seed-clean` line now says the ids start at 1 again and no longer lists `sqlite3`
      as a need.

### Registration rollback (B6)
- [x] End-to-end environment with at least one global ingredient, where the activation email fails (the SMTP
      server is unreachable). Registration returns the existing email-failure error (503), and no row remains for
      that email/user in `users`, `activation_tokens` or `user_ingredients`.
      (`internal/server/registration_rollback_test.go`)

### Errors (B8)
- [x] No new client-visible error text. A foreign key error, if one ever occurs in a handler path, is logged, and
      the client gets the existing generic message. (security review)

### Quality
- [x] `go vet ./...` and `go test ./...` (after `make build-frontend`) are green.
- [x] `go test -race` on `internal/database`, `internal/server` and the other changed packages (`cmd/seed`) is
      green.
- [x] No existing test needed a fix. None was fixed by disabling foreign keys.
- [x] `go.mod` / `go.sum` are unchanged (no new dependency needed).

## Documentation
- [x] `docs/architecture.md` describes:
  - foreign keys on and a 5 s busy timeout for every app connection
  - the rollback journal mode
  - `cmd/migrate` running with foreign keys off, followed by `foreign_key_check`
- [x] The `AGENTS.md` "Migrations" rule has the short note for migration authors (foreign keys off during
      migrations, the check afterwards).
- [x] `docs/infrastructure.md` (Backups) states that the database stays in rollback-journal mode and that WAL is
      deliberately not used.
- [x] `TODO.md` is deleted, and no file references it.
- [x] No new TODO items or follow-up tasks were created.

## Integration
- [x] On a copy of `data/app.db` (production data with the orphaned activation tokens; developer and QA, separately):
  - [x] the new `cmd/migrate up` applies the cleanup migration, and its check is clean
  - [x] `down` (cleanup) → `up` succeeds
  - [x] the number of deleted rows is recorded in `status.md` and matches the orphans that
        `PRAGMA foreign_key_check` reported beforehand: 2 `activation_tokens` rows
- [x] The local `data/app.db` is migrated to `20261017110038` (check empty, integrity ok). The dev server check is
      covered by the end-to-end tests, per QA, because the user did not require a browser check for this
      backend-only change. Every `internal/server` test runs on `database.New` (foreign keys on) and covers login,
      diary create/update/delete and weight create/update/delete. QA also ran the app's statements with foreign keys
      on against a migrated copy of `data/app.db`.
- [ ] Production (DevOps):
  - [ ] the release order is backup → new `cmd/migrate up` (clean check) → release
  - [ ] `PRAGMA foreign_key_check` on the live file is empty afterwards
  - [ ] the pod runs the new version with no errors in the log
  - [ ] the site answers 200, and login plus adding a diary entry work
