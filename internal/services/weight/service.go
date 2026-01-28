package weight

import (
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

type Service struct {
	weightRepo repositories.WeightHistoryRepository
	logger     *slog.Logger
}

func New(weightRepo repositories.WeightHistoryRepository, logger *slog.Logger) *Service {
	return &Service{
		weightRepo: weightRepo,
		logger:     logger.With("service", "weight"),
	}
}

// GetWeightHistory retrieves all weight entries for a user
func (s *Service) GetWeightHistory(userID int) ([]*models.WeightHistory, error) {
	s.logger.Debug("GetWeightHistory called", "user_id", userID)
	return s.weightRepo.GetByUserID(userID)
}

// GetWeightHistoryByDateRange retrieves weight entries for a user within a date range
// If no dates are provided, returns all weight history
func (s *Service) GetWeightHistoryByDateRange(userID int, dateFrom, dateTo string) ([]*models.WeightHistory, error) {
	s.logger.Debug("GetWeightHistoryByDateRange called",
		"user_id", userID,
		"date_from", dateFrom,
		"date_to", dateTo)

	// If dates are empty, return all weight history
	if dateFrom == "" && dateTo == "" {
		return s.weightRepo.GetByUserID(userID)
	}

	// If only one date is provided, default the other
	if dateFrom == "" {
		dateFrom = "1970-01-01"
	}
	if dateTo == "" {
		dateTo = time.Now().Format("2006-01-02")
	}

	return s.weightRepo.GetByUserIDAndDateRange(userID, dateFrom, dateTo)
}

// CreateWeightEntry adds a new weight entry
func (s *Service) CreateWeightEntry(userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error) {
	s.logger.Debug("CreateWeightEntry called",
		"user_id", userID,
		"weight", weight)

	return s.weightRepo.Create(userID, weight, recordedAt)
}

// UpdateWeightEntry updates an existing weight entry
func (s *Service) UpdateWeightEntry(id, userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error) {
	s.logger.Debug("UpdateWeightEntry called",
		"id", id,
		"user_id", userID,
		"weight", weight)

	return s.weightRepo.Update(id, userID, weight, recordedAt)
}

// DeleteWeightEntry deletes a weight entry
func (s *Service) DeleteWeightEntry(id, userID int) error {
	s.logger.Debug("DeleteWeightEntry called",
		"id", id,
		"user_id", userID)

	return s.weightRepo.Delete(id, userID)
}
