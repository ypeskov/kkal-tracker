package services

import (
	"database/sql"
	"errors"
	"log/slog"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	userRepo   repositories.UserRepository
	jwtService *auth.JWTService
	logger     *slog.Logger
}

func NewAuthService(userRepo repositories.UserRepository, jwtService *auth.JWTService, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		s.logger.Error("failed to get user by email", "error", err)
		return nil, "", err
	}

	if !user.CheckPassword(password) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) GetCurrentUser(userID int) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("failed to get user by ID", "error", err)
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Register(email, password, languageCode string) (*models.User, string, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("failed to check existing user", "error", err)
		return nil, "", err
	}

	if existingUser != nil {
		return nil, "", errors.New("user already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, "", err
	}

	// Create new user with hashed password
	user, err := s.userRepo.CreateWithLanguage(email, string(hashedPassword), languageCode)
	if err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, "", err
	}

	// Note: CreateWithLanguage already copies ingredients, so we don't need to do it again
	// The ingredient copying is handled in the repository transaction
	// Ingredient copying is now handled in CreateWithLanguage

	// Generate token for new user
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate token for new user", "error", err)
		return nil, "", err
	}

	return user, token, nil
}
