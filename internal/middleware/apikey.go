package middleware

import (
	"log/slog"
	"net/http"

	apikeyservice "ypeskov/kkal-tracker/internal/services/apikey"

	"github.com/labstack/echo/v4"
)

type APIKeyMiddleware struct {
	apiKeyService *apikeyservice.Service
	logger        *slog.Logger
}

func NewAPIKeyMiddleware(apiKeyService *apikeyservice.Service, logger *slog.Logger) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKeyService: apiKeyService,
		logger:        logger,
	}
}

func (m *APIKeyMiddleware) RequireAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		rawKey := c.Request().Header.Get("X-API-Key")
		if rawKey == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "API key required")
		}

		apiKey, err := m.apiKeyService.ValidateKey(rawKey)
		if err != nil {
			m.logger.Debug("API key validation failed", "error", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired API key")
		}

		c.Set("user_id", apiKey.UserID)

		return next(c)
	}
}
