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
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Printf("Usage: %s <email> <password> [language_code]\n", os.Args[0])
		fmt.Printf("Language codes: en_US, uk_UA, ru_UA (default: en_US)\n")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	languageCode := "en_US"
	if len(os.Args) == 4 {
		languageCode = os.Args[3]
	}

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

	user, err := userRepo.CreateWithLanguage(email, password, languageCode)
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}

	fmt.Printf("User created successfully: ID=%d, Email=%s, Language=%s\n", user.ID, user.Email, languageCode)
}