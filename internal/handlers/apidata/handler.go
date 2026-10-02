package apidata

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	ingredientservice "ypeskov/kkal-tracker/internal/services/ingredient"
	weightservice "ypeskov/kkal-tracker/internal/services/weight"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	calorieService    calorieservice.Servicer
	weightService     weightservice.Servicer
	ingredientService ingredientservice.Servicer
	logger            *slog.Logger
}

type DataResponse struct {
	Weight []WeightEntry `json:"weight,omitempty"`
	Food   []FoodEntry   `json:"food,omitempty"`
}

type WeightEntry struct {
	Weight     float64   `json:"weight"`
	RecordedAt time.Time `json:"recorded_at"`
}

type FoodEntry struct {
	ID           int       `json:"id"`
	Food         string    `json:"food"`
	Calories     int       `json:"calories"`
	Weight       float64   `json:"weight"`
	KcalPer100g  float64   `json:"kcal_per_100g"`
	Fats         *float64  `json:"fats,omitempty"`
	Carbs        *float64  `json:"carbs,omitempty"`
	Proteins     *float64  `json:"proteins,omitempty"`
	MealDatetime time.Time `json:"meal_datetime"`
}

func New(calorieService calorieservice.Servicer, weightService weightservice.Servicer,
	ingredientService ingredientservice.Servicer, logger *slog.Logger) *Handler {
	return &Handler{
		calorieService:    calorieService,
		weightService:     weightService,
		ingredientService: ingredientService,
		logger:            logger.With("handler", "apidata"),
	}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/data", h.GetData)
	g.GET("/ingredients", h.GetIngredients)
	g.POST("/food", h.CreateFood)
	g.PUT("/food/:id", h.UpdateFood)
	g.DELETE("/food/:id", h.DeleteFood)
}

