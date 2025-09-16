package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/database"
	"ypeskov/kkal-tracker/internal/logger"
	"ypeskov/kkal-tracker/internal/repositories"

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
	appLogger := logger.New(cfg)

	db, err := database.New(cfg.DatabasePath, appLogger)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	err = database.Migrate(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	userRepo := repositories.NewUserRepository(db, appLogger, repositories.DialectSQLite)

	user, err := userRepo.CreateWithLanguage(email, string(hashedPassword), languageCode)
	if err != nil {
		log.Fatal("Failed to create user:", err)
	}

	fmt.Printf("User created successfully: ID=%d, Email=%s, Language=%s\n", user.ID, user.Email, languageCode)
}