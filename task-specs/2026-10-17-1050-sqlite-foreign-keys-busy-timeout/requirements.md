# SQLite: enforce foreign keys and wait on locks (busy timeout)

## Overview
Both items of `TODO.md` in one backend task.

1. **Foreign keys.** The schema declares foreign keys with `ON DELETE CASCADE` / `ON DELETE SET NULL`. SQLite checks
   them only when `PRAGMA foreign_keys = ON` is set on each connection, and no connection sets it. As a result
   nothing is checked and no cascade runs. Deleting a user leaves the user's rows behind, and a row can point to a
   parent that does not exist. Production already has 2 such rows.
2. **Busy timeout.** No connection sets `busy_timeout`. A write that meets a lock held by another connection fails
   at once with `SQLITE_BUSY`, and the client gets a 500. Lock holders include another request, the other pod
   during a rollout, the nightly backup, and `cmd/migrate` run against the live file. With a timeout the write
   waits for the lock instead of failing.

Business value: the declared integrity rules are enforced, and concurrent writes no longer cause random 500 errors.

The user's instruction: do the whole task without creating new tasks. No new `TODO.md` items and no follow-up
proposals. A review finding that this change needs is fixed inside this task.

## Current State

### Connections
- `internal/database/database.go`: `New(path, log)` does `sql.Open("sqlite", path)` and `Ping`. It sets no pragmas
  and no pool limits (database/sql defaults: unlimited open connections).
- `database.New` is the only place that opens SQLite from Go. Its callers:
  - `cmd/web/main.go:31`
  - `cmd/migrate/main.go:30`
  - `cmd/seed/main.go:32`
  - `scripts/create_user.go:39`, which also runs `database.Migrate` on the same connection, line 45
  - the end-to-end test environment, `internal/server/smoke_test.go:200`, which then calls `goose.Up` directly on
    that connection, lines 206-212. Every `internal/server/*_test.go` uses this environment.
- Repository and service unit tests use in-memory fakes, not SQLite.
- The `sqlite3` CLI is used in these places:
  - `make seed-clean` writes (`Makefile:94`)
  - `scripts/install-local-db.sh`, the `prod-db-*` skills and the backup CronJob
    (`kubernetes/base/configmap-backup.yaml`, `sqlite3 .backup`) only read

### Declared foreign keys (8)
| Child table | Column | Parent | ON DELETE | Migration |
|---|---|---|---|---|
| `calorie_entries` | `user_id` | `users` | CASCADE | `20250907200000_initial_schema` |
| `global_ingredient_names` | `ingredient_id` | `global_ingredients` | CASCADE | `20250907200002_add_ingredients_table` |
| `user_ingredients` | `user_id` | `users` | CASCADE | `20250907200002_add_ingredients_table` |
| `user_ingredients` | `global_ingredient_id` (nullable) | `global_ingredients` | SET NULL | `20250907200002_add_ingredients_table` |
| `weight_history` | `user_id` | `users` | CASCADE | `20250907200004_add_user_profile_fields` |
| `activation_tokens` | `user_id` | `users` | CASCADE | `20251031120000_add_user_activation` |
| `api_keys` | `user_id` | `users` | CASCADE | `20260211193105_add_api_keys` |
| `user_preferences` | `user_id` | `users` | CASCADE | `20261017065036_add_user_preferences` |

None of the child tables is itself a parent, so there are no multi-level cascades.

Production, checked read-only on the live file:
- `PRAGMA foreign_key_check` finds exactly 2 orphaned rows, both in `activation_tokens`, referencing deleted
  `users`. Nothing else.
- `journal_mode` is `delete`.
- `integrity_check` is `ok`.

### Code that deletes or re-points parent rows (checked for reliance on enforcement being off)
- **Users.** `UserRepositoryImpl.Delete` (`internal/repositories/user.go:272`) is the only place, and it is called
  only by the registration rollback in `internal/services/auth/service.go`:
  - `:147`: token creation failed, so the user is deleted.
  - `:164-169`: the email failed, so the token is deleted, then the user.

  `CreateWithLanguage` copies the global ingredients into `user_ingredients` when the user is created. With
  enforcement on, the rollback also removes these rows by cascade; today they would be left behind. The explicit
  token deletion before the user deletion becomes redundant but harmless.
