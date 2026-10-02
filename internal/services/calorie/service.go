package calorie

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

const (
	// MaxMealItems limits the number of foods in one meal request
	MaxMealItems = 50

	// Sanity limits for meal items: they catch unit mix-ups (kilograms instead of grams,
	// calories of the whole portion instead of calories per 100 g). Pure fat has 900 kcal per 100 g.
	maxItemWeight      = 10000.0
	maxKcalPer100g     = 900.0
	maxNutrientPer100g = 100.0
	maxFoodNameLength  = 200
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
		logger:         logger.With("service", "calorie"),
	}
}

func (s *Service) CreateEntry(req *CreateEntryRequest) (*CreateEntryResult, error) {
	s.logger.Debug("CreateEntry called", "user_id", req.UserID, "food", req.Food, "calories", req.Calories, "weight", req.Weight)

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

	result := &CreateEntryResult{
		Entry:                entry,
		NewIngredientCreated: newIngredientCreated,
	}

	s.logger.Debug("CreateEntry completed successfully", "user_id", req.UserID, "entry_id", entry.ID)
	return result, nil
}

// CreateMeal stores several foods eaten at once. Every item refers to an existing ingredient by ID
// or describes a new one explicitly: nothing is matched or created by name behind the caller's back.
// The meal is atomic: if any item is invalid, nothing is saved.
func (s *Service) CreateMeal(req *CreateMealRequest) (*CreateMealResult, error) {
	s.logger.Debug("CreateMeal called", "user_id", req.UserID, "items", len(req.Items))

	if len(req.Items) == 0 {
		return nil, ErrEmptyMeal
	}
	if len(req.Items) > MaxMealItems {
		return nil, ErrTooManyItems
	}

	ingredients, err := s.ingredientRepo.GetAllUserIngredients(req.UserID)
	if err != nil {
		s.logger.Error("Failed to get user ingredients", "error", err, "user_id", req.UserID)
		return nil, err
	}
	ingredientsByID := make(map[int]*models.UserIngredient, len(ingredients))
	ingredientsByName := make(map[string]*models.UserIngredient, len(ingredients))
	for _, ingredient := range ingredients {
		ingredientsByID[ingredient.ID] = ingredient
		ingredientsByName[normalizeFoodName(ingredient.Name)] = ingredient
	}

	entries := make([]models.NewMealEntry, 0, len(req.Items))
	ingredientIDs := make([]*int, len(req.Items))
	// Normalized names of the new ingredients of this request -> index of the item that introduced them
	newNames := make(map[string]int)

	for i, item := range req.Items {
		if item.Weight <= 0 || item.Weight > maxItemWeight {
			return nil, itemError(i, "weight must be greater than 0 and at most %g grams", maxItemWeight)
		}

		var entry models.NewMealEntry
		switch {
		case item.IngredientID != nil && item.NewIngredient != nil:
			return nil, itemError(i, "set either ingredient_id or new_ingredient, not both")

		case item.IngredientID != nil:
			if item.OneOff {
				return nil, itemError(i, "one_off is only allowed together with new_ingredient")
			}
			ingredient, ok := ingredientsByID[*item.IngredientID]
			if !ok {
				return nil, itemError(i, "ingredient %d not found", *item.IngredientID)
			}
			kcalPer100g := ingredient.KcalPer100g
			if item.KcalPer100g != nil {
				if *item.KcalPer100g < 0 || *item.KcalPer100g > maxKcalPer100g {
					return nil, itemError(i, "kcal_per_100g must be between 0 and %g", maxKcalPer100g)
				}
				kcalPer100g = *item.KcalPer100g
			} else if kcalPer100g < 0 || kcalPer100g > maxKcalPer100g {
				// The ingredient list accepts any value, the diary must not get a nonsensical calorie count from it
				return nil, itemError(i, "ingredient %q has %g kcal per 100 g, expected a value between 0 and %g; send kcal_per_100g to override it",
					ingredient.Name, kcalPer100g, maxKcalPer100g)
			}
			ingredientIDs[i] = &ingredient.ID
			entry = models.NewMealEntry{
				Food:        ingredient.Name,
				KcalPer100g: kcalPer100g,
				Fats:        ingredient.Fats,
				Carbs:       ingredient.Carbs,
				Proteins:    ingredient.Proteins,
			}

		case item.NewIngredient != nil:
			if item.KcalPer100g != nil {
				return nil, itemError(i, "kcal_per_100g of a new food belongs inside new_ingredient")
			}
			food := item.NewIngredient
			name := cleanFoodName(food.Name)
			if name == "" {
				return nil, itemError(i, "new_ingredient.name is required")
			}
			if utf8.RuneCountInString(name) > maxFoodNameLength {
				return nil, itemError(i, "new_ingredient.name must be at most %d characters", maxFoodNameLength)
			}
			if food.KcalPer100g < 0 || food.KcalPer100g > maxKcalPer100g {
				return nil, itemError(i, "new_ingredient.kcal_per_100g must be between 0 and %g", maxKcalPer100g)
			}
			nutrients := []struct {
				field string
				value *float64
			}{{"fats", food.Fats}, {"carbs", food.Carbs}, {"proteins", food.Proteins}}
			for _, n := range nutrients {
				if n.value != nil && (*n.value < 0 || *n.value > maxNutrientPer100g) {
					return nil, itemError(i, "new_ingredient.%s must be between 0 and %g (grams per 100 g)", n.field, maxNutrientPer100g)
				}
			}

			if !item.OneOff {
				normalized := normalizeFoodName(name)
				if existing, ok := ingredientsByName[normalized]; ok {
					return nil, &DuplicateIngredientError{Index: i, Existing: existing}
				}
				if first, ok := newNames[normalized]; ok {
					return nil, itemError(i, "the same new ingredient is already listed in items[%d], merge them into one item", first)
				}
				newNames[normalized] = i
			}
			entry = models.NewMealEntry{
				Food:             name,
				KcalPer100g:      food.KcalPer100g,
				Fats:             food.Fats,
				Carbs:            food.Carbs,
				Proteins:         food.Proteins,
				SaveAsIngredient: !item.OneOff,
			}

		default:
			return nil, itemError(i, "ingredient_id or new_ingredient is required")
		}

		entry.Weight = item.Weight
		entry.Calories = int(math.Round(item.Weight * entry.KcalPer100g / 100))
		entries = append(entries, entry)
	}

	// The diary keeps meal times in UTC, days are compared as UTC dates
	mealDatetime := req.MealDatetime.UTC()
	if req.MealDatetime.IsZero() {
		mealDatetime = time.Now().UTC()
	}

	created, err := s.calorieRepo.CreateMeal(req.UserID, entries, mealDatetime)
	if err != nil {
		s.logger.Error("Failed to create meal", "error", err, "user_id", req.UserID)
		return nil, err
	}

	result := &CreateMealResult{
		Entries: make([]MealEntryResult, len(created)),
		Day:     mealDatetime.Format("2006-01-02"),
	}
	for i, c := range created {
		entryResult := MealEntryResult{Entry: c.Entry, IngredientID: ingredientIDs[i]}
		if c.Ingredient != nil {
			entryResult.IngredientID = &c.Ingredient.ID
			entryResult.IngredientCreated = true
			s.logger.Info("New user ingredient created", "user_id", req.UserID, "food", c.Ingredient.Name, "kcalPer100g", c.Ingredient.KcalPer100g)
		}
		result.Entries[i] = entryResult
		result.TotalCalories += c.Entry.Calories
	}

	s.logger.Info("Meal created", "user_id", req.UserID, "entries", len(created), "calories", result.TotalCalories, "meal_datetime", mealDatetime)

	// The meal is already saved: a failure here must not turn the request into an error,
	// otherwise the client would retry and store the meal twice
	if dayTotal, err := s.GetTotalCaloriesForDate(req.UserID, result.Day); err == nil {
		result.DayTotalCalories = &dayTotal
	}

	s.logger.Debug("CreateMeal completed successfully", "user_id", req.UserID, "entries", len(created))
	return result, nil
}

