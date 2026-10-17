package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// BusyTimeout is how long a connection waits for a lock held by another connection (another request, the other pod
// during a rollout, the backup, cmd/migrate) before the statement fails with SQLITE_BUSY
const BusyTimeout = 5 * time.Second

// New opens the application database: every pooled connection enforces foreign keys and waits BusyTimeout on locks.
// databasePath is a plain file path
func New(databasePath string, log *slog.Logger) (*sql.DB, error) {
	return open(databasePath, true, log)
}

// open opens the database with the pragmas in the DSN. modernc.org/sqlite runs the _pragma parameters on every
// connection it opens, so they hold for each connection of the pool, not only the first one
func open(databasePath string, foreignKeys bool, log *slog.Logger) (*sql.DB, error) {
	// The driver cuts the DSN at the first '?', so such a path would silently open another file. With an empty path
	// nothing is cut: the whole DSN becomes the file name and no pragma is applied
	if strings.TrimSpace(databasePath) == "" {
		return nil, errors.New("database path is empty")
	}
	if strings.Contains(databasePath, "?") {
		return nil, fmt.Errorf("database path must not contain '?': %q", databasePath)
	}

	dir := filepath.Dir(databasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dsn(databasePath, foreignKeys))
	if err != nil {
		log.Error("failed to open database", "error", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// dsn adds the connection pragmas to a plain file path. The journal mode is deliberately left as is (rollback
// journal, "delete")
func dsn(databasePath string, foreignKeys bool) string {
	foreignKeysValue := 0
	if foreignKeys {
		foreignKeysValue = 1
	}
	return fmt.Sprintf("%s?_pragma=foreign_keys(%d)&_pragma=busy_timeout(%d)",
		databasePath, foreignKeysValue, BusyTimeout.Milliseconds())
}
