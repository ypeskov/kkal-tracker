package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

const (
	// The cleanup migration and the one before it
	cleanupVersion       = 20261017110038
	beforeCleanupVersion = 20261017065036

	// An id no parent row has
	missingID = 999999
)

// userTables are the tables with a user_id foreign key to users (ON DELETE CASCADE)
var userTables = []string{
	"calorie_entries", "weight_history", "activation_tokens", "api_keys", "user_preferences", "user_ingredients",
}

var migrationsDir = filepath.Join("..", "..", MigrationsDir)

func TestMain(m *testing.M) {
	goose.SetLogger(goose.NopLogger())
	os.Exit(m.Run())
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newMigrator opens a migrator on a new database file in a temporary directory
func newMigrator(t *testing.T) (*Migrator, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "app.db")
	migrator, err := NewMigrator(path, migrationsDir, discardLogger())
	if err != nil {
		t.Fatalf("failed to open the migration connection: %v", err)
	}
	t.Cleanup(func() { migrator.Close() })
	return migrator, path
}

// newAppDB returns an app connection to a fully migrated database and the database path
func newAppDB(t *testing.T) (*sql.DB, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "app.db")
	if err := MigrateUp(path, migrationsDir, discardLogger()); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	return openApp(t, path), path
}

func openApp(t *testing.T, path string) *sql.DB {
	t.Helper()

	db, err := New(path, discardLogger())
	if err != nil {
		t.Fatalf("failed to open the database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) sql.Result {
	t.Helper()

	result, err := db.Exec(query, args...)
	if err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return result
}

func insertID(t *testing.T, db *sql.DB, query string, args ...any) int64 {
	t.Helper()

	id, err := mustExec(t, db, query, args...).LastInsertId()
	if err != nil {
		t.Fatalf("failed to read the inserted id: %v", err)
	}
	return id
}

func insertUser(t *testing.T, db *sql.DB, email string) int64 {
	t.Helper()
	return insertID(t, db, `INSERT INTO users (email, password_hash) VALUES (?, 'hash')`, email)
}

func insertGlobalIngredient(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()

	id := insertID(t, db, `INSERT INTO global_ingredients (kcal_per_100g) VALUES (100)`)
	mustExec(t, db, `INSERT INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (?, 'en_US', ?)`,
		id, name)
	return id
}

// insertUserRows inserts one row of the user into every table that references users. tag keeps unique columns
// unique across users
func insertUserRows(t *testing.T, db *sql.DB, userID, globalIngredientID int64, tag string) {
	t.Helper()

	mustExec(t, db, `INSERT INTO calorie_entries (user_id, food, calories, weight, kcal_per_100g, meal_datetime, updated_at)
		VALUES (?, ?, 150, 100, 150, '2026-01-15 08:30:00', '2026-01-15 08:30:00')`, userID, "Oatmeal "+tag)
	mustExec(t, db, `INSERT INTO weight_history (user_id, weight) VALUES (?, 80.5)`, userID)
	mustExec(t, db, `INSERT INTO activation_tokens (user_id, token, expires_at) VALUES (?, ?, '2030-01-01 00:00:00')`,
		userID, "token-"+tag)
	mustExec(t, db, `INSERT INTO api_keys (user_id, name, key_hash, key_prefix) VALUES (?, 'agent', ?, 'kk_')`,
		userID, "hash-"+tag)
	mustExec(t, db, `INSERT INTO user_preferences (user_id, pref_key, value) VALUES (?, 'reports', '{}')`, userID)
	mustExec(t, db, `INSERT INTO user_ingredients (user_id, name, kcal_per_100g, global_ingredient_id) VALUES (?, ?, 100, ?)`,
		userID, "Apple "+tag, globalIngredientID)
}

// dump returns the rows of the query as strings ("value|value|..."), for comparing table contents
func dump(t *testing.T, db *sql.DB, query string, args ...any) []string {
	t.Helper()

	rows, err := db.Query(query, args...)
	if err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		t.Fatalf("failed to read columns: %v", err)
	}
	var result []string
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			t.Fatalf("failed to scan a row: %v", err)
		}
		fields := make([]string, len(values))
		for i, value := range values {
			fields[i] = fmt.Sprint(value)
		}
		result = append(result, strings.Join(fields, "|"))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("failed to read rows: %v", err)
	}
	return result
}

