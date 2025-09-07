package services

import (
	"database/sql"
	"errors"
	"log/slog"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound      = errors.New("user not found")
)

type AuthService struct {
	userRepo   *models.UserRepository
	jwtService *auth.JWTService
	logger     *slog.Logger
}

func NewAuthService(userRepo *models.UserRepository, jwtService *auth.JWTService, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrInvalidCredentials
		}
		s.logger.Error("Failed to get user by email", "error", err)
		return nil, "", err
	}

	if !user.CheckPassword(password) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err)
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) GetCurrentUser(userID int) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		s.logger.Error("Failed to get user by ID", "error", err)
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Register(email, password string) (*models.User, string, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("Failed to check existing user", "error", err)
		return nil, "", err
	}
	
	if existingUser != nil {
		return nil, "", errors.New("user already exists")
	}

	// Create new user
	user := &models.User{
		Email: email,
	}
	
	if err := user.SetPassword(password); err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, "", err
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, "", err
	}

	// Generate token for new user
	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.logger.Error("Failed to generate token for new user", "error", err)
		return nil, "", err
	}

	return user, token, nil
}