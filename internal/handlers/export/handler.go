package export

import (
	"fmt"
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/repositories"
	exportservice "ypeskov/kkal-tracker/internal/services/export"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	exportService *exportservice.Service
	userRepo      repositories.UserRepository
	logger        *slog.Logger
}

type Request struct {
	DateFrom     string `json:"date_from" validate:"required,datetime=2006-01-02"`
	DateTo       string `json:"date_to" validate:"required,datetime=2006-01-02"`
	DataType     string `json:"data_type" validate:"required,oneof=weight food both"`
	DeliveryType string `json:"delivery_type" validate:"required,oneof=download email"`
}

func New(exportService *exportservice.Service, userRepo repositories.UserRepository, logger *slog.Logger) *Handler {
	return &Handler{
		exportService: exportService,
		userRepo:      userRepo,
		logger:        logger.With("handler", "export"),
	}
}

// Export handles data export requests
func (h *Handler) Export(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("Export called", "user_id", userID)

	var req Request
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.Error("Validation failed", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get user info for email and language
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user info")
	}

	// Determine language
	language := "en_US"
	if user.Language != nil {
		language = *user.Language
	}

	// Build service request
	serviceReq := &exportservice.ExportRequest{
		UserID:       userID,
		DateFrom:     req.DateFrom,
		DateTo:       req.DateTo,
		DataType:     exportservice.ExportDataType(req.DataType),
		DeliveryType: exportservice.DeliveryType(req.DeliveryType),
		Language:     language,
		UserEmail:    user.Email,
	}

	// Call export service
	result, err := h.exportService.Export(serviceReq)
	if err != nil {
		h.logger.Error("Export failed", "error", err, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Export failed")
	}

	// If email delivery, return success message
	if req.DeliveryType == "email" {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Export sent to email",
		})
	}

	// For download, return the file
	c.Response().Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", result.FileName))

	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", result.FileBytes)
}

// RegisterRoutes registers all export-related routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.Export)
}
