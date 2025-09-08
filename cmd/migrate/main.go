package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/database"
	"ypeskov/kkal-tracker/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	var command string
	flag.StringVar(&command, "cmd", "up", "Migration command: up, down, status")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	cfg := config.New()

	appLogger := logger.New(cfg)

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		appLogger.Error("Failed to initialize database:", err)
		os.Exit(1)
	}
	defer db.Close()

	switch command {
	case "up":
		fmt.Println("Running migrations...")
		err = database.Migrate(db)
		if err != nil {
			log.Fatal("Failed to run migrations:", err)
		}
		fmt.Println("Migrations completed successfully!")

	case "down":
		fmt.Println("Rolling back last migration...")
		err = database.MigrateDown(db)
		if err != nil {
			log.Fatal("Failed to rollback migration:", err)
		}
		fmt.Println("Rollback completed successfully!")

	case "status":
		fmt.Println("Migration status:")
		err = database.MigrateStatus(db)
		if err != nil {
			log.Fatal("Failed to get migration status:", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: up, down, status")
		os.Exit(1)
	}
}
