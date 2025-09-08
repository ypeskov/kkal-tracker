package main

import (
	"os"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/database"
	"ypeskov/kkal-tracker/internal/logger"
	"ypeskov/kkal-tracker/internal/server"
	"ypeskov/kkal-tracker/web"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		// Can't use structured logger yet, config not loaded
		// This is acceptable for .env loading warnings
	}

	cfg := config.New()
	log := logger.New(cfg)
	log.Info("Loading config...", "config", cfg)

	if err != nil {
		log.Warn("Error loading .env file", "error", err)
	}

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	srv := server.New(cfg, log, db, web.StaticFiles)

	log.Info("Starting server", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