// dumpAll returns the content of the tables with foreign keys and of their parents. The user ingredient named
// "Dangling" is left out: the cleanup migration changes it on purpose
func dumpAll(t *testing.T, db *sql.DB) map[string][]string {
	t.Helper()

	tables := append([]string{"users", "global_ingredients", "global_ingredient_names"}, userTables...)
	result := map[string][]string{}
	for _, table := range tables {
		query := "SELECT * FROM " + table + " ORDER BY rowid"
		if table == "user_ingredients" {
			query = "SELECT * FROM user_ingredients WHERE name <> 'Dangling' ORDER BY rowid"
		}
		result[table] = dump(t, db, query)
	}
	return result
}

func count(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func foreignKeyViolations(t *testing.T, db *sql.DB) []string {
	t.Helper()
	return dump(t, db, "PRAGMA foreign_key_check")
}

// assertPragmas checks foreign_keys and busy_timeout on several connections of the pool held open at the same time
func assertPragmas(t *testing.T, db *sql.DB, wantForeignKeys int) {
	t.Helper()

	ctx := context.Background()
	const connections = 3
	for i := range connections {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("connection %d: %v", i, err)
		}
		defer conn.Close()

		var foreignKeys, busyTimeout int
		if err := conn.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
			t.Fatalf("connection %d: PRAGMA foreign_keys: %v", i, err)
		}
		if err := conn.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			t.Fatalf("connection %d: PRAGMA busy_timeout: %v", i, err)
		}
		if foreignKeys != wantForeignKeys {
			t.Errorf("connection %d: foreign_keys = %d, want %d", i, foreignKeys, wantForeignKeys)
		}
		if busyTimeout != 5000 {
			t.Errorf("connection %d: busy_timeout = %d, want 5000", i, busyTimeout)
		}
	}

	if open := db.Stats().OpenConnections; open < connections {
		t.Errorf("open connections = %d, want at least %d distinct connections", open, connections)
	}
}

func TestAppConnectionsEnforceForeignKeysAndWaitOnLocks(t *testing.T) {
	db, _ := newAppDB(t)
	assertPragmas(t, db, 1)
}

func TestMigrationConnectionsHaveForeignKeysOff(t *testing.T) {
	migrator, _ := newMigrator(t)
	assertPragmas(t, migrator.db, 0)
}

func TestJournalModeStaysDelete(t *testing.T) {
	db := openApp(t, filepath.Join(t.TempDir(), "fresh.db"))
	mustExec(t, db, "CREATE TABLE t (id INTEGER PRIMARY KEY)")

	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if mode != "delete" {
		t.Errorf("journal_mode = %q, want delete", mode)
	}
}

// Paths the DSN cannot carry: '?' would open another file, an empty path would become a file named after the DSN
// with no pragmas applied. Both connection kinds reject them, and no file is created
func TestUnusableDatabasePathsAreRejected(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	tests := []struct{ name, path string }{
		{"question mark", filepath.Join(dir, "app.db?mode=ro")},
		{"empty", ""},
		{"blank", "  "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if db, err := New(tt.path, discardLogger()); err == nil {
				db.Close()
				t.Errorf("New accepted the path %q", tt.path)
			}
			if migrator, err := NewMigrator(tt.path, migrationsDir, discardLogger()); err == nil {
				migrator.Close()
				t.Errorf("NewMigrator accepted the path %q", tt.path)
			}
		})
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to list %s: %v", dir, err)
	}
	if len(entries) != 0 {
		t.Errorf("files were created: %v", entries)
	}
}