// cleanFoodName drops invisible characters (control and format ones, e.g. a zero-width space)
// and collapses whitespace, so that two names that look the same are the same
func cleanFoodName(name string) string {
	visible := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, name)
	return strings.Join(strings.Fields(visible), " ")
}

// normalizeFoodName makes food names comparable: case, extra spaces and the Russian "ё" are ignored
func normalizeFoodName(name string) string {
	return strings.ReplaceAll(strings.ToLower(cleanFoodName(name)), "ё", "е")
}

func itemError(index int, format string, args ...any) *MealItemError {
	return &MealItemError{Index: index, Message: fmt.Sprintf(format, args...)}
}

func (s *Service) DeleteEntry(entryID, userID int) error {
	s.logger.Debug("DeleteEntry called", "entry_id", entryID, "user_id", userID)

	err := s.calorieRepo.Delete(entryID, userID)
	if errors.Is(err, repositories.ErrNotFound) {
		s.logger.Debug("DeleteEntry failed - entry not found", "entry_id", entryID, "user_id", userID)
		return ErrEntryNotFound
	}
	if err != nil {
		s.logger.Error("Failed to delete calorie entry", "error", err, "entry_id", entryID, "user_id", userID)
		return err
	}

	s.logger.Info("Calorie entry deleted", "entry_id", entryID, "user_id", userID)
	s.logger.Debug("DeleteEntry completed successfully", "entry_id", entryID, "user_id", userID)
	return nil
}

