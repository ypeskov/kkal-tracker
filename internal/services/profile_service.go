package services

import (
	"database/sql"
	"log/slog"

	"ypeskov/kkal-tracker/internal/dto"
	"ypeskov/kkal-tracker/internal/repositories"
)

type ProfileService struct {
	db       *sql.DB
	userRepo repositories.UserRepository
	logger   *slog.Logger
}

func NewProfileService(db *sql.DB, userRepo repositories.UserRepository, logger *slog.Logger) *ProfileService {
	return &ProfileService{
		db:       db,
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetProfile retrieves the user profile and returns a DTO
func (s *ProfileService) GetProfile(userID int) (*dto.ProfileResponse, error) {
	s.logger.Debug("Getting profile", slog.Int("user_id", userID))

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("Failed to get user", slog.String("error", err.Error()))
		return nil, err
	}

	// Default language if not set
	language := "en_US"
	if user.Language != nil {
		language = *user.Language
	}

	// Convert domain model to DTO
	response := &dto.ProfileResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Age:       user.Age,
		Height:    user.Height,
		Weight:    user.Weight,
		Language:  language,
	}

	return response, nil
}

// UpdateProfile updates the user profile, passing DTO directly to repository
func (s *ProfileService) UpdateProfile(userID int, req *dto.ProfileUpdateRequest) error {
	s.logger.Debug("Updating profile", slog.Int("user_id", userID))

	// Start transaction for atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	// Update user profile (including weight in users table)
	if err := s.userRepo.UpdateProfile(userID, req); err != nil {
		s.logger.Error("Failed to update profile", slog.String("error", err.Error()))
		return err
	}

	// If weight is provided, also add it to weight history for tracking
	if req.Weight != nil {
		if err := s.userRepo.AddWeightEntry(userID, *req.Weight); err != nil {
			s.logger.Error("Failed to add weight entry", slog.String("error", err.Error()))
			return err
		}
		s.logger.Debug("Added weight to history",
			slog.Int("user_id", userID),
			slog.Float64("weight", *req.Weight))
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("Profile updated successfully", slog.Int("user_id", userID))
	return nil
}