- **Global ingredients.** No application code deletes them. Only `make seed-clean` does, in the `sqlite3` CLI, where
  foreign keys are off by default: `DELETE FROM global_ingredients; DELETE FROM global_ingredient_names;`.
  - This leaves `user_ingredients.global_ingredient_id` pointing to deleted rows.
  - `global_ingredients` uses `AUTOINCREMENT`, so the re-seed gets new ids. `cmd/seed/sql/01_ingredients.sql`
    inserts `global_ingredients` without ids but inserts the names with hard-coded `ingredient_id` 1..50. After a
    `seed-clean`, the names therefore point to deleted ids. This is already broken today; with enforcement on, the
    seed fails with a foreign key error.
- **Seed user data.** `cmd/seed/sql/02_user_data.sql` creates `example@example.com` with `INSERT OR IGNORE` and then
  writes `user_ingredients` and `calorie_entries` for the hard-coded `user_id` 1. On a database where id 1 is not
  that user (a production copy, where user 1 is a real account, or a database whose user 1 was deleted), the rows go
  to the wrong user. With enforcement on, they fail if user 1 does not exist.
- **Other writes.** Every other `DELETE` in `internal/repositories/queries.go` targets a leaf table
  (`calorie_entries`, `weight_history`, `user_ingredients`, `activation_tokens`, `api_keys`). No `UPDATE` changes a
  foreign key column. Inserts take the parent id from the authenticated user or from a `SELECT` of existing rows. No
  application path depends on enforcement being off.
- **Migrations.**
  - goose v3.28.0 runs each SQL migration in its own transaction unless the file has `-- +goose NO TRANSACTION`.
  - `PRAGMA foreign_keys` is a no-op inside a transaction, so a migration cannot switch enforcement off for itself.
  - With enforcement on, `DROP TABLE <parent>` performs an implicit `DELETE` that fires the `ON DELETE` actions. The
    Down of `20250907200004_add_user_profile_fields.sql` rebuilds `users` (`DROP TABLE users` + rename). Under
    enforcement it would cascade-delete every user's diary, weight history, ingredients, tokens, API keys and
    preferences.
  - Any future migration that rebuilds a parent table, which is the usual way to change a column in SQLite, has the
    same hazard.
  - SQLite's documented procedure for schema changes: foreign keys off, change, `PRAGMA foreign_key_check`,
    foreign keys back on.

### Concurrency
- Rollback journal (`delete`) and no busy timeout, so a writer that meets a lock gets `SQLITE_BUSY` immediately.
- The Deployment has 1 replica with the default `RollingUpdate` strategy, so the old and the new pod use the file at
  the same time during a rollout.
- Production migrations are applied with `cmd/migrate` against the live file while the app runs.
- Transactions use `db.Begin()` (deferred). The production paths that use transactions write first: user creation
  and meal logging.
- `profile.Service.UpdateProfile` (`internal/services/profile/service.go:98`) opens a transaction but writes through
  the repository on another pool connection (`REFACTORING.md` M4). This constrains pool limits and the transaction
  lock mode, see B7.
- `IngredientRepositoryImpl.CopyGlobalIngredientsToUser` reads, then writes, in one deferred transaction, but nothing
  calls it.

## Requirements

### Backend

**B1. App connections enforce foreign keys and wait on locks.**
- Every connection that `database.New` hands out has `foreign_keys = 1` and `busy_timeout = 5000` (ms). This
  includes every connection the pool opens later, not only the first one.
- Set both through the `_pragma` DSN parameters of `modernc.org/sqlite`: `_pragma=foreign_keys(1)` and
  `_pragma=busy_timeout(5000)`. The driver applies them when it opens each connection.