func (h *Handler) GetData(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	dataType := c.QueryParam("type")
	dateFrom := c.QueryParam("from")
	dateTo := c.QueryParam("to")

	if dataType == "" {
		dataType = "both"
	}
	if dataType != "weight" && dataType != "food" && dataType != "both" {
		return echo.NewHTTPError(http.StatusBadRequest, "type must be weight, food, or both")
	}

	if dateFrom == "" || dateTo == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "from and to date parameters are required (YYYY-MM-DD)")
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", dateFrom); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid from date format, expected YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", dateTo); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid to date format, expected YYYY-MM-DD")
	}

	response := &DataResponse{}

	if dataType == "weight" || dataType == "both" {
		weightData, err := h.weightService.GetWeightHistoryByDateRange(userID, dateFrom, dateTo)
		if err != nil {
			h.logger.Error("Failed to fetch weight data", "error", err, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch data")
		}
		response.Weight = mapWeightEntries(weightData)
	}

	if dataType == "food" || dataType == "both" {
		foodData, err := h.calorieService.GetEntriesByDateRange(userID, dateFrom, dateTo)
		if err != nil {
			h.logger.Error("Failed to fetch food data", "error", err, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch data")
		}
		response.Food = mapFoodEntries(foodData)
	}

	return c.JSON(http.StatusOK, response)
}

// GetIngredients returns the whole ingredient list of the user: a client picks the ingredients
// of a meal from it and sends their IDs to CreateFood
func (h *Handler) GetIngredients(c *echo.Context) error {
	userID := c.Get("user_id").(int)

	ingredients, err := h.ingredientService.GetIngredientsWithUsage(userID)
	if err != nil {
		h.logger.Error("Failed to fetch ingredients", "error", err, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch data")
	}

	response := IngredientsResponse{Ingredients: make([]IngredientItem, len(ingredients))}
	for i, ingredient := range ingredients {
		response.Ingredients[i] = IngredientItem{
			IngredientInfo: mapIngredient(&ingredient.UserIngredient),
			TimesUsed:      ingredient.TimesUsed,
			LastUsed:       ingredient.LastUsed,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// CreateFood stores one meal. All items are saved together or not at all.
func (h *Handler) CreateFood(c *echo.Context) error {
	userID := c.Get("user_id").(int)

	var req CreateFoodRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	serviceReq := &calorieservice.CreateMealRequest{
		UserID: userID,
		Items:  make([]calorieservice.MealItem, len(req.Items)),
	}
	if req.MealDatetime != "" {
		mealDatetime, err := time.Parse(time.RFC3339, req.MealDatetime)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid meal_datetime format. Use ISO 8601 format")
		}
		serviceReq.MealDatetime = mealDatetime
	}
	for i, item := range req.Items {
		serviceItem := calorieservice.MealItem{
			IngredientID: item.IngredientID,
			OneOff:       item.OneOff,
			Weight:       item.Weight,
			KcalPer100g:  item.KcalPer100g,
		}
		if item.NewIngredient != nil {
			if item.NewIngredient.KcalPer100g == nil {
				return echo.NewHTTPError(http.StatusBadRequest,
					fmt.Sprintf("items[%d]: new_ingredient.kcal_per_100g is required", i))
			}
			serviceItem.NewIngredient = &calorieservice.NewIngredient{
				Name:        item.NewIngredient.Name,
				KcalPer100g: *item.NewIngredient.KcalPer100g,
				Fats:        item.NewIngredient.Fats,
				Carbs:       item.NewIngredient.Carbs,
				Proteins:    item.NewIngredient.Proteins,
			}
		}
		serviceReq.Items[i] = serviceItem
	}

	result, err := h.calorieService.CreateMeal(serviceReq)
	if err != nil {
		var itemErr *calorieservice.MealItemError
		var duplicateErr *calorieservice.DuplicateIngredientError
		switch {
		case errors.As(err, &duplicateErr):
			return c.JSON(http.StatusConflict, DuplicateIngredientResponse{
				Message:            duplicateErr.Error(),
				ItemIndex:          duplicateErr.Index,
				ExistingIngredient: mapIngredient(duplicateErr.Existing),
			})
		case errors.As(err, &itemErr):
			return echo.NewHTTPError(http.StatusBadRequest, itemErr.Error())
		case errors.Is(err, calorieservice.ErrEmptyMeal), errors.Is(err, calorieservice.ErrTooManyItems):
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		h.logger.Error("Failed to create meal", "error", err, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save food")
	}

	response := CreateFoodResponse{
		Entries:          make([]CreatedFoodEntry, len(result.Entries)),
		TotalCalories:    result.TotalCalories,
		Day:              result.Day,
		DayTotalCalories: result.DayTotalCalories,
	}
	for i, entry := range result.Entries {
		response.Entries[i] = CreatedFoodEntry{
			FoodEntry:         mapFoodEntry(entry.Entry),
			IngredientID:      entry.IngredientID,
			IngredientCreated: entry.IngredientCreated,
		}
	}

	return c.JSON(http.StatusCreated, response)
}

// UpdateFood corrects a stored diary entry: its weight, ingredient, calorie value or time
func (h *Handler) UpdateFood(c *echo.Context) error {
	userID := c.Get("user_id").(int)

	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid entry ID")
	}

	var req UpdateFoodRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	serviceReq := &calorieservice.UpdateMealEntryRequest{
		EntryID:      entryID,
		UserID:       userID,
		Weight:       req.Weight,
		IngredientID: req.IngredientID,
		KcalPer100g:  req.KcalPer100g,
	}
	if req.MealDatetime != "" {
		mealDatetime, err := time.Parse(time.RFC3339, req.MealDatetime)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid meal_datetime format. Use ISO 8601 format")
		}
		serviceReq.MealDatetime = &mealDatetime
	}

	result, err := h.calorieService.UpdateMealEntry(serviceReq)
	if err != nil {
		var validationErr *calorieservice.EntryValidationError
		switch {
		case errors.As(err, &validationErr):
			return echo.NewHTTPError(http.StatusBadRequest, validationErr.Error())
		case errors.Is(err, calorieservice.ErrEntryNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "Food entry not found")
		}
		h.logger.Error("Failed to update food entry", "error", err, "user_id", userID, "entry_id", entryID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update food")
	}

	return c.JSON(http.StatusOK, UpdateFoodResponse{
		Entry:            mapFoodEntry(result.Entry),
		Day:              result.Day,
		DayTotalCalories: result.DayTotalCalories,
	})
}

// DeleteFood removes a diary entry, e.g. one that was just stored by mistake
func (h *Handler) DeleteFood(c *echo.Context) error {
	userID := c.Get("user_id").(int)

	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid entry ID")
	}

	if err := h.calorieService.DeleteEntry(entryID, userID); err != nil {
		if errors.Is(err, calorieservice.ErrEntryNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Food entry not found")
		}
		h.logger.Error("Failed to delete food entry", "error", err, "user_id", userID, "entry_id", entryID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete food")
	}

	return c.NoContent(http.StatusNoContent)
}

func mapIngredient(ingredient *models.UserIngredient) IngredientInfo {
	return IngredientInfo{
		ID:          ingredient.ID,
		Name:        ingredient.Name,
		KcalPer100g: ingredient.KcalPer100g,
		Fats:        ingredient.Fats,
		Carbs:       ingredient.Carbs,
		Proteins:    ingredient.Proteins,
	}
}

func mapWeightEntries(data []*models.WeightHistory) []WeightEntry {
	entries := make([]WeightEntry, len(data))
	for i, w := range data {
		entries[i] = WeightEntry{
			Weight:     w.Weight,
			RecordedAt: w.RecordedAt,
		}
	}
	return entries
}

func mapFoodEntries(data []*models.CalorieEntry) []FoodEntry {
	entries := make([]FoodEntry, len(data))
	for i, e := range data {
		entries[i] = mapFoodEntry(e)
	}
	return entries
}

func mapFoodEntry(e *models.CalorieEntry) FoodEntry {
	return FoodEntry{
		ID:           e.ID,
		Food:         e.Food,
		Calories:     e.Calories,
		Weight:       e.Weight,
		KcalPer100g:  e.KcalPer100g,
		Fats:         e.Fats,
		Carbs:        e.Carbs,
		Proteins:     e.Proteins,
		MealDatetime: e.MealDatetime,
	}
}
