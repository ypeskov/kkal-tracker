package preferences

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	preferencesservice "ypeskov/kkal-tracker/internal/services/preferences"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	preferencesService preferencesservice.Servicer
	logger             *slog.Logger
}

func New(preferencesService preferencesservice.Servicer, logger *slog.Logger) *Handler {
	return &Handler{
		preferencesService: preferencesService,
		logger:             logger.With("handler", "preferences"),
	}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/:key", h.GetPreference)
	g.PUT("/:key", h.SavePreference)
}

// GetPreference returns the user's value of the key, with defaults for anything not stored
func (h *Handler) GetPreference(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	key := c.Param("key")

	value, err := h.preferencesService.Get(userID, key)
	if err != nil {
		return h.errorResponse(err, userID, key, "Failed to load preferences")
	}
	return c.JSON(http.StatusOK, value)
}

// SavePreference replaces the user's value of the key with the request body and returns the stored value
func (h *Handler) SavePreference(c *echo.Context) error {
	userID := c.Get("user_id").(int)
	key := c.Param("key")

	// The size is limited by the route's BodyLimit; a body without Content-Length fails here when it is too large
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		if errors.Is(err, echo.ErrStatusRequestEntityTooLarge) {
			return err
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	value, err := h.preferencesService.Save(userID, key, body)
	if err != nil {
		return h.errorResponse(err, userID, key, "Failed to save preferences")
	}
	return c.JSON(http.StatusOK, value)
}

// errorResponse maps a service error to a response; internal errors are logged and answered with failureMessage
func (h *Handler) errorResponse(err error, userID int, key, failureMessage string) error {
	var validationErr *preferencesservice.ValidationError
	switch {
	case errors.Is(err, preferencesservice.ErrUnknownKey):
		return echo.NewHTTPError(http.StatusNotFound, "Unknown preference")
	case errors.Is(err, preferencesservice.ErrInvalidBody):
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	case errors.As(err, &validationErr):
		return echo.NewHTTPError(http.StatusBadRequest, validationErr.Message)
	default:
		h.logger.Error(failureMessage, "error", err, "user_id", userID, "key", key)
		return echo.NewHTTPError(http.StatusInternalServerError, failureMessage)
	}
}
