package apidata

import (
	"log/slog"
	"net/http"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	weightservice "ypeskov/kkal-tracker/internal/services/weight"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	calorieService *calorieservice.Service
	weightService  *weightservice.Service
	logger         *slog.Logger
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
	Food         string    `json:"food"`
	Calories     int       `json:"calories"`
	Weight       float64   `json:"weight"`
	KcalPer100g  float64   `json:"kcal_per_100g"`
	Fats         *float64  `json:"fats,omitempty"`
	Carbs        *float64  `json:"carbs,omitempty"`
	Proteins     *float64  `json:"proteins,omitempty"`
	MealDatetime time.Time `json:"meal_datetime"`
}

func New(calorieService *calorieservice.Service, weightService *weightservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		calorieService: calorieService,
		weightService:  weightService,
		logger:         logger.With("handler", "apidata"),
	}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/data", h.GetData)
}

func (h *Handler) GetData(c echo.Context) error {
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
		entries[i] = FoodEntry{
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
	return entries
}
