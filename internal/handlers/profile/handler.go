package profile

import (
	"log/slog"
	"net/http"

	profileservice "ypeskov/kkal-tracker/internal/services/profile"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	profileService *profileservice.Service
	logger         *slog.Logger
}

func NewProfileHandler(profileService *profileservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		profileService: profileService,
		logger:         logger,
	}
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("GetProfile called", "user_id", userID)

	// Service returns DTO directly
	profile, err := h.profileService.GetProfile(userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user profile")
	}

	h.logger.Debug("GetProfile returning profile", "user_id", userID, "email", profile.Email)
	return c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("UpdateProfile called", "user_id", userID)

	// Bind to DTO which has validation tags
	var req profileservice.ProfileUpdateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Debug("UpdateProfile failed - invalid request body", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.logger.Debug("UpdateProfile failed - validation error", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get current profile to preserve email (email is used as login and cannot be changed by users)
	currentProfile, err := h.profileService.GetProfile(userID)
	if err != nil {
		h.logger.Error("Failed to get current profile", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get current profile")
	}

	// Always use current email - users cannot change their login email
	// Service layer still supports email updates for potential future admin functionality
	req.Email = currentProfile.Email

	h.logger.Debug("UpdateProfile request", "user_id", userID, "email", req.Email, "first_name", req.FirstName, "last_name", req.LastName)

	// Pass DTO directly to service - service handles conversion
	if err := h.profileService.UpdateProfile(userID, &req); err != nil {
		h.logger.Error("Failed to update profile", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}

	// Return updated profile
	h.logger.Debug("UpdateProfile successful", "user_id", userID)
	return h.GetProfile(c)
}

// RegisterRoutes registers the profile routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/profile", h.GetProfile)
	g.PUT("/profile", h.UpdateProfile)
}