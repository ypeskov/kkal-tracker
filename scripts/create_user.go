package main

import (
	"fmt"
	"log"
	"os"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/database"
	"ypeskov/kkal-tracker/internal/models"

	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <email> <password>\n", os.Args[0])
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	cfg := config.New()

	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	err = database.Migrate(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	userRepo := models.NewUserRepository(db)

	user, err := userRepo.Create(email, password)
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}

	fmt.Printf("User created successfully: ID=%d, Email=%s\n", user.ID, user.Email)
}