- Verify the DSN syntax and the per-connection behavior in the v1.60.1 docs or source (AGENTS.md "Libraries" rule).
- A one-time `db.Exec("PRAGMA …")` is not acceptable: it affects only one pooled connection.
- `DATABASE_PATH` stays a plain file path, and its documented format does not change. `internal/database` builds the
  DSN in one place.
- The timeout value is a named constant.

**B2. Every Go entry point gets the right connection.**
- `cmd/web`, `cmd/seed`, the user creation in `scripts/create_user.go` and the test environments use the app
  connection (B1).
- No other `sql.Open("sqlite", …)` exists outside `internal/database`.

**B3. Migrations run with foreign keys off and are checked afterwards.**
- `cmd/migrate` (`up`, `down`, `status`) opens its connection with `busy_timeout = 5000` and with foreign keys
  explicitly off. Rationale: see "Migrations" in Current State. Under enforcement, a parent-table rebuild in a
  migration would silently cascade-delete user data.
- After `up` and after `down`, the tool runs `PRAGMA foreign_key_check` on the same database.
- If the check finds any violation, the command fails with a non-zero exit code. The message lists each violating
  child table with its number of rows and the parent table. It contains no row content.
- A clean check prints nothing extra.
- The same migration path (helper) is used by:
  - `scripts/create_user.go` for its migration step; the user is then created on an app connection
  - the test environment in `internal/server/smoke_test.go`, instead of calling `goose.Up` directly
  - any new test that migrates a database

  The tested path is then the production path.
- The API shape is the developer's choice: a separate constructor or an option, and whether the helper opens its
  own connection or takes the migrations directory as a parameter.

**B4. A cleanup migration removes the existing orphans.**
- A new goose SQL migration with a timestamp after `20261017065036`, for example `cleanup_orphaned_rows`. Create it
  with the goose CLI pinned to the `go.mod` version
  (`go run github.com/pressly/goose/v3/cmd/goose@v3.28.0 -dir migrations create … sql`), as in the previous task.
- Up is a regular transactional migration. Every condition is "the parent row is missing" (`NOT EXISTS`), never
  hard-coded ids, so it is safe on any copy of the database:
  - delete the rows of `calorie_entries`, `weight_history`, `activation_tokens`, `api_keys`, `user_preferences` and
    `user_ingredients` whose `user_id` has no row in `users`
  - delete the rows of `global_ingredient_names` whose `ingredient_id` has no row in `global_ingredients`
  - set `user_ingredients.global_ingredient_id = NULL` where it is not NULL and has no row in `global_ingredients`.
    This mirrors `ON DELETE SET NULL`: the user's ingredient stays.
- It succeeds whatever the connection's foreign key mode is, because deleting child rows and setting a nullable
  reference to NULL never violate a constraint.
- It is idempotent: on a clean database it changes nothing.
- Down has no statements, only a comment that the deleted rows cannot be restored. `migrate down` of this version
  succeeds (verify that goose v3.28.0 accepts an empty Down).
- The migration changes no schema.

**B5. Seeds and the Makefile work under enforcement.**
- `make migrate-up && make seed` on a fresh database succeeds with `cmd/seed` on an app connection (foreign keys
  on).
- `cmd/seed/sql/02_user_data.sql` refers to the example user by email (a subquery on `users.email`) instead of the
  hard-coded `1`, for both `user_ingredients` and `calorie_entries`.
- `make seed-clean`:
  - starts the `sqlite3` script with `PRAGMA foreign_keys = ON;`, so deleting `global_ingredients` cascades to the
    names and sets the users' `global_ingredient_id` to NULL
  - resets the `AUTOINCREMENT` counter of `global_ingredients` (and of `global_ingredient_names`) in
    `sqlite_sequence`, so the re-seeded ids 1..N match the hard-coded `ingredient_id` values of `01_ingredients.sql`

  After `make seed-clean` on a database with users, there are no foreign key violations, and every
  `global_ingredient_names` row belongs to an existing ingredient.
- Update the `README.md` line about `seed-clean` if its described behavior changes.

**B6. The registration rollback is covered.**
- No code change is required.
- A test proves that a registration whose activation email fails leaves no row for that user in `users`,
  `activation_tokens` or `user_ingredients`. The test seeds at least one global ingredient, so ingredients are
  actually copied.

