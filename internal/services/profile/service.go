package profile

import (
	"database/sql"
	"log/slog"

	"ypeskov/kkal-tracker/internal/repositories"
)

type Service struct {
	db       *sql.DB
	userRepo repositories.UserRepository
	logger   *slog.Logger
}

func New(db *sql.DB, userRepo repositories.UserRepository, logger *slog.Logger) *Service {
	return &Service{
		db:       db,
		userRepo: userRepo,
		logger:   logger.With("service", "profile"),
	}
}

// GetProfile retrieves the user profile and returns a DTO
func (s *Service) GetProfile(userID int) (*ProfileResponse, error) {
	s.logger.Debug("GetProfile called", "user_id", userID)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("Failed to get user", "user_id", userID, "error", err)
		return nil, err
	}

	// Default language if not set
	language := "en_US"
	if user.Language != nil {
		language = *user.Language
	}

	// Convert domain model to DTO
	response := &ProfileResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Age:       user.Age,
		Height:    user.Height,
		Weight:    user.Weight,
		Language:  language,
	}

	s.logger.Debug("GetProfile completed successfully", "user_id", userID, "email", response.Email)
	return response, nil
}

// UpdateProfile updates the user profile, passing DTO directly to repository
func (s *Service) UpdateProfile(userID int, req *ProfileUpdateRequest) error {
	s.logger.Debug("UpdateProfile called", "user_id", userID, "email", req.Email, "first_name", req.FirstName, "last_name", req.LastName)

	// Start transaction for atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("Failed to begin transaction", "user_id", userID, "error", err)
		return err
	}
	defer tx.Rollback()

	// Update user profile (including weight in users table)
	if err := s.userRepo.UpdateProfile(userID, req.FirstName, req.LastName, req.Email, req.Age, req.Height, req.Weight, req.Language); err != nil {
		s.logger.Error("Failed to update profile", "user_id", userID, "error", err)
		return err
	}

	// If weight is provided, also add it to weight history for tracking
	if req.Weight != nil {
		if err := s.userRepo.AddWeightEntry(userID, *req.Weight); err != nil {
			s.logger.Error("Failed to add weight entry", "user_id", userID, "weight", *req.Weight, "error", err)
			return err
		}
		s.logger.Debug("Added weight to history", "user_id", userID, "weight", *req.Weight)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", "user_id", userID, "error", err)
		return err
	}

	s.logger.Debug("UpdateProfile completed successfully", "user_id", userID)
	return nil
}
