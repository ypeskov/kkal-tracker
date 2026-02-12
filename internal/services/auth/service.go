package auth

import (
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
	emailservice "ypeskov/kkal-tracker/internal/services/email"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo         repositories.UserRepository
	tokenRepo        repositories.ActivationTokenRepository
	jwtService       *auth.JWTService
	emailService     *emailservice.Service
	logger           *slog.Logger
}

func New(
	userRepo repositories.UserRepository,
	tokenRepo repositories.ActivationTokenRepository,
	jwtService *auth.JWTService,
	emailService *emailservice.Service,
	logger *slog.Logger,
) *Service {
	return &Service{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		jwtService:   jwtService,
		emailService: emailService,
		logger:       logger.With("service", "auth"),
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

	// Check if user is activated
	if !user.IsActive {
		s.logger.Debug("Login failed - user not activated", "email", email, "user_id", user.ID)
		return nil, "", ErrUserNotActivated
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

func (s *Service) Register(email, password, languageCode string, skipActivation bool) (*models.User, string, error) {
	s.logger.Debug("Register called", "email", email, "language", languageCode, "skip_activation", skipActivation)

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("failed to check existing user", "error", err)
		return nil, "", err
	}

	if existingUser != nil {
		s.logger.Debug("Register failed - user already exists", "email", email)
		return nil, "", ErrUserAlreadyExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, "", err
	}

	if skipActivation {
		// CLI/Admin path: Create ACTIVE user immediately, no email
		s.logger.Debug("Creating active user (skipActivation=true)", "email", email)

		user, err := s.userRepo.CreateWithLanguage(email, string(hashedPassword), languageCode, true)
		if err != nil {
			s.logger.Error("failed to create active user", "error", err)
			return nil, "", err
		}

		// Generate JWT token for immediate use
		token, err := s.jwtService.GenerateToken(user.ID, user.Email)
		if err != nil {
			s.logger.Error("Failed to generate token for new active user", "error", err)
			return nil, "", err
		}

		s.logger.Info("Active user created successfully", "email", email, "user_id", user.ID)
		return user, token, nil
	} else {
		// Web registration path: Create INACTIVE user, send activation email
		s.logger.Debug("Creating inactive user (skipActivation=false)", "email", email)

		user, err := s.userRepo.CreateWithLanguage(email, string(hashedPassword), languageCode, false)
		if err != nil {
			s.logger.Error("failed to create inactive user", "error", err)
			return nil, "", err
		}

		// Generate activation token (expires in 24 hours)
		expiresAt := time.Now().Add(24 * time.Hour)
		activationToken, err := s.tokenRepo.Create(user.ID, expiresAt)
		if err != nil {
			s.logger.Error("Failed to create activation token", "error", err, "user_id", user.ID)
			// Rollback: delete the created user
			if deleteErr := s.userRepo.Delete(user.ID); deleteErr != nil {
				s.logger.Error("Failed to rollback user creation", "error", deleteErr, "user_id", user.ID)
			}
			return nil, "", err
		}

		// Send activation email
		language := languageCode
		if user.Language != nil {
			language = *user.Language
		}

		err = s.emailService.SendActivationEmail(user.Email, activationToken.Token, language)
		if err != nil {
			s.logger.Error("Failed to send activation email", "error", err, "user_id", user.ID)

			// Rollback: delete activation token and user
			if tokenDeleteErr := s.tokenRepo.Delete(activationToken.Token); tokenDeleteErr != nil {
				s.logger.Error("Failed to rollback activation token", "error", tokenDeleteErr, "user_id", user.ID)
			}
			if userDeleteErr := s.userRepo.Delete(user.ID); userDeleteErr != nil {
				s.logger.Error("Failed to rollback user creation", "error", userDeleteErr, "user_id", user.ID)
			}

			s.logger.Warn("Registration rolled back due to email failure", "email", email, "user_id", user.ID)
			// Return error to frontend - registration failed
			return nil, "", ErrEmailSendFailed
		}

		s.logger.Info("Inactive user created, activation email sent", "email", email, "user_id", user.ID)
		return user, "", nil // No JWT token for inactive users
	}
}

// ActivateUser activates a user account using an activation token
func (s *Service) ActivateUser(token string) error {
	tokenPreview := token
	if len(token) >= 8 {
		tokenPreview = token[:8] + "..."
	}
	s.logger.Debug("ActivateUser called", "token", tokenPreview)

	// Get activation token from repository
	activationToken, err := s.tokenRepo.GetByToken(token)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			s.logger.Debug("Activation failed - token not found", "token", tokenPreview)
			return ErrInvalidToken
		}
		s.logger.Error("Failed to get activation token", "error", err)
		return err
	}

	// Check if token is expired
	if activationToken.IsExpired() {
		s.logger.Debug("Activation failed - token expired", "token", token[:8]+"...", "expires_at", activationToken.ExpiresAt)
		// Delete expired token
		_ = s.tokenRepo.Delete(token)
		return ErrInvalidToken
	}

	// Activate the user
	err = s.userRepo.ActivateUser(activationToken.UserID)
	if err != nil {
		s.logger.Error("Failed to activate user", "error", err, "user_id", activationToken.UserID)
		return err
	}

	// Delete used activation token
	err = s.tokenRepo.Delete(token)
	if err != nil {
		s.logger.Error("Failed to delete activation token", "error", err, "token", token[:8]+"...")
		// Don't return error - user is already activated
	}

	s.logger.Info("User activated successfully", "user_id", activationToken.UserID)
	return nil
}
