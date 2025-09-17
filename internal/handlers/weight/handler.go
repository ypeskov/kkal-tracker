package weight

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	weightservice "ypeskov/kkal-tracker/internal/services/weight"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	weightService *weightservice.Service
	logger        *slog.Logger
}

type CreateWeightRequest struct {
	Weight     float64 `json:"weight" validate:"required,min=1,max=500"`
	RecordedAt *string `json:"recorded_at,omitempty"`
}

type UpdateWeightRequest struct {
	Weight     float64 `json:"weight" validate:"required,min=1,max=500"`
	RecordedAt *string `json:"recorded_at,omitempty"`
}

func NewHandler(weightService *weightservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		weightService: weightService,
		logger:        logger.With("handler", "weight"),
	}
}

// GetWeightHistory returns all weight entries for the authenticated user
func (h *Handler) GetWeightHistory(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("GetWeightHistory called", "user_id", userID)

	history, err := h.weightService.GetWeightHistory(userID)
	if err != nil {
		h.logger.Error("Failed to get weight history", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get weight history")
	}

	return c.JSON(http.StatusOK, history)
}

// GetWeightHistoryByDateRange returns weight entries within a date range
func (h *Handler) GetWeightHistoryByDateRange(c echo.Context) error {
	userID := c.Get("user_id").(int)
	dateFrom := c.QueryParam("from")
	dateTo := c.QueryParam("to")

	h.logger.Debug("GetWeightHistoryByDateRange called",
		"user_id", userID,
		"from", dateFrom,
		"to", dateTo)

	history, err := h.weightService.GetWeightHistoryByDateRange(userID, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("Failed to get weight history by date range", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get weight history")
	}

	return c.JSON(http.StatusOK, history)
}

// CreateWeightEntry creates a new weight entry
func (h *Handler) CreateWeightEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var req CreateWeightRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.Error("Validation failed", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse the date string if provided
	var recordedAt *time.Time
	if req.RecordedAt != nil && *req.RecordedAt != "" {
		parsed, err := time.Parse("2006-01-02", *req.RecordedAt)
		if err != nil {
			h.logger.Error("Failed to parse recorded_at date", "error", err, "date", *req.RecordedAt)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Expected YYYY-MM-DD")
		}
		recordedAt = &parsed
	}

	h.logger.Debug("CreateWeightEntry called",
		"user_id", userID,
		"weight", req.Weight)

	entry, err := h.weightService.CreateWeightEntry(userID, req.Weight, recordedAt)
	if err != nil {
		h.logger.Error("Failed to create weight entry", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create weight entry")
	}

	return c.JSON(http.StatusCreated, entry)
}

// UpdateWeightEntry updates an existing weight entry
func (h *Handler) UpdateWeightEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid weight entry ID")
	}

	var req UpdateWeightRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.Error("Validation failed", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse the date string if provided
	var recordedAt *time.Time
	if req.RecordedAt != nil && *req.RecordedAt != "" {
		parsed, err := time.Parse("2006-01-02", *req.RecordedAt)
		if err != nil {
			h.logger.Error("Failed to parse recorded_at date", "error", err, "date", *req.RecordedAt)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Expected YYYY-MM-DD")
		}
		recordedAt = &parsed
	}

	h.logger.Debug("UpdateWeightEntry called",
		"id", id,
		"user_id", userID,
		"weight", req.Weight)

	entry, err := h.weightService.UpdateWeightEntry(id, userID, req.Weight, recordedAt)
	if err != nil {
		h.logger.Error("Failed to update weight entry", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update weight entry")
	}

	return c.JSON(http.StatusOK, entry)
}

// DeleteWeightEntry deletes a weight entry
func (h *Handler) DeleteWeightEntry(c echo.Context) error {
	userID := c.Get("user_id").(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid weight entry ID")
	}

	h.logger.Debug("DeleteWeightEntry called",
		"id", id,
		"user_id", userID)

	if err := h.weightService.DeleteWeightEntry(id, userID); err != nil {
		h.logger.Error("Failed to delete weight entry", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete weight entry")
	}

	return c.NoContent(http.StatusNoContent)
}

// RegisterRoutes registers all weight-related routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetWeightHistoryByDateRange)
	g.POST("", h.CreateWeightEntry)
	g.PUT("/:id", h.UpdateWeightEntry)
	g.DELETE("/:id", h.DeleteWeightEntry)
}