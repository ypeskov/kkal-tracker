package profile

import (
	"database/sql"
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories/sqlite"

	"github.com/labstack/echo/v4"
)

type ProfileHandler struct {
	db       *sql.DB
	userRepo *sqlite.UserRepository
	logger   *slog.Logger
}

func NewProfileHandler(db *sql.DB, logger *slog.Logger) *ProfileHandler {
	return &ProfileHandler{
		db:       db,
		userRepo: sqlite.NewUserRepository(db, logger),
		logger:   logger,
	}
}

// GetProfile returns the current user's profile
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("Getting profile", slog.Int("user_id", userID))

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user profile")
	}

	// Get latest weight
	weight, err := h.userRepo.GetLatestWeight(userID)
	if err != nil {
		h.logger.Error("Failed to get latest weight", slog.String("error", err.Error()))
		// Don't fail the request if we can't get weight
		weight = nil
	}

	// Default language if not set
	language := "en_US"
	if user.Language != nil {
		language = *user.Language
	}

	response := models.ProfileResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Age:       user.Age,
		Height:    user.Height,
		Weight:    weight,
		Language:  language,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateProfile updates the current user's profile
func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("Updating profile", slog.Int("user_id", userID))

	var req models.ProfileUpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		h.logger.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}
	defer tx.Rollback()

	// Update user profile
	if err := h.userRepo.UpdateProfile(userID, &req); err != nil {
		h.logger.Error("Failed to update profile", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}

	// If weight is provided, add it to weight history
	if req.Weight != nil {
		if err := h.userRepo.AddWeightEntry(userID, *req.Weight); err != nil {
			h.logger.Error("Failed to add weight entry", slog.String("error", err.Error()))
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update weight")
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		h.logger.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update profile")
	}

	// Return updated profile
	return h.GetProfile(c)
}

// RegisterRoutes registers the profile routes
func (h *ProfileHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/profile", h.GetProfile)
	g.PUT("/profile", h.UpdateProfile)
}