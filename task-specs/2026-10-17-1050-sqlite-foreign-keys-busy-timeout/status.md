# Task Status: SQLite: enforce foreign keys and wait on locks (busy timeout)

## Current Phase
**Phase**: `done`
**Assigned to**: `-`
**Updated**: 2026-10-17

Chain (backend-only): BA → Backend Dev → Backend Review → Security Review → QA Review → BA Acceptance → DevOps.

User instruction: "делай и не плоди новые таски" ("do it, and don't create new tasks"). This means:
- no new TODO items and no follow-up tasks
- review findings that this change needs are fixed inside this task
- as few questions to the user as possible

Development starts after the user approves `requirements.md` and `acceptance-criteria.md` (workflow step 2.5).

**Release note for DevOps: the release contains a migration:**
`migrations/20261017110038_cleanup_orphaned_rows.sql`.
- What it does:
  - deletes the rows whose parent is missing; in production that is the 2 orphaned `activation_tokens` rows
    found by the read-only `PRAGMA foreign_key_check`
  - sets a dangling `user_ingredients.global_ingredient_id` to NULL; production has none
- No schema change. Down is a no-op.
- Neither the deploy nor the server applies it. Per the `deploy` skill, stop before the release and agree with the
  user how it is applied, then record the agreement here.
- Order:
  1. backup of the live file with `sqlite3 .backup` on the server, then `integrity_check`
  2. the **new** `cmd/migrate up`, cross-built for linux/arm64 from this release's commit, against
     `<K8S_REPO_SERVER>/data/app.db`. The new tool runs migrations with foreign keys off, waits up to 5 s on locks
     held by the running app, and fails if `PRAGMA foreign_key_check` finds violations afterwards. Expect a clean
     check.
  3. the release with `deploy.sh`
- Compatibility:
  - The running v5.16.0 tolerates the migrated database: no schema change, and the deleted rows are activation
    tokens of users that no longer exist, which no code path can use.
  - The new version also tolerates a database that has not been migrated: SQLite checks foreign keys only on the
    rows a statement changes, so existing orphans cause no errors. The order above is still the one to follow.
  - Rolling back the image to v5.16.0 after the release needs no database change.
- Post-release checks:
  - pod running, no errors in the log
  - site 200
  - login and adding a diary entry work
  - `PRAGMA foreign_key_check` on the live file (read-only) is empty

## Backend Development (2026-10-17)

### Library findings (subtask 1)
- **modernc.org/sqlite v1.60.1** (module cache, `conn.go` `newConn`, `sqlite.go` `applyQueryParams`, `driver.go`
  docs):
  - A plain file name may carry `?` + query parameters. `newConn` cuts the query off a name that does not start with
    `file:` before `sqlite3_open_v2`, so `<path>?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)` opens `<path>`.
    Consequence: a path that itself contains `?` would open another file, so `database.open` rejects such a path.
  - `_pragma` may repeat; each value is executed as `PRAGMA <value>`; `busy_timeout` is sorted first.
  - `Driver.Open` → `newConn` → `applyQueryParams` runs for every new connection, so every pooled connection gets the
    pragmas (verified by the test with 3 connections held open at once).
- **goose v3.28.0** (`migration.go`, `migration_sql.go`, `internal/sqlparser/parser.go`):
  - `ParseSQLMigration` starts with `useTx = true` and sets it to false only on `-- +goose NO TRANSACTION`;
    `runSQLMigration` then runs all statements of the file plus the version row in one `db.BeginTx`.
  - An empty Down (only `-- +goose Down` and comments) parses to zero statements; it runs a transaction with only the
    version row delete and logs `EMPTY`. Verified: `migrate down` of the cleanup prints `EMPTY …` and succeeds.

### Decisions verified
- **D1** holds: `database.Migrator` opens its own pool with `foreign_keys(0)` + `busy_timeout(5000)`; `up`, `upTo`
  and `down` run `PRAGMA foreign_key_check` afterwards. A test with a synthetic parent-table rebuild shows the
  children survive (with the migrator temporarily switched to foreign keys on, the same test loses them).
- **D2** holds: `20261017110038_cleanup_orphaned_rows.sql`, `NOT EXISTS` conditions only.
- **D3** holds: no pool limit, no `_txlock`, no `journal_mode`. Checked the transactions: `CreateWithLanguage` and
  `CreateMeal` write first; `profile.UpdateProfile` runs no statement inside its transaction (a deferred `BEGIN`
  takes no lock), so its write on another connection does not wait on itself. A fresh database reports `delete`.
