package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"ypeskov/kkal-tracker/internal/auth"

	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
	logger     *slog.Logger
}

func NewAuthMiddleware(jwtService *auth.JWTService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		logger:     logger,
	}
}

func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header required")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return echo.NewHTTPError(http.StatusUnauthorized, "Bearer token required")
		}

		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			m.logger.Debug("Invalid token", "error", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		return next(c)
	}
}