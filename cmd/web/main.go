package main

import (
	"errors"
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
		// Can't use structured log yet, config not loaded
		// This is acceptable for .env loading warnings
	}

	cfg := config.New()
	log := logger.New(cfg)

	// A missing .env file is expected when configuration comes from the environment
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Warn("error loading .env file", "error", err)
	}

	db, err := database.New(cfg.DatabasePath, log)
	if err != nil {
		log.Error("failed to connect to database, exiting", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	s := server.New(cfg, log, db, web.StaticFiles)

	srv := s.Start()

	log.Info("Starting server", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
