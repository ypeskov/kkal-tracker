# TODO

## Enable SQLite foreign keys
The migrations declare foreign keys with `ON DELETE CASCADE` / `ON DELETE SET NULL` (users, user ingredients,
activation tokens, API keys, weight history, ...), but SQLite enforces them only when `PRAGMA foreign_keys = ON` is
set on every connection, and `internal/database/database.go` opens the database without it. So no foreign key is
checked and no cascade runs: deleting a user leaves its rows behind.

- Enable it for every pooled connection: with `modernc.org/sqlite`, the DSN parameter `?_pragma=foreign_keys(1)`
  (the same for `cmd/migrate` and the tests)
- Before enabling it in production: run `PRAGMA foreign_key_check` on a copy of the production database (the
  `prod-db-snapshot` skill), list orphaned rows and agree with the user how to clean them up (a migration or a
  one-off script)
- Tests: deleting a user removes its dependent rows; inserting a row with a missing parent fails
- Check every place that deletes rows for behavior that relied on foreign keys being off

## SQLite busy timeout
`internal/database/database.go` sets neither `busy_timeout` nor WAL mode, so two writes at the same moment can fail
with `SQLITE_BUSY` and a 500. Consider `_pragma=busy_timeout(5000)` (and WAL, checking the backup job copes with the
`-wal` file), together with the foreign keys change above.