- **D4** holds: `database.BusyTimeout = 5 * time.Second`. Test: a write under another pool's `BEGIN IMMEDIATE` lock
  waits ~200 ms and succeeds; with the constant temporarily 0 the same test fails with `SQLITE_BUSY`.

### Deviations
- **`make seed-clean` no longer uses the `sqlite3` CLI.** It runs `go run cmd/seed/main.go -clean`: on an app
  connection (foreign keys on), `DELETE FROM global_ingredients` (cascades to the names, sets the users'
  `global_ingredient_id` to NULL), `DELETE FROM global_ingredient_names` (also removes names orphaned before
  enforcement) and `DELETE FROM sqlite_sequence WHERE name IN ('global_ingredients', 'global_ingredient_names')`,
  then the seed files, all in one transaction (review iteration 1). Reason: the old target ran `sqlite3 ./data/app.db` (hard-coded)
  while `cmd/seed` writes to `DATABASE_PATH`, so the required check "`seed-clean` on a copy via `DATABASE_PATH`"
  would have deleted the dev database's ingredients and seeded the copy. Now both steps use the same path and the
  same enforcing connection, and `seed-clean` no longer needs `sqlite3`. The effect required by B5 is unchanged.
- `database.open` rejects a `DATABASE_PATH` containing `?` (see the driver finding) or an empty/blank one (the
  driver cuts the query only at a `?` position >= 1, so with `""` the whole DSN became a file name with no pragmas).
  The documented format (a plain file path) is unchanged.

### API (internal/database)
- `New(path, log)`: app pool, `foreign_keys(1)` + `busy_timeout(5000)`; `BusyTimeout` constant
- `NewMigrator(path, migrationsDir, log)` → `Up`, `UpTo(version)`, `Down`, `Status`, `Close`; `MigrateUp(path, dir,
  log)` opens, migrates up and closes (used by `scripts/create_user.go` and `internal/server` tests);
  `MigrationsDir = "migrations"`; `ErrForeignKeyViolations`. The old `Migrate/MigrateDown/MigrateStatus(db)` are
  removed, so no caller can migrate on an app connection.
- Violation message, e.g.: `foreign key check failed after the migration: global_ingredient_names: 1 row(s)
  reference a missing global_ingredients row; weight_history: 2 row(s) reference a missing users row`.

### Manual checks
- Fresh database (`DATABASE_PATH` in the scratchpad): `make migrate-up && make seed` exit 0, `foreign_key_check`
  empty, `journal_mode` `delete`; 1 user, 86 entries, 50 user ingredients, 50/150 global ingredients/names.
- Fresh database whose user 1 is another account: the seed rows went to `example@example.com` (86 entries, 50
  ingredients), none to user 1.
- Copy of `data/app.db` (production data, version `20261017065036`; backup `tmp/app.db.bak_20261017_fk`,
  `integrity_check` ok). Before: `foreign_key_check` = 2 rows, both `activation_tokens` → `users`; no dangling
  `global_ingredient_id`.
  | Step | Exit | users | calorie_entries | weight_history | activation_tokens | api_keys | user_preferences | user_ingredients | global_ingredients | names | fk_check |
  |---|---|---|---|---|---|---|---|---|---|---|---|
  | before | | 4 | 3615 | 336 | 2 | 1 | 1 | 363 | 50 | 150 | 2 |
  | `up` | 0 | 4 | 3615 | 336 | 0 | 1 | 1 | 363 | 50 | 150 | 0 |
  | `down` (EMPTY) | 0 | 4 | 3615 | 336 | 0 | 1 | 1 | 363 | 50 | 150 | 0 |
  | `up` again | 0 | 4 | 3615 | 336 | 0 | 1 | 1 | 363 | 50 | 150 | 0 |

  Deleted rows: 2, both `activation_tokens`, matching the 2 rows reported beforehand. `integrity_check` ok.
- Copy with 2 orphaned `weight_history` rows and 1 orphaned `global_ingredient_names` row inserted after the last
  migration: `cmd/migrate -cmd=up` (go run and the built binary) exits 1 with the message above; `status` exits 0.
