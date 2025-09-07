package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"ypeskov/kkal-tracker/internal/services"

	"github.com/labstack/echo/v4"
)

type CalorieHandler struct {
	calorieService *services.CalorieService
	logger         *slog.Logger
}

type CreateCalorieEntryRequest struct {
	Food         string   `json:"food" validate:"required"`
	Calories     int      `json:"calories" validate:"required,min=1"`
	Weight       float64  `json:"weight" validate:"required,min=0.1"`
	KcalPer100g  float64  `json:"kcalPer100g" validate:"required,min=0.1"`
	Fats         *float64 `json:"fats,omitempty"`
	Carbs        *float64 `json:"carbs,omitempty"`
	Proteins     *float64 `json:"proteins,omitempty"`
	MealDatetime string   `json:"meal_datetime" validate:"required"`
}

func NewCalorieHandler(calorieService *services.CalorieService, logger *slog.Logger) *CalorieHandler {
	return &CalorieHandler{
		calorieService: calorieService,
		logger:         logger,
	}
}

func (h *CalorieHandler) GetEntries(c echo.Context) error {
	userID := c.Get("user_id").(int)
	date := c.QueryParam("date")

	entries, err := h.calorieService.GetEntriesByDate(userID, date)
	if err != nil {
		h.logger.Error("Failed to get calorie entries", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, entries)
}

func (h *CalorieHandler) CreateEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var req CreateCalorieEntryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Parse the meal datetime
	mealDatetime, err := time.Parse(time.RFC3339, req.MealDatetime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid meal_datetime format. Use ISO 8601 format")
	}

	entry, err := h.calorieService.CreateEntry(userID, req.Food, req.Calories, req.Weight, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins, mealDatetime)
	if err != nil {
		h.logger.Error("Failed to create calorie entry", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusCreated, entry)
}

func (h *CalorieHandler) DeleteEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)
	
	idParam := c.Param("id")
	entryID, err := strconv.Atoi(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid entry ID")
	}

	err = h.calorieService.DeleteEntry(entryID, userID)
	if err != nil {
		h.logger.Error("Failed to delete calorie entry", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *CalorieHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetEntries)
	g.POST("", h.CreateEntry)
	g.DELETE("/:id", h.DeleteEntry)
}