**B7. Pool and transaction constraints.** These are deliberate, not open:
- Do not limit the pool to one connection (`SetMaxOpenConns(1)`). `UpdateProfile` holds a transaction on one
  connection while it writes through another, so a single connection would deadlock.
- Do not switch to `_txlock=immediate`. The same idle transaction would hold the write lock, and its own write
  through the other connection would wait out the busy timeout and fail.
- No WAL: `journal_mode` stays `delete`, and nothing sets or changes it. The reasons are under "WAL decision" below.
- Keep the pool settings unchanged unless verification shows a need. In that case, record the reason in
  `status.md`.

**B8. Errors.**
- A foreign key violation, should one ever happen, is an internal error: it is logged and the client gets the
  existing generic message (AGENTS.md rule).
- No new client-visible error text and no new error mapping are needed. No application path can produce such a
  violation (see Current State).

**B9. Tests.** Use real SQLite databases in `t.TempDir()`, migrated through the B3 helper.
1. **Pragmas.**
   - Hold several connections of the app pool open at the same time (`db.Conn` ×3 or more). Each reports
     `PRAGMA foreign_keys` = 1 and `PRAGMA busy_timeout` = 5000.
   - The migration connection reports `foreign_keys` = 0 and `busy_timeout` = 5000.
2. **Enforcement.** For each of the 8 foreign keys, an `INSERT` with a missing parent fails with a foreign key
   constraint error (table-driven).
3. **Cascade from users.**
   - Deleting a user removes that user's rows from `calorie_entries`, `weight_history`, `activation_tokens`,
     `api_keys`, `user_preferences` and `user_ingredients`.
   - The rows of a second user stay unchanged.
4. **Global ingredients.**
   - Deleting a global ingredient removes its `global_ingredient_names` rows.
   - It sets `user_ingredients.global_ingredient_id` to NULL, and the user ingredient row itself stays.
5. **Busy timeout.**
   - A second, independent `*sql.DB` on the same file (as another process or pod would be) holds a write lock
     (`BEGIN IMMEDIATE` + a write).
   - A write through the app pool starts and the lock is released after about 200 ms.
   - The app write succeeds with no `SQLITE_BUSY`, and it took at least the hold time.
6. **Cleanup migration.**
   - Migrate up to the version before the cleanup (`goose.UpTo`).
   - With foreign keys off, insert valid rows and one orphan for each of the 8 foreign keys. For
     `global_ingredient_id`: a user ingredient of an existing user that points to a missing global ingredient.
   - Run the remaining migrations.
   - Expected: the orphans are gone, that `user_ingredients` row still exists with `global_ingredient_id` NULL, every
     valid row is unchanged, and `PRAGMA foreign_key_check` is empty.
   - Running the migration again on the cleaned data, or on a clean database, changes nothing.
7. **Check after migrations.**
   - A fully migrated database with an orphan inserted afterwards (foreign keys off): the B3 `up` (nothing pending)
     returns an error that names the child table.
   - A clean database: no error.
8. **Registration rollback.** Covered by B6, in the end-to-end environment. For example, point the SMTP port of the
   configuration at a closed port.
9. **Existing tests.** The whole suite stays green: `go vet ./...`, `go test ./...`, and `go test -race` on the
   changed packages. A test that breaks because it inserts rows without parents is fixed by inserting the parents.
   Turning foreign keys off in a test is not a fix.

**B10. Documentation and TODO.**
- `docs/architecture.md` (Database line):
  - app connections enforce foreign keys and wait 5 s on locks (DSN pragmas, every pooled connection)
  - the journal mode stays rollback (`delete`)
  - `cmd/migrate` runs migrations with foreign keys off and fails if `PRAGMA foreign_key_check` finds violations
    afterwards
- `AGENTS.md`, "Migrations" rule: one short addition saying the same for migration authors. A parent-table rebuild
  is safe because the migration runs with foreign keys off. The post-migration check catches the rows the
  migration breaks.