- `make seed-clean` on a migrated copy with 4 users: exit 0; `foreign_key_check` empty; 50 global ingredients with ids
  1..50, 150 names, 0 names without an ingredient; 363 user ingredients kept (313 with `global_ingredient_id` NULL,
  the example user's 50 re-linked by `02_user_data.sql`).
- `scripts/create_user.go` on a seeded scratch database: user created on an app connection, migrations through
  `MigrateUp`, `foreign_key_check` empty.
- Local `data/app.db` migrated with `make migrate-up` to `20261017110038`: `foreign_key_check` empty,
  `integrity_check` ok, `journal_mode` `delete`. The dev server was not started (by instruction), so the
  login/diary/weight check on the dev server is left to QA.

### Tests
- `internal/database/database_test.go` (real SQLite in `t.TempDir()`, migrated through the `Migrator`):
  pragmas on 3 simultaneously held app connections (1/5000) and migration connections (0/5000); journal mode
  `delete`; `?` in the path rejected; 8 foreign keys enforced (table-driven); user delete cascades to 6 tables, the
  other user unchanged; global ingredient delete cascades names and sets NULL, user ingredient kept; busy timeout
  (B9.5); cleanup migration (B9.6: 8 orphans, valid rows unchanged, dangling link NULL, check empty, Down no-op,
  second run changes nothing); check after `up`/`down` (B9.7: error names child and parent, no row values; clean
  database no error); parent-table rebuild keeps the children.
- `internal/server/registration_rollback_test.go` (B6): SMTP on a closed port, one global ingredient; registration
  returns 503; `sqlite_sequence` proves the user, token and ingredient copy were written; no rows remain in `users`,
  `activation_tokens`, `user_ingredients`. With foreign keys temporarily off the test fails (1 user ingredient left).
- `internal/server/smoke_test.go` migrates through `database.MigrateUp`; no test calls `goose.Up`.
- No existing test needed a fix.
- `go vet ./...` ok; `go test ./...` ok; `go test -race -count=1 ./internal/database/ ./internal/server/` ok
  (`cmd/...` and `scripts` have no tests); `go.mod`/`go.sum` unchanged.

## Review Fixes, Iteration 1 (2026-10-17)
Source: `agent-reviews/backend-review.md`. The required item and all four suggestions are done.
- **Required, empty path:** `database.open` rejects an empty or whitespace-only path next to the `?` check.
  `TestUnusableDatabasePathsAreRejected` (table-driven: `?`, `""`, `"  "`) checks `New` and `NewMigrator` and that
  no file is created in the working directory. With the guard disabled, the `""` case fails.
- **Suggestion 1:** `seed -clean` deletes `global_ingredient_names` explicitly after `global_ingredients`. With that
  line removed, `TestSeedCleanResetsGlobalIngredients` (pre-existing orphaned name) fails: check 1 row, 151 names.
- **Suggestion 2:** the clean step and all seed files run in one transaction (`seed` in `cmd/seed/main.go`), so
  `make seed` and `make seed-clean` are all-or-nothing. The files are read before anything is deleted.
  `TestSeedFailureKeepsGlobalIngredients`: clean + a broken file leaves the ingredients, names and links unchanged.
- **Suggestion 3:** a failed seed (clean or file) is reported once, through `log.Fatalf`; the error names the file.
- **Suggestion 4:** `scripts/create_user.go`: all four `log.Fatal` messages have a space after the colon.
- Manual check on a new copy of `data/app.db` (4 users): `make seed-clean` exit 0, `foreign_key_check` empty, 50
  global ingredients with ids 1..50, 150 names, 0 names without an ingredient, 363 user ingredients kept (313 with
  NULL link). On another copy, `cmd/seed -clean` with an extra broken file: one `Seeding failed` line, exit 1, the
  ingredients (50/150) and diary entries (3701) unchanged. `data/app.db` itself was not touched.
- `go vet ./...` ok; `go test ./...` ok; `go test -race -count=1 ./internal/database/ ./internal/server/
  ./cmd/seed/` ok; `go.mod`/`go.sum` unchanged.

## QA Review (2026-10-17)
Clean, no `agent-reviews/qa-review.md`.
- Checks: `go vet ./...` ok; `go test -count=1 ./...` ok; `go test -race -count=1 ./internal/database/ ./cmd/seed/
  ./internal/server/` ok. Coverage: `internal/database` 77.8% (`Status` 0%, only the `cmd/migrate` path), `cmd/seed`
  56.3% (`main` 0%; `seed` 81.2%, `cleanGlobalIngredients` 80%, `executeSQLFile` 100%). No frontend change, so no tsc/lint.
- Mutation checks (on a scratch copy of the repository): the tests fail when the app pool has foreign keys off
  (pragmas, enforcement, registration rollback), when `BusyTimeout` is 0, when the migrator has foreign keys on
  (parent rebuild loses the children), when `checkForeignKeys` is skipped, when the cleanup migration does not set
  NULL or deletes the dangling user ingredient, when `seed -clean` does not reset `sqlite_sequence` or commits the
  clean step separately, and when the `?` guard is removed.
- Subtask 7 repeated on a new copy of the pre-migration backup (`tmp/app.db.bak_20261017_fk`, version
  `20261017065036`, `foreign_key_check` = 2 `activation_tokens` rows): built `cmd/migrate` `up` → exit 0, 2
  `activation_tokens` rows deleted, all other counts unchanged (4/3615/336/1/1/363/50/150), check clean; `down` →
  `EMPTY`, exit 0; `up` again → exit 0; `status` → exit 0; `integrity_check` ok, `journal_mode` `delete`. Same numbers
  as the developer's table.
- `cmd/migrate` exit codes on copies: orphans inserted after the last migration (2 `weight_history`, 1
  `global_ingredient_names`, 1 `api_keys`): `up` → exit 1, message names child and parent tables with counts, no row
  values; `down` → rolls back, then exit 1 with the same message; `status` → exit 0; `up` again (applies the
  cleanup) → exit 0, check clean. `DATABASE_PATH` with `?` → exit 1, no file created; unknown `-cmd` → exit 1.
  Busy timeout of the migrator: another process holding `BEGIN EXCLUSIVE` 2 s → `up` waits ~1.75 s, exit 0; held 7 s
  → `SQLITE_BUSY` after 5.02 s, exit 1.
- Seeds on scratch databases (`DATABASE_PATH`): fresh `make migrate-up && make seed` exit 0, check empty, ids 1..50,
  150 names, `journal_mode` `delete`. `make seed-clean` on a migrated production copy (4 users): exit 0, check empty,
  50 global ingredients with ids 1..50, 150 names, 0 names without an ingredient, 363 user ingredients kept (313
  NULL links, 0 dangling); a second `seed-clean` gives the same. The README lines for `seed` / `seed-clean` are
  correct. Pre-existing, unchanged by this task: every `seed` / `seed-clean` run appends the example user's 86 diary
  entries again, and a second plain `make seed` adds 50 nameless global ingredients (ids 51..100) that the next
  `seed-clean` removes.
- App behavior with foreign keys on: every `internal/server` test runs on `database.New` (via
  `newSmokeEnvironmentWith`) and covers login, diary create/update/delete, weight create/update/delete and ingredient
  create/update/delete. In addition, on a copy of the migrated `data/app.db` (`sqlite3` with `foreign_keys = ON`,
  `busy_timeout = 5000`), the repository's statements for the most active real user (diary entry, weight entry,
  custom and linked ingredient: insert, update, delete) succeed, and an insert with a missing user is rejected.
  `data/app.db` itself was not touched. The browser check on the dev server (login, diary entry, weight entry) is
  left to the user/BA acceptance, since no server was started.

