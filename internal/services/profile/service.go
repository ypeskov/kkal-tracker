package profile

import (
	"database/sql"
	"log/slog"

	"ypeskov/kkal-tracker/internal/repositories"
)

type Service struct {
	db             *sql.DB
	userRepo       repositories.UserRepository
	weightHistRepo repositories.WeightHistoryRepository
	logger         *slog.Logger
}

func New(db *sql.DB, userRepo repositories.UserRepository, weightHistRepo repositories.WeightHistoryRepository, logger *slog.Logger) *Service {
	return &Service{
		db:             db,
		userRepo:       userRepo,
		weightHistRepo: weightHistRepo,
		logger:         logger.With("service", "profile"),
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

	// Get latest weight from weight history
	var weight *float64
	latestWeight, err := s.weightHistRepo.GetLatestByUserID(userID)
	if err != nil {
		s.logger.Warn("Failed to get latest weight", "user_id", userID, "error", err)
		// Continue without weight data
	} else if latestWeight != nil {
		weight = &latestWeight.Weight
	}

	// Convert domain model to DTO
	response := &ProfileResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Age:       user.Age,
		Height:    user.Height,
		Weight:    weight,
		Gender:    user.Gender,
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

	// Update user profile (without weight - now managed only in weight history)
	if err := s.userRepo.UpdateProfile(userID, req.FirstName, req.LastName, req.Email, req.Age, req.Height, req.Gender, nil, req.Language); err != nil {
		s.logger.Error("Failed to update profile", "user_id", userID, "error", err)
		return err
	}

	// Weight is now managed only through weight history, not profile updates

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", "user_id", userID, "error", err)
		return err
	}

	s.logger.Debug("UpdateProfile completed successfully", "user_id", userID)
	return nil
}
