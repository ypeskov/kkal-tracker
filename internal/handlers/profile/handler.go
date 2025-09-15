package profile

import (
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/dto"
	"ypeskov/kkal-tracker/internal/services"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	profileService *services.ProfileService
	logger         *slog.Logger
}

func NewProfileHandler(profileService *services.ProfileService, logger *slog.Logger) *Handler {
	return &Handler{
		profileService: profileService,
		logger:         logger,
	}
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)

	// Service returns DTO directly
	profile, err := h.profileService.GetProfile(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user profile")
	}

	return c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)

	// Bind to DTO which has validation tags
	var req dto.ProfileUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Pass DTO directly to service - service handles conversion
	if err := h.profileService.UpdateProfile(userID, &req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}

	// Return updated profile
	return h.GetProfile(c)
}

// RegisterRoutes registers the profile routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/profile", h.GetProfile)
	g.PUT("/profile", h.UpdateProfile)
}