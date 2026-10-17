package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/pressly/goose/v3"
)

// MigrationsDir is the migrations directory relative to the repository root, where the commands run
const MigrationsDir = "migrations"

// Migrator applies the goose migrations on its own connection with foreign keys off. Under enforcement, a migration
// that rebuilds a parent table (DROP TABLE + rename) would cascade-delete the children, and PRAGMA foreign_keys
// cannot be switched off inside the transaction goose runs each migration in. After up and down, PRAGMA
// foreign_key_check reports the rows a migration left without a parent
type Migrator struct {
	db  *sql.DB
	dir string
}

// NewMigrator opens the database for migrations: foreign keys off, BusyTimeout on locks
func NewMigrator(databasePath, migrationsDir string, log *slog.Logger) (*Migrator, error) {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, err
	}

	db, err := open(databasePath, false, log)
	if err != nil {
		return nil, err
	}

	return &Migrator{db: db, dir: migrationsDir}, nil
}

// MigrateUp applies all pending migrations on a separate migration connection and closes it
func MigrateUp(databasePath, migrationsDir string, log *slog.Logger) error {
	migrator, err := NewMigrator(databasePath, migrationsDir, log)
	if err != nil {
		return err
	}
	defer migrator.Close()

	return migrator.Up()
}

func (m *Migrator) Close() error {
	return m.db.Close()
}

// Up applies all pending migrations, then checks foreign keys (also when nothing was pending)
func (m *Migrator) Up() error {
	if err := goose.Up(m.db, m.dir); err != nil {
		return err
	}
	return m.checkForeignKeys()
}

// UpTo applies the pending migrations up to and including version, then checks foreign keys
func (m *Migrator) UpTo(version int64) error {
	if err := goose.UpTo(m.db, m.dir, version); err != nil {
		return err
	}
	return m.checkForeignKeys()
}

// Down rolls back the last migration, then checks foreign keys
func (m *Migrator) Down() error {
	if err := goose.Down(m.db, m.dir); err != nil {
		return err
	}
	return m.checkForeignKeys()
}

// Status prints the status of every migration
func (m *Migrator) Status() error {
	return goose.Status(m.db, m.dir)
}

// ErrForeignKeyViolations means the database has rows whose parent row is missing
var ErrForeignKeyViolations = errors.New("foreign key check failed")

// checkForeignKeys runs PRAGMA foreign_key_check. The error names each child table with its number of rows and the
// parent table, never row content
func (m *Migrator) checkForeignKeys() error {
	rows, err := m.db.Query("PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("failed to run foreign key check: %w", err)
	}
	defer rows.Close()

	type reference struct{ table, parent string }
	counts := map[reference]int{}
	for rows.Next() {
		// Columns: child table, rowid (NULL for WITHOUT ROWID tables), parent table, foreign key index
		var ref reference
		var rowID, foreignKeyID any
		if err := rows.Scan(&ref.table, &rowID, &ref.parent, &foreignKeyID); err != nil {
			return fmt.Errorf("failed to read foreign key check: %w", err)
		}
		counts[ref]++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to read foreign key check: %w", err)
	}

	if len(counts) == 0 {
		return nil
	}

	violations := make([]string, 0, len(counts))
	for ref, count := range counts {
		violations = append(violations, fmt.Sprintf("%s: %d row(s) reference a missing %s row", ref.table, count, ref.parent))
	}
	slices.Sort(violations)

	return fmt.Errorf("%w after the migration: %s", ErrForeignKeyViolations, strings.Join(violations, "; "))
}