## Phase History
| Date       | Phase       | Agent            | Notes |
|------------|-------------|------------------|-------|
| 2026-10-17 | analysis    | business-analyst | Specs written. Current state: no connection sets `foreign_keys` or `busy_timeout` (`internal/database/database.go`); 8 foreign keys declared; production has 2 orphaned `activation_tokens` rows. No application path relies on enforcement being off. What breaks under enforcement: `make seed-clean` (AUTOINCREMENT ids vs hard-coded `ingredient_id`), `02_user_data.sql` (hard-coded `user_id` 1), migrations run on an enforcing connection (the Down of `20250907200004` drops `users` → cascade). Decisions D1-D4 in `requirements.md`; no questions for the user |
| 2026-10-17 | backend-dev | business-analyst | First phase set; waiting for the user's approval of the specs |
| 2026-10-17 | backend-review | backend-developer | Implemented B1-B10: DSN pragmas, `database.Migrator` with `foreign_key_check`, cleanup migration `20261017110038`, seeds by email, `seed-clean` via `cmd/seed -clean` (deviation, see above), tests, docs, `TODO.md` deleted. Dev DB migrated (backup `tmp/app.db.bak_20261017_fk`) |
| 2026-10-17 | backend-dev | backend-code-reviewer | Review iteration 1: no critical issues; 1 improvement required (empty `DATABASE_PATH` bypasses the DSN pragmas, reject it in `database.open`), 4 suggestions; see `agent-reviews/backend-review.md`. vet/build/test/-race green |
| 2026-10-17 | backend-review | backend-developer | Review iteration 1 fixed: empty/blank `DATABASE_PATH` rejected (+ test); seed `-clean`: explicit names delete, clean + seed files in one transaction, single error report (+ `cmd/seed/main_test.go`); `create_user.go` messages. `seed-clean` re-verified on a copy |
| 2026-10-17 | security-review | backend-code-reviewer | Review iteration 2: clean. Empty/blank path guard, seed clean + files in one transaction (explicit names delete, sequence reset in-transaction verified by tests), single error report, create_user messages. vet/build/test green, -race on internal/database and cmd/seed green; report deleted |
| 2026-10-17 | qa-review | security-reviewer | Security review iteration 1: clean, no report. Data loss: the only parent delete in app code is the registration rollback of the just-created user (cascade now also removes its copied `user_ingredients`, as intended); no `OR REPLACE` can delete a parent (`01_ingredients.sql` inserts `global_ingredients` without id and the table has no unique key, so REPLACE never fires a cascade); migrations run with foreign keys off, so parent rebuilds in Up/Down cannot cascade; cleanup migration uses `NOT EXISTS` on NOT NULL columns only (matches `foreign_key_check` exactly), empty Down; `seed -clean` is not in the image, one transaction, only NULLs `global_ingredient_id`. `foreign_key_check` error has table names and counts only (test asserts no token value or id). `DATABASE_PATH`: `?` and empty/blank rejected, no DSN parameter injection (driver v1.60.1 `newConn` checked; a `file:` prefix still gets the pragmas). FK errors reach clients only as existing generic 500s (no string-based error mapping). No secrets in the diff; `go.mod`/`go.sum` unchanged, no dependency audit needed |
| 2026-10-17 | ba-acceptance | qa-engineer | QA review iteration 1: clean, no report. vet/test/-race green; all acceptance criteria mapped to tests; mutation checks confirm the key tests fail on regressions; subtask 7 repeated on a new pre-migration copy (2 `activation_tokens` deleted, down/up clean); `cmd/migrate` exit codes and busy wait checked on copies; seeds re-checked; dev-server browser check left manual (see QA Review) |
| 2026-10-17 | devops | business-analyst | Acceptance passed. Every criterion was checked against the code (`internal/database/database.go`, `migrate.go`, the cleanup migration, `02_user_data.sql`, `Makefile`, callers, docs, `TODO.md` deleted) and the developer's and QA's results. Accepted deviation: `seed-clean` runs `cmd/seed -clean` (same effect, uses `DATABASE_PATH`). The dev-server criterion is covered by the end-to-end tests with foreign keys on, per QA; the user did not require a browser check. Unchecked: the production steps (backup → new `cmd/migrate up` → release, post-release checks) |
| 2026-10-17 | done | devops-engineer | Production DB backed up (`app.db.bak_20261017_114935`, integrity ok); `foreign_key_check` before: 2 rows (activation_tokens → users); new `cmd/migrate up` (linux/arm64) applied `20261017110038_cleanup_orphaned_rows.sql`, exit 0, check clean, integrity ok, journal mode delete. Released v5.17.0 with `deploy.sh`: pod running on `ypeskov/kcal-tracker:5.17.0`, 0 restarts, no errors in the log; site 200; login endpoint answers; `foreign_key_check` on the live file (read-only) empty |

## Review Iterations
- **Iteration 1 (backend review)**: 2026-10-17 — issues: empty-path guard in `database.open` (required); suggestions for `seed -clean` (explicit names delete, atomicity, double logging) and `scripts/create_user.go` messages. Report: `agent-reviews/backend-review.md`
- **Iteration 2 (backend review)**: 2026-10-17 — clean; `agent-reviews/backend-review.md` deleted
- **Iteration 1 (security review)**: 2026-10-17 — clean; no `agent-reviews/security-review.md`
- **Iteration 1 (QA review)**: 2026-10-17 — clean; no `agent-reviews/qa-review.md`
- **BA acceptance**: 2026-10-17 — passed; status → `devops`. Open: the production steps only
