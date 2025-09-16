package calorie

import (
	"errors"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

type Service struct {
	calorieRepo    repositories.CalorieEntryRepository
	ingredientRepo repositories.IngredientRepository
	logger         *slog.Logger
}

func New(calorieRepo repositories.CalorieEntryRepository,
	ingredientRepo repositories.IngredientRepository,
	logger *slog.Logger) *Service {
	return &Service{
		calorieRepo:    calorieRepo,
		ingredientRepo: ingredientRepo,
		logger:         logger,
	}
}

func (s *Service) CreateEntry(req *CreateEntryRequest) (*CreateEntryResult, error) {
	// Validate calories
	if req.Calories <= 0 {
		return nil, errors.New("calories must be greater than 0")
	}

	// Validate weight
	if req.Weight <= 0 {
		return nil, errors.New("weight must be greater than 0")
	}

	// Validate kcal per 100g
	if req.KcalPer100g <= 0 {
		return nil, errors.New("kcal per 100g must be greater than 0")
	}

	// Validate food name
	if req.Food == "" {
		return nil, errors.New("food name is required")
	}

	newIngredientCreated := false

	// Check if ingredient exists, if not create it
	_, err := s.ingredientRepo.GetUserIngredientByName(req.UserID, req.Food)
	if err != nil {
		// Ingredient doesn't exist, create it
		_, createErr := s.ingredientRepo.CreateOrUpdateUserIngredient(req.UserID, req.Food, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins)
		if createErr != nil {
			s.logger.Error("Failed to create user ingredient", "error", createErr, "user_id", req.UserID, "food", req.Food)
			// Don't fail the calorie entry creation if ingredient creation fails
		} else {
			s.logger.Info("New user ingredient created", "user_id", req.UserID, "food", req.Food, "kcalPer100g", req.KcalPer100g)
			newIngredientCreated = true
		}
	}

	entry, err := s.calorieRepo.Create(req.UserID, req.Food, req.Calories, req.Weight, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins, req.MealDatetime)
	if err != nil {
		s.logger.Error("Failed to create calorie entry", "error", err, "user_id", req.UserID)
		return nil, err
	}

	s.logger.Info("Calorie entry created", "user_id", req.UserID, "food", req.Food, "calories", req.Calories, "weight", req.Weight, "kcalPer100g", req.KcalPer100g, "meal_datetime", req.MealDatetime)

	return &CreateEntryResult{
		Entry:                entry,
		NewIngredientCreated: newIngredientCreated,
	}, nil
}

func (s *Service) DeleteEntry(entryID, userID int) error {
	err := s.calorieRepo.Delete(entryID, userID)
	if err != nil {
		s.logger.Error("Failed to delete calorie entry", "error", err, "entry_id", entryID, "user_id", userID)
		return err
	}

	s.logger.Info("Calorie entry deleted", "entry_id", entryID, "user_id", userID)
	return nil
}

func (s *Service) GetTotalCaloriesForDate(userID int, date string) (int, error) {
	// Use date range with same date for both from and to
	entries, err := s.GetEntriesByDateRange(userID, date, date)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, entry := range entries {
		total += entry.Calories
	}

	return total, nil
}

func (s *Service) GetWeeklyStats(userID int, startDate string) (map[string]int, error) {
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

func (s *Service) GetEntriesByDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error) {
	entries, err := s.calorieRepo.GetByUserIDAndDateRange(userID, dateFrom, dateTo)
	if err != nil {
		s.logger.Error("failed to get calorie entries by date range", "error", err, "user_id", userID, "date_from", dateFrom, "date_to", dateTo)
		return nil, err
	}

	// Return empty slice instead of nil for better API responses
	if entries == nil {
		entries = []*models.CalorieEntry{}
	}

	return entries, nil
}

func (s *Service) UpdateEntry(req *UpdateEntryRequest) (*models.CalorieEntry, error) {
	// Validate calories
	if req.Calories <= 0 {
		return nil, errors.New("calories must be greater than 0")
	}

	// Validate weight
	if req.Weight <= 0 {
		return nil, errors.New("weight must be greater than 0")
	}

	// Validate kcal per 100g
	if req.KcalPer100g <= 0 {
		return nil, errors.New("kcal per 100g must be greater than 0")
	}

	// Validate food name
	if req.Food == "" {
		return nil, errors.New("food name is required")
	}

	entry, err := s.calorieRepo.Update(req.EntryID, req.UserID, req.Food, req.Calories, req.Weight, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins, req.MealDatetime)
	if err != nil {
		s.logger.Error("Failed to update calorie entry", "error", err, "entry_id", req.EntryID, "user_id", req.UserID)
		return nil, err
	}

	s.logger.Info("Calorie entry updated", "entry_id", req.EntryID, "user_id", req.UserID, "food", req.Food, "calories", req.Calories, "weight", req.Weight, "kcalPer100g", req.KcalPer100g, "meal_datetime", req.MealDatetime)
	return entry, nil
}