- `docs/infrastructure.md`, Backups: the database stays in rollback-journal mode. WAL is deliberately not used
  because of the backup pod (one sentence, see "WAL decision").
- `README.md`: `seed-clean` line if needed (B5).
- Delete `TODO.md`: both of its items are done and it holds nothing else. Nothing links to it.

### WAL decision (not part of this task)
Keep the rollback journal and only add the busy timeout. With a busy timeout, the reported problem (`SQLITE_BUSY` →
500) is solved at this application's load: one replica and a few users.

WAL would add these risks:
- It is a persistent change of the database file.
- It adds `-wal`/`-shm` files that must be writable by every process that opens the database. That includes the
  backup CronJob pod, a different image and possibly a different user on the same hostPath, as well as the `sqlite3`
  reads over SSH in the `prod-db-*` skills.
- It changes what a plain file copy contains.

The benefit, readers never blocking writers, is not needed now. By the user's instruction, no follow-up task is
created; the choice is documented (B10).

### Frontend
None.

### DevOps
- No manifest or image change.
- The release contains a migration, which is applied manually per the `deploy` skill. See the release note in
  `status.md`.

## Subtasks

### Backend Developer
1. Verify, before coding:
   - in the `modernc.org/sqlite` v1.60.1 docs or source: the `_pragma` DSN syntax with a plain file path, and that it
     is applied to every new connection
   - in goose v3.28.0: one transaction per SQL migration, and an empty Down is accepted

   Record both findings in `status.md`.
2. `internal/database`:
   - B1: app connection DSN and the named timeout constant
   - B3: migration connection (foreign keys off, busy timeout), the migration helpers for up, down and status, and
     the post-migration `foreign_key_check` with a readable, content-free error
3. Switch `cmd/migrate`, `cmd/seed`, `scripts/create_user.go` and the test environment in
   `internal/server/smoke_test.go` to B1/B3 (B2, B3).
4. The cleanup migration (B4).
5. `cmd/seed/sql/02_user_data.sql` and `make seed-clean` (B5). Check by hand:
   - on a fresh database: `make migrate-up && make seed`
   - on a copy of `data/app.db`: `make seed-clean`, then `sqlite3 … 'PRAGMA foreign_key_check;'` is empty

   Do not run these on `data/app.db` itself unless the user's dev data may be re-seeded. Use a copy via
   `DATABASE_PATH`.
6. Tests B9.1-B9.9, including the registration rollback (B6).
7. On a copy of `data/app.db` (a production copy that may contain the orphans): new `cmd/migrate` `up`, then
   `down` (of the cleanup), then `up`. Each step succeeds, and the check is clean after each `up`. Record the
   numbers in `status.md`. Then migrate the local `data/app.db` (`make migrate-up`).
8. Documentation and TODO (B10).
9. `go vet ./...`, `go test ./...` (after `make build-frontend`), `go test -race` on the changed packages.
   `status.md` → `backend-review`.

### Frontend Developer
None.

### DevOps Engineer
Release with the migration, following the release note in `status.md`: backup, then the new `cmd/migrate up`
(cross-built for linux/arm64), then `deploy.sh`. Ask the user for confirmation before touching production.

### Reviews
- Backend review, security review and QA review as usual.
- QA also runs the B9 tests with `-race` and repeats subtask 7 on a fresh copy of `data/app.db`.

## Open Questions

### For the user
None. The cleanup approach (a migration) and the scope come from the user's instruction. The release itself is
confirmed with the user in the DevOps step.

### Technical decisions taken in this analysis (the developer verifies them, and reports a deviation with its reason in `status.md` instead of asking)
- **D1:** `cmd/migrate` runs with foreign keys off and checks `foreign_key_check` afterwards (B3). It does not run
  with foreign keys on.
- **D2:** Orphans are cleaned by a migration, not a one-off script (B4), so every copy of the database is cleaned
  the same way.
- **D3:** No WAL, no pool limit, no `_txlock=immediate` (B7).
- **D4:** The busy timeout is 5000 ms. This is longer than any write path of the app and well below the HTTP
  timeouts.
