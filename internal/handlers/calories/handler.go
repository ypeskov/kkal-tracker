package calories

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	calorieService *calorieservice.Service
	logger         *slog.Logger
}

func New(calorieService *calorieservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		calorieService: calorieService,
		logger:         logger.With("handler", "calories"),
	}
}

func (h *Handler) GetEntries(c echo.Context) error {
	userID := c.Get("user_id").(int)
	dateFrom := c.QueryParam("dateFrom")
	dateTo := c.QueryParam("dateTo")
	h.logger.Debug("GetEntries called", "user_id", userID, "dateFrom", dateFrom, "dateTo", dateTo)

	var entries []*models.CalorieEntry
	var err error

	if dateFrom != "" && dateTo != "" {
		entries, err = h.calorieService.GetEntriesByDateRange(userID, dateFrom, dateTo)
	} else {
		// If no date parameters provided, return empty array
		entries = []*models.CalorieEntry{}
	}

	if err != nil {
		h.logger.Error("Failed to get calorie entries", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("GetEntries returning entries", "count", len(entries))
	return c.JSON(http.StatusOK, entries)
}

func (h *Handler) CreateEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var req CreateEntryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse the meal datetime
	mealDatetime, err := time.Parse(time.RFC3339, req.MealDatetime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid meal_datetime format. Use ISO 8601 format")
	}

	serviceReq := &calorieservice.CreateEntryRequest{
		UserID:       userID,
		Food:         req.Food,
		Calories:     req.Calories,
		Weight:       req.Weight,
		KcalPer100g:  req.KcalPer100g,
		Fats:         req.Fats,
		Carbs:        req.Carbs,
		Proteins:     req.Proteins,
		MealDatetime: mealDatetime,
	}

	result, err := h.calorieService.CreateEntry(serviceReq)
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
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse the meal datetime
	mealDatetime, err := time.Parse(time.RFC3339, req.MealDatetime)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid meal_datetime format. Use ISO 8601 format")
	}

	serviceReq := &calorieservice.UpdateEntryRequest{
		EntryID:      entryID,
		UserID:       userID,
		Food:         req.Food,
		Calories:     req.Calories,
		Weight:       req.Weight,
		KcalPer100g:  req.KcalPer100g,
		Fats:         req.Fats,
		Carbs:        req.Carbs,
		Proteins:     req.Proteins,
		MealDatetime: mealDatetime,
	}

	entry, err := h.calorieService.UpdateEntry(serviceReq)
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
