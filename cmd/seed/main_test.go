package main

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"ypeskov/kkal-tracker/internal/database"

	"github.com/pressly/goose/v3"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newSeededDB returns an app connection to a migrated database seeded with the real seed files
func newSeededDB(t *testing.T) (*sql.DB, []sqlFile) {
	t.Helper()

	goose.SetLogger(goose.NopLogger())
	path := filepath.Join(t.TempDir(), "app.db")
	if err := database.MigrateUp(path, filepath.Join("..", "..", database.MigrationsDir), discardLogger()); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
	db, err := database.New(path, discardLogger())
	if err != nil {
		t.Fatalf("failed to open the database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	files, err := readSQLFiles("sql")
	if err != nil {
		t.Fatalf("failed to read the seed files: %v", err)
	}
	if err := seed(db, files, false, discardLogger()); err != nil {
		t.Fatalf("failed to seed: %v", err)
	}
	return db, files
}

func queryInt(t *testing.T, db *sql.DB, query string) int {
	t.Helper()

	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return n
}

func TestSeedCleanResetsGlobalIngredients(t *testing.T) {
	db, files := newSeededDB(t)
	ctx := context.Background()

	// A user whose ingredient links a global ingredient, and a name orphaned while foreign keys were not enforced
	if _, err := db.Exec(`INSERT INTO users (email, password_hash) VALUES ('owner@example.com', 'hash')`); err != nil {
		t.Fatalf("failed to insert a user: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO user_ingredients (user_id, name, kcal_per_100g, global_ingredient_id)
		SELECT id, 'Linked', 45, 1 FROM users WHERE email = 'owner@example.com'`); err != nil {
		t.Fatalf("failed to insert a user ingredient: %v", err)
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to get a connection: %v", err)
	}
	for _, stmt := range []string{
		"PRAGMA foreign_keys = OFF",
		`INSERT INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (999, 'de_DE', 'Geist')`,
		"PRAGMA foreign_keys = ON",
	} {
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	conn.Close()

	if err := seed(db, files, true, discardLogger()); err != nil {
		t.Fatalf("seed -clean failed: %v", err)
	}

	if n := queryInt(t, db, "SELECT COUNT(*) FROM pragma_foreign_key_check"); n != 0 {
		t.Errorf("foreign_key_check rows = %d, want 0", n)
	}
	if n := queryInt(t, db, "SELECT COUNT(*) FROM global_ingredients WHERE id BETWEEN 1 AND 50"); n != 50 {
		t.Errorf("global ingredients with ids 1..50 = %d, want 50", n)
	}
	if n := queryInt(t, db, "SELECT COUNT(*) FROM global_ingredients"); n != 50 {
		t.Errorf("global ingredients = %d, want 50", n)
	}
	if n := queryInt(t, db, "SELECT COUNT(*) FROM global_ingredient_names"); n != 150 {
		t.Errorf("global ingredient names = %d, want 150", n)
	}
	var link sql.NullInt64
	if err := db.QueryRow(`SELECT global_ingredient_id FROM user_ingredients WHERE name = 'Linked'`).Scan(&link); err != nil {
		t.Fatalf("the user's ingredient is gone: %v", err)
	}
	if link.Valid {
		t.Errorf("global_ingredient_id = %d, want NULL", link.Int64)
	}
}

func TestSeedFailureKeepsGlobalIngredients(t *testing.T) {
	db, files := newSeededDB(t)
	before := queryInt(t, db, "SELECT COUNT(*) FROM global_ingredients")
	beforeNames := queryInt(t, db, "SELECT COUNT(*) FROM global_ingredient_names")
	beforeLinks := queryInt(t, db, "SELECT COUNT(*) FROM user_ingredients WHERE global_ingredient_id IS NOT NULL")

	broken := append(files, sqlFile{name: "99_broken.sql", content: "INSERT INTO missing_table VALUES (1);"})
	if err := seed(db, broken, true, discardLogger()); err == nil {
		t.Fatal("seed with a broken file succeeded")
	}

	if n := queryInt(t, db, "SELECT COUNT(*) FROM global_ingredients"); n != before {
		t.Errorf("global ingredients = %d, want %d (the clean step must be rolled back)", n, before)
	}
	if n := queryInt(t, db, "SELECT COUNT(*) FROM global_ingredient_names"); n != beforeNames {
		t.Errorf("global ingredient names = %d, want %d", n, beforeNames)
	}
	if n := queryInt(t, db, "SELECT COUNT(*) FROM user_ingredients WHERE global_ingredient_id IS NOT NULL"); n != beforeLinks {
		t.Errorf("linked user ingredients = %d, want %d", n, beforeLinks)
	}
}
