package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/database"
	"ypeskov/kkal-tracker/internal/logger"
)

func main() {
	// Load configuration
	cfg := config.New()

	// Initialize logger
	appLogger := logger.New(cfg)

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Get SQL directory path
	sqlDir := "cmd/seed/sql"
	if len(os.Args) > 1 {
		sqlDir = os.Args[1]
	}

	// Read all SQL files
	sqlFiles, err := readSQLFiles(sqlDir)
	if err != nil {
		log.Fatalf("Failed to read SQL files: %v", err)
	}

	if len(sqlFiles) == 0 {
		appLogger.Info("No SQL files found in directory", "path", sqlDir)
		return
	}

	// Execute each SQL file
	for _, file := range sqlFiles {
		appLogger.Info("Executing SQL file", "file", file.name)
		
		if err := executeSQLFile(db, file.content); err != nil {
			appLogger.Error("Failed to execute SQL file", "file", file.name, "error", err)
			log.Fatalf("Seeding failed: %v", err)
		}
		
		appLogger.Info("Successfully executed SQL file", "file", file.name)
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

func executeSQLFile(db *sql.DB, content string) error {
	// Split the content by semicolons to handle multiple statements
	// This is a simple approach - for production, consider using a proper SQL parser
	statements := strings.Split(content, ";")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}