func TestForeignKeysAreEnforced(t *testing.T) {
	db, _ := newAppDB(t)
	userID := insertUser(t, db, "owner@example.com")
	globalID := insertGlobalIngredient(t, db, "Apple")

	tests := []struct {
		name  string
		query string
		args  []any
	}{
		{"calorie_entries.user_id", `INSERT INTO calorie_entries (user_id, food, calories, weight, kcal_per_100g,
			meal_datetime, updated_at) VALUES (?, 'Oatmeal', 150, 100, 150, '2026-01-15', '2026-01-15')`, []any{missingID}},
		{"weight_history.user_id", `INSERT INTO weight_history (user_id, weight) VALUES (?, 80)`, []any{missingID}},
		{"activation_tokens.user_id", `INSERT INTO activation_tokens (user_id, token, expires_at)
			VALUES (?, 'token', '2030-01-01')`, []any{missingID}},
		{"api_keys.user_id", `INSERT INTO api_keys (user_id, name, key_hash, key_prefix)
			VALUES (?, 'agent', 'hash', 'kk_')`, []any{missingID}},
		{"user_preferences.user_id", `INSERT INTO user_preferences (user_id, pref_key, value)
			VALUES (?, 'reports', '{}')`, []any{missingID}},
		{"user_ingredients.user_id", `INSERT INTO user_ingredients (user_id, name, kcal_per_100g, global_ingredient_id)
			VALUES (?, 'Apple', 100, ?)`, []any{missingID, globalID}},
		{"user_ingredients.global_ingredient_id", `INSERT INTO user_ingredients (user_id, name, kcal_per_100g,
			global_ingredient_id) VALUES (?, 'Pear', 100, ?)`, []any{userID, missingID}},
		{"global_ingredient_names.ingredient_id", `INSERT INTO global_ingredient_names (ingredient_id, language_code, name)
			VALUES (?, 'en_US', 'Pear')`, []any{missingID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := db.Exec(tt.query, tt.args...)
			if err == nil || !strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
				t.Fatalf("insert with a missing parent: error = %v, want a foreign key constraint error", err)
			}
		})
	}
}

func TestDeletingUserCascadesToTheUsersRows(t *testing.T) {
	db, _ := newAppDB(t)
	globalID := insertGlobalIngredient(t, db, "Apple")
	userA := insertUser(t, db, "a@example.com")
	userB := insertUser(t, db, "b@example.com")
	insertUserRows(t, db, userA, globalID, "a")
	insertUserRows(t, db, userB, globalID, "b")

	before := map[string][]string{}
	for _, table := range userTables {
		before[table] = dump(t, db, "SELECT * FROM "+table+" WHERE user_id = ? ORDER BY rowid", userB)
	}

	mustExec(t, db, "DELETE FROM users WHERE id = ?", userA)

	for _, table := range userTables {
		if n := count(t, db, "SELECT COUNT(*) FROM "+table+" WHERE user_id = ?", userA); n != 0 {
			t.Errorf("%s: %d rows of the deleted user remain", table, n)
		}
		after := dump(t, db, "SELECT * FROM "+table+" WHERE user_id = ? ORDER BY rowid", userB)
		if len(after) != 1 || !slices.Equal(after, before[table]) {
			t.Errorf("%s: the other user's rows changed: %v, want %v", table, after, before[table])
		}
	}
}

func TestDeletingGlobalIngredientRemovesNamesAndUnlinksUserIngredients(t *testing.T) {
	db, _ := newAppDB(t)
	deleted := insertGlobalIngredient(t, db, "Apple")
	mustExec(t, db, `INSERT INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (?, 'uk_UA', 'Яблуко')`,
		deleted)
	kept := insertGlobalIngredient(t, db, "Pear")
	userID := insertUser(t, db, "owner@example.com")
	mustExec(t, db, `INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, global_ingredient_id)
		VALUES (?, 'Apple', 52, 0.2, ?), (?, 'Pear', 57, 0.1, ?)`, userID, deleted, userID, kept)

	mustExec(t, db, "DELETE FROM global_ingredients WHERE id = ?", deleted)

	if n := count(t, db, "SELECT COUNT(*) FROM global_ingredient_names WHERE ingredient_id = ?", deleted); n != 0 {
		t.Errorf("%d names of the deleted ingredient remain", n)
	}
	if n := count(t, db, "SELECT COUNT(*) FROM global_ingredient_names WHERE ingredient_id = ?", kept); n != 1 {
		t.Errorf("names of the other ingredient = %d, want 1", n)
	}

	got := dump(t, db, `SELECT name, kcal_per_100g, fats, global_ingredient_id FROM user_ingredients
		WHERE user_id = ? ORDER BY name`, userID)
	want := []string{"Apple|52|0.2|<nil>", fmt.Sprintf("Pear|57|0.1|%d", kept)}
	if !slices.Equal(got, want) {
		t.Errorf("user ingredients = %q, want %q", got, want)
	}
}