func (s *Service) GetTotalCaloriesForDate(userID int, date string) (int, error) {
	s.logger.Debug("GetTotalCaloriesForDate called", "user_id", userID, "date", date)

	// Use date range with same date for both from and to
	entries, err := s.GetEntriesByDateRange(userID, date, date)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, entry := range entries {
		total += entry.Calories
	}

	s.logger.Debug("GetTotalCaloriesForDate completed successfully", "user_id", userID, "date", date, "total", total)
	return total, nil
}

func (s *Service) GetWeeklyStats(userID int, startDate string) (map[string]int, error) {
	s.logger.Debug("GetWeeklyStats called", "user_id", userID, "start_date", startDate)

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

	s.logger.Debug("GetWeeklyStats completed successfully", "user_id", userID, "start_date", startDate, "days_count", len(stats))
	return stats, nil
}

func (s *Service) GetEntriesByDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error) {
	s.logger.Debug("GetEntriesByDateRange called", "user_id", userID, "date_from", dateFrom, "date_to", dateTo)

	entries, err := s.calorieRepo.GetByUserIDAndDateRange(userID, dateFrom, dateTo)
	if err != nil {
		s.logger.Error("failed to get calorie entries by date range", "error", err, "user_id", userID, "date_from", dateFrom, "date_to", dateTo)
		return nil, err
	}

	// Return empty slice instead of nil for better API responses
	if entries == nil {
		entries = []*models.CalorieEntry{}
	}

	s.logger.Debug("GetEntriesByDateRange completed successfully", "user_id", userID, "date_from", dateFrom, "date_to", dateTo, "count", len(entries))
	return entries, nil
}

func (s *Service) UpdateEntry(req *UpdateEntryRequest) (*models.CalorieEntry, error) {
	s.logger.Debug("UpdateEntry called", "entry_id", req.EntryID, "user_id", req.UserID, "food", req.Food, "calories", req.Calories, "weight", req.Weight)

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
	s.logger.Debug("UpdateEntry completed successfully", "entry_id", req.EntryID, "user_id", req.UserID)
	return entry, nil
}
