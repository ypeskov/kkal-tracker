package calories

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/services"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	calorieService *services.CalorieService
	logger         *slog.Logger
}

type CreateEntryRequest struct {
	Food         string   `json:"food" validate:"required"`
	Calories     int      `json:"calories" validate:"required,min=1"`
	Weight       float64  `json:"weight" validate:"required,min=0.1"`
	KcalPer100g  float64  `json:"kcalPer100g" validate:"required,min=0.1"`
	Fats         *float64 `json:"fats,omitempty"`
	Carbs        *float64 `json:"carbs,omitempty"`
	Proteins     *float64 `json:"proteins,omitempty"`
	MealDatetime string   `json:"meal_datetime" validate:"required"`
}

func NewHandler(calorieService *services.CalorieService, logger *slog.Logger) *Handler {
	return &Handler{
		calorieService: calorieService,
		logger:         logger,
	}
}

func (h *Handler) GetEntries(c echo.Context) error {
	userID := c.Get("user_id").(int)
	date := c.QueryParam("date")
	dateFrom := c.QueryParam("dateFrom")
	dateTo := c.QueryParam("dateTo")

	var entries []*models.CalorieEntry
	var err error

	if dateFrom != "" && dateTo != "" {
		entries, err = h.calorieService.GetEntriesByDateRange(userID, dateFrom, dateTo)
	} else {
		entries, err = h.calorieService.GetEntriesByDate(userID, date)
	}

	if err != nil {
		h.logger.Error("Failed to get calorie entries", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, entries)
}

func (h *Handler) CreateEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var req CreateEntryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Parse the meal datetime
	mealDatetime, err := time.Parse(time.RFC3339, req.MealDatetime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid meal_datetime format. Use ISO 8601 format")
	}

	result, err := h.calorieService.CreateEntry(userID,
		req.Food, req.Calories, req.Weight, req.KcalPer100g,
		req.Fats, req.Carbs, req.Proteins, mealDatetime)
	if err != nil {
		h.logger.Error("Failed to create calorie entry", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) DeleteEntry(c echo.Context) error {
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

func (h *Handler) UpdateEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)

	idParam := c.Param("id")
	entryID, err := strconv.Atoi(idParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid entry ID")
	}

	var req CreateEntryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Parse the meal datetime
	mealDatetime, err := time.Parse(time.RFC3339, req.MealDatetime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid meal_datetime format. Use ISO 8601 format")
	}

	entry, err := h.calorieService.UpdateEntry(entryID, userID, req.Food, req.Calories, req.Weight, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins, mealDatetime)
	if err != nil {
		h.logger.Error("Failed to update calorie entry", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, entry)
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetEntries)
	g.POST("", h.CreateEntry)
	g.PUT("/:id", h.UpdateEntry)
	g.DELETE("/:id", h.DeleteEntry)
}