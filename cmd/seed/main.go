package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/database"
	"ypeskov/kkal-tracker/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	clean := flag.Bool("clean", false, "Delete the global ingredients before seeding (dev only)")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Warning: Error loading .env file: %v", err)
	}
	// Load configuration
	cfg := config.New()

	// Initialize logger
	appLogger := logger.New(cfg)

	// Initialize database (an app connection: foreign keys are enforced)
	db, err := database.New(cfg.DatabasePath, appLogger)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Get SQL directory path
	sqlDir := "cmd/seed/sql"
	if flag.NArg() > 0 {
		sqlDir = flag.Arg(0)
	}

	// Read all SQL files (before anything is deleted)
	sqlFiles, err := readSQLFiles(sqlDir)
	if err != nil {
		log.Fatalf("Failed to read SQL files: %v", err)
	}

	if len(sqlFiles) == 0 {
		appLogger.Info("No SQL files found in directory", "path", sqlDir)
		return
	}

	if err := seed(db, sqlFiles, *clean, appLogger); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	appLogger.Info("Database seeding completed successfully", "files_executed", len(sqlFiles))
}

type sqlFile struct {
	name    string
	content string
}

func readSQLFiles(dir string) ([]sqlFile, error) {
	var files []sqlFile

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .sql files
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		files = append(files, sqlFile{
			name:    filepath.Base(path),
			content: string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort files by name to ensure consistent execution order
	sort.Slice(files, func(i, j int) bool {
		return files[i].name < files[j].name
	})

	return files, nil
}

// seed runs the clean step (if requested) and the SQL files in one transaction, so a failing file leaves the
// database as it was: seed-clean never ends with the global ingredients deleted and not re-seeded
func seed(db *sql.DB, files []sqlFile, clean bool, appLogger *slog.Logger) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if clean {
		if err := cleanGlobalIngredients(tx); err != nil {
			return fmt.Errorf("failed to delete the global ingredients: %w", err)
		}
		appLogger.Info("Deleted the global ingredients")
	}

	for _, file := range files {
		appLogger.Info("Executing SQL file", "file", file.name)
		if err := executeSQLFile(tx, file.content); err != nil {
			return fmt.Errorf("%s: %w", file.name, err)
		}
		appLogger.Info("Successfully executed SQL file", "file", file.name)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// cleanGlobalIngredients deletes the global ingredients before a re-seed. Foreign keys are enforced, so the users'
// ingredients keep their rows with global_ingredient_id set to NULL. The names are deleted by cascade; the explicit
// delete also removes names that were orphaned before enforcement. The id counters are reset, so the re-seeded
// ingredients get the ids 1..N the names in 01_ingredients.sql refer to
func cleanGlobalIngredients(tx *sql.Tx) error {
	statements := []string{
		"DELETE FROM global_ingredients",
		"DELETE FROM global_ingredient_names",
		"DELETE FROM sqlite_sequence WHERE name IN ('global_ingredients', 'global_ingredient_names')",
	}
	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nStatement: %s", err, stmt)
		}
	}

	return nil
}

func executeSQLFile(tx *sql.Tx, content string) error {
	// Split the content by semicolons to handle multiple statements
	// This is a simple approach - for production, consider using a proper SQL parser
	statements := strings.Split(content, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// Add semicolon back for proper SQL statement
		stmt = stmt + ";"

		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\nStatement: %s", err, stmt)
		}
	}

	return nil
}
