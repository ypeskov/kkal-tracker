package services

import (
	"errors"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

var (
	ErrInvalidDate = errors.New("invalid date format")
	ErrEntryNotFound = errors.New("calorie entry not found")
)

type CalorieService struct {
	calorieRepo *models.CalorieEntryRepository
	logger      *slog.Logger
}

func NewCalorieService(calorieRepo *models.CalorieEntryRepository, logger *slog.Logger) *CalorieService {
	return &CalorieService{
		calorieRepo: calorieRepo,
		logger:      logger,
	}
}

func (s *CalorieService) GetEntriesByDate(userID int, date string) ([]*models.CalorieEntry, error) {
	// If no date provided, use today
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	entries, err := s.calorieRepo.GetByUserIDAndDate(userID, date)
	if err != nil {
		s.logger.Error("Failed to get calorie entries", "error", err, "user_id", userID, "date", date)
		return nil, err
	}

	// Return empty slice instead of nil for better API responses
	if entries == nil {
		entries = []*models.CalorieEntry{}
	}

	return entries, nil
}

func (s *CalorieService) CreateEntry(userID int, food string, calories int, date string) (*models.CalorieEntry, error) {
	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, ErrInvalidDate
	}

	// Validate calories
	if calories <= 0 {
		return nil, errors.New("calories must be greater than 0")
	}

	// Validate food name
	if food == "" {
		return nil, errors.New("food name is required")
	}

	entry, err := s.calorieRepo.Create(userID, food, calories, date)
	if err != nil {
		s.logger.Error("Failed to create calorie entry", "error", err, "user_id", userID)
		return nil, err
	}

	s.logger.Info("Calorie entry created", "user_id", userID, "food", food, "calories", calories, "date", date)
	return entry, nil
}

func (s *CalorieService) DeleteEntry(entryID, userID int) error {
	err := s.calorieRepo.Delete(entryID, userID)
	if err != nil {
		s.logger.Error("Failed to delete calorie entry", "error", err, "entry_id", entryID, "user_id", userID)
		return err
	}

	s.logger.Info("Calorie entry deleted", "entry_id", entryID, "user_id", userID)
	return nil
}

func (s *CalorieService) GetTotalCaloriesForDate(userID int, date string) (int, error) {
	entries, err := s.GetEntriesByDate(userID, date)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, entry := range entries {
		total += entry.Calories
	}

	return total, nil
}

func (s *CalorieService) GetWeeklyStats(userID int, startDate string) (map[string]int, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, ErrInvalidDate
	}

	stats := make(map[string]int)
	
	for i := 0; i < 7; i++ {
		currentDate := start.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		
		total, err := s.GetTotalCaloriesForDate(userID, dateStr)
		if err != nil {
			return nil, err
		}
		
		stats[dateStr] = total
	}

	return stats, nil
}