func TestAppWriteWaitsForLockHeldByAnotherConnection(t *testing.T) {
	db, path := newAppDB(t)
	ctx := context.Background()

	// Another process or pod: an independent pool on the same file
	other, err := open(path, true, discardLogger())
	if err != nil {
		t.Fatalf("failed to open the second database: %v", err)
	}
	defer other.Close()
	holder, err := other.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to get a connection: %v", err)
	}
	defer holder.Close()

	if _, err := holder.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		t.Fatalf("BEGIN IMMEDIATE: %v", err)
	}
	if _, err := holder.ExecContext(ctx, `INSERT INTO users (email, password_hash) VALUES ('holder@example.com', 'hash')`); err != nil {
		t.Fatalf("write under the lock: %v", err)
	}

	const hold = 200 * time.Millisecond
	type outcome struct {
		err     error
		elapsed time.Duration
	}
	done := make(chan outcome, 1)
	start := time.Now()
	go func() {
		_, err := db.Exec(`INSERT INTO users (email, password_hash) VALUES ('app@example.com', 'hash')`)
		done <- outcome{err, time.Since(start)}
	}()

	time.Sleep(hold)
	if _, err := holder.ExecContext(ctx, "COMMIT"); err != nil {
		t.Fatalf("COMMIT: %v", err)
	}

	result := <-done
	if result.err != nil {
		t.Fatalf("app write under another connection's lock failed: %v", result.err)
	}
	if result.elapsed < hold {
		t.Errorf("app write took %v, want at least the lock hold time %v", result.elapsed, hold)
	}
	if n := count(t, db, "SELECT COUNT(*) FROM users"); n != 2 {
		t.Errorf("users = %d, want 2", n)
	}
}

func TestCleanupMigrationRemovesOrphans(t *testing.T) {
	migrator, _ := newMigrator(t)
	if err := migrator.UpTo(beforeCleanupVersion); err != nil {
		t.Fatalf("failed to migrate to the version before the cleanup: %v", err)
	}
	db := migrator.db // foreign keys off, as on a database written without enforcement

	// Valid rows
	globalID := insertGlobalIngredient(t, db, "Apple")
	userID := insertUser(t, db, "owner@example.com")
	insertUserRows(t, db, userID, globalID, "valid")
	mustExec(t, db, `INSERT INTO user_ingredients (user_id, name, kcal_per_100g) VALUES (?, 'Custom', 200)`, userID)
	valid := dumpAll(t, db)

	// One orphan per foreign key
	insertUserRows(t, db, missingID, globalID, "orphan")
	mustExec(t, db, `INSERT INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (?, 'en_US', 'Ghost')`,
		missingID)
	mustExec(t, db, `INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, global_ingredient_id)
		VALUES (?, 'Dangling', 52, 0.2, ?)`, userID, missingID)
	if n := len(foreignKeyViolations(t, db)); n != 8 {
		t.Fatalf("violations before the cleanup = %d, want 8", n)
	}

	if err := migrator.Up(); err != nil {
		t.Fatalf("failed to run the cleanup migration: %v", err)
	}

	for _, table := range userTables {
		if n := count(t, db, "SELECT COUNT(*) FROM "+table+" WHERE user_id = ?", missingID); n != 0 {
			t.Errorf("%s: %d orphaned rows remain", table, n)
		}
	}
	if n := count(t, db, "SELECT COUNT(*) FROM global_ingredient_names WHERE ingredient_id = ?", missingID); n != 0 {
		t.Errorf("global_ingredient_names: %d orphaned rows remain", n)
	}
	dangling := dump(t, db, `SELECT user_id, kcal_per_100g, fats, global_ingredient_id FROM user_ingredients
		WHERE name = 'Dangling'`)
	if want := []string{fmt.Sprintf("%d|52|0.2|<nil>", userID)}; !slices.Equal(dangling, want) {
		t.Errorf("user ingredient with a missing global ingredient = %q, want %q", dangling, want)
	}
	cleaned := dumpAll(t, db)
	for table, rows := range valid {
		if !slices.Equal(cleaned[table], rows) {
			t.Errorf("%s: valid rows changed:\n got %q\nwant %q", table, cleaned[table], rows)
		}
	}
	if violations := foreignKeyViolations(t, db); len(violations) != 0 {
		t.Errorf("foreign_key_check after the cleanup = %q, want empty", violations)
	}

	// Down is a no-op, and running the cleanup again on the clean data changes nothing
	if err := migrator.Down(); err != nil {
		t.Fatalf("failed to roll back the cleanup migration: %v", err)
	}
	if version, err := goose.GetDBVersion(db); err != nil || version != beforeCleanupVersion {
		t.Fatalf("version after down = %d (%v), want %d", version, err, beforeCleanupVersion)
	}
	if after := dumpAll(t, db); !mapsEqual(after, cleaned) {
		t.Errorf("down changed data")
	}
	if err := migrator.Up(); err != nil {
		t.Fatalf("failed to run the cleanup migration again: %v", err)
	}
	if after := dumpAll(t, db); !mapsEqual(after, cleaned) {
		t.Errorf("running the cleanup again changed data")
	}
	if again := dump(t, db, `SELECT global_ingredient_id FROM user_ingredients WHERE name = 'Dangling'`); len(again) != 1 {
		t.Errorf("the user ingredient with the dangling reference is gone after the second run")
	}
}

