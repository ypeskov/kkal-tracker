package services

import (
	"errors"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

var (
	ErrInvalidDate   = errors.New("invalid date format")
	ErrEntryNotFound = errors.New("calorie entry not found")
)

type CalorieService struct {
	calorieRepo    *models.CalorieEntryRepository
	ingredientRepo *models.IngredientRepository
	logger         *slog.Logger
}

func NewCalorieService(calorieRepo *models.CalorieEntryRepository, ingredientRepo *models.IngredientRepository, logger *slog.Logger) *CalorieService {
	return &CalorieService{
		calorieRepo:    calorieRepo,
		ingredientRepo: ingredientRepo,
		logger:         logger,
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

type CreateEntryResult struct {
	Entry                *models.CalorieEntry `json:"entry"`
	NewIngredientCreated bool                 `json:"new_ingredient_created"`
}

func (s *CalorieService) CreateEntry(userID int,
	food string, calories int, weight float64, kcalPer100g float64,
	fats, carbs, proteins *float64, mealDatetime time.Time) (*CreateEntryResult, error) {

	// Validate calories
	if calories <= 0 {
		return nil, errors.New("calories must be greater than 0")
	}

	// Validate weight
	if weight <= 0 {
		return nil, errors.New("weight must be greater than 0")
	}

	// Validate kcal per 100g
	if kcalPer100g <= 0 {
		return nil, errors.New("kcal per 100g must be greater than 0")
	}

	// Validate food name
	if food == "" {
		return nil, errors.New("food name is required")
	}

	newIngredientCreated := false

	// Check if ingredient exists, if not create it
	_, err := s.ingredientRepo.GetUserIngredientByName(userID, food)
	if err != nil {
		// Ingredient doesn't exist, create it
		_, createErr := s.ingredientRepo.CreateOrUpdateUserIngredient(userID, food, kcalPer100g, fats, carbs, proteins)
		if createErr != nil {
			s.logger.Error("Failed to create user ingredient", "error", createErr, "user_id", userID, "food", food)
			// Don't fail the calorie entry creation if ingredient creation fails
		} else {
			s.logger.Info("New user ingredient created", "user_id", userID, "food", food, "kcalPer100g", kcalPer100g)
			newIngredientCreated = true
		}
	}

	entry, err := s.calorieRepo.Create(userID, food, calories, weight, kcalPer100g, fats, carbs, proteins, mealDatetime)
	if err != nil {
		s.logger.Error("Failed to create calorie entry", "error", err, "user_id", userID)
		return nil, err
	}

	s.logger.Info("Calorie entry created", "user_id", userID, "food", food, "calories", calories, "weight", weight, "kcalPer100g", kcalPer100g, "meal_datetime", mealDatetime)

	return &CreateEntryResult{
		Entry:                entry,
		NewIngredientCreated: newIngredientCreated,
	}, nil
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

func (s *CalorieService) GetEntriesByDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error) {
	entries, err := s.calorieRepo.GetByUserIDAndDateRange(userID, dateFrom, dateTo)
	if err != nil {
		s.logger.Error("Failed to get calorie entries by date range", "error", err, "user_id", userID, "date_from", dateFrom, "date_to", dateTo)
		return nil, err
	}

	// Return empty slice instead of nil for better API responses
	if entries == nil {
		entries = []*models.CalorieEntry{}
	}

	return entries, nil
}

func (s *CalorieService) UpdateEntry(entryID, userID int, food string, calories int, weight float64, kcalPer100g float64, fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error) {
	// Validate calories
	if calories <= 0 {
		return nil, errors.New("calories must be greater than 0")
	}

	// Validate weight
	if weight <= 0 {
		return nil, errors.New("weight must be greater than 0")
	}

	// Validate kcal per 100g
	if kcalPer100g <= 0 {
		return nil, errors.New("kcal per 100g must be greater than 0")
	}

	// Validate food name
	if food == "" {
		return nil, errors.New("food name is required")
	}

	entry, err := s.calorieRepo.Update(entryID, userID, food, calories, weight, kcalPer100g, fats, carbs, proteins, mealDatetime)
	if err != nil {
		s.logger.Error("Failed to update calorie entry", "error", err, "entry_id", entryID, "user_id", userID)
		return nil, err
	}

	s.logger.Info("Calorie entry updated", "entry_id", entryID, "user_id", userID, "food", food, "calories", calories, "weight", weight, "kcalPer100g", kcalPer100g, "meal_datetime", mealDatetime)
	return entry, nil
}
