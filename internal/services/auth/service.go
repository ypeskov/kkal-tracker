package auth

import (
	"database/sql"
	"errors"
	"log/slog"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo   repositories.UserRepository
	jwtService *auth.JWTService
	logger     *slog.Logger
}

func New(userRepo repositories.UserRepository, jwtService *auth.JWTService, logger *slog.Logger) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger.With("service", "auth"),
	}
}

func (s *Service) Login(email, password string) (*models.User, string, error) {
	s.logger.Debug("Login called", "email", email)

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrInvalidCredentials
		}
		s.logger.Error("failed to get user by email", "error", err)
		return nil, "", err
	}

	if !user.CheckPassword(password) {
		s.logger.Debug("Login failed - invalid password", "email", email)
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return nil, "", err
	}

	s.logger.Debug("Login successful", "email", email, "user_id", user.ID)
	return user, token, nil
}

func (s *Service) GetCurrentUser(userID int) (*models.User, error) {
	s.logger.Debug("GetCurrentUser called", "user_id", userID)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		s.logger.Error("failed to get user by ID", "error", err)
		return nil, err
	}

	s.logger.Debug("GetCurrentUser successful", "user_id", userID, "email", user.Email)
	return user, nil
}

func (s *Service) Register(email, password, languageCode string) (*models.User, string, error) {
	s.logger.Debug("Register called", "email", email, "language", languageCode)

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("failed to check existing user", "error", err)
		return nil, "", err
	}

	if existingUser != nil {
		s.logger.Debug("Register failed - user already exists", "email", email)
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

	s.logger.Debug("Register successful", "email", email, "user_id", user.ID, "language", languageCode)
	return user, token, nil
}