func mapsEqual(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, rows := range a {
		if !slices.Equal(rows, b[key]) {
			return false
		}
	}
	return true
}

func TestMigrationsFailOnForeignKeyViolations(t *testing.T) {
	migrator, _ := newMigrator(t)
	if err := migrator.Up(); err != nil {
		t.Fatalf("clean database: up = %v, want no error", err)
	}
	if err := migrator.Up(); err != nil {
		t.Fatalf("clean database, nothing pending: up = %v, want no error", err)
	}

	const secret = "secret-token-value"
	mustExec(t, migrator.db, `INSERT INTO activation_tokens (user_id, token, expires_at) VALUES (?, ?, '2030-01-01')`,
		missingID, secret)

	assertViolation := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, ErrForeignKeyViolations) {
			t.Fatalf("%s: error = %v, want %v", name, err, ErrForeignKeyViolations)
		}
		message := err.Error()
		if !strings.Contains(message, "activation_tokens: 1 row(s) reference a missing users row") {
			t.Errorf("%s: error %q does not name the child and the parent table", name, message)
		}
		if strings.Contains(message, secret) || strings.Contains(message, fmt.Sprint(missingID)) {
			t.Errorf("%s: error %q contains row values", name, message)
		}
	}

	assertViolation("up with nothing pending", migrator.Up())
	assertViolation("down", migrator.Down())
}

// A migration that rebuilds a parent table (the usual way to change a column in SQLite) keeps the children, because
// migrations run with foreign keys off. On an enforcing connection DROP TABLE would cascade-delete them
func TestParentTableRebuildKeepsChildren(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"00001_create.sql": `-- +goose Up
CREATE TABLE parents (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL);
CREATE TABLE children (id INTEGER PRIMARY KEY AUTOINCREMENT, parent_id INTEGER NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES parents (id) ON DELETE CASCADE);
-- +goose Down
DROP TABLE children;
DROP TABLE parents;
`,
		"00002_rebuild_parents.sql": `-- +goose Up
CREATE TABLE parents_new (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, note TEXT);
INSERT INTO parents_new (id, name) SELECT id, name FROM parents;
DROP TABLE parents;
ALTER TABLE parents_new RENAME TO parents;
-- +goose Down
SELECT 1;
`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	path := filepath.Join(t.TempDir(), "app.db")
	migrator, err := NewMigrator(path, dir, discardLogger())
	if err != nil {
		t.Fatalf("failed to open the migration connection: %v", err)
	}
	defer migrator.Close()
	if err := migrator.UpTo(1); err != nil {
		t.Fatalf("failed to create the tables: %v", err)
	}

	app := openApp(t, path)
	parentID := insertID(t, app, `INSERT INTO parents (name) VALUES ('parent')`)
	mustExec(t, app, `INSERT INTO children (parent_id) VALUES (?), (?)`, parentID, parentID)

	if err := migrator.Up(); err != nil {
		t.Fatalf("failed to rebuild the parent table: %v", err)
	}
	if n := count(t, app, "SELECT COUNT(*) FROM children"); n != 2 {
		t.Errorf("children after the rebuild = %d, want 2", n)
	}
}
