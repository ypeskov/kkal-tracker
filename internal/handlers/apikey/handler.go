package apikey

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	apikeyservice "ypeskov/kkal-tracker/internal/services/apikey"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	apiKeyService *apikeyservice.Service
	logger        *slog.Logger
}

type CreateRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=100"`
	ExpiryDays *int   `json:"expiry_days,omitempty" validate:"omitempty,min=1,max=3650"`
}

type CreateResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	KeyPrefix string     `json:"key_prefix"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type APIKeyResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"key_prefix"`
	ExpiresAt *time.Time `json:"expires_at"`
	IsRevoked bool       `json:"is_revoked"`
	CreatedAt time.Time  `json:"created_at"`
}

func New(apiKeyService *apikeyservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		apiKeyService: apiKeyService,
		logger:        logger.With("handler", "apikey"),
	}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("", h.CreateAPIKey)
	g.GET("", h.ListAPIKeys)
	g.POST("/:id/revoke", h.RevokeAPIKey)
	g.DELETE("/:id", h.DeleteAPIKey)
}

func (h *Handler) CreateAPIKey(c echo.Context) error {
	userID := c.Get("user_id").(int)

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	apiKey, rawKey, err := h.apiKeyService.CreateKey(userID, req.Name, req.ExpiryDays)
	if err != nil {
		h.logger.Error("Failed to create API key", "error", err, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create API key")
	}

	return c.JSON(http.StatusCreated, CreateResponse{
		ID:        apiKey.ID,
		Name:      apiKey.Name,
		Key:       rawKey,
		KeyPrefix: apiKey.KeyPrefix,
		ExpiresAt: apiKey.ExpiresAt,
		CreatedAt: apiKey.CreatedAt,
	})
}

func (h *Handler) ListAPIKeys(c echo.Context) error {
	userID := c.Get("user_id").(int)

	keys, err := h.apiKeyService.GetUserKeys(userID)
	if err != nil {
		h.logger.Error("Failed to list API keys", "error", err, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list API keys")
	}

	response := make([]APIKeyResponse, len(keys))
	for i, key := range keys {
		response[i] = APIKeyResponse{
			ID:        key.ID,
			Name:      key.Name,
			KeyPrefix: key.KeyPrefix,
			ExpiresAt: key.ExpiresAt,
			IsRevoked: key.IsRevoked,
			CreatedAt: key.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) RevokeAPIKey(c echo.Context) error {
	userID := c.Get("user_id").(int)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid key ID")
	}

	if err := h.apiKeyService.RevokeKey(id, userID); err != nil {
		h.logger.Error("Failed to revoke API key", "error", err, "id", id, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to revoke API key")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "API key revoked"})
}

func (h *Handler) DeleteAPIKey(c echo.Context) error {
	userID := c.Get("user_id").(int)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid key ID")
	}

	if err := h.apiKeyService.DeleteKey(id, userID); err != nil {
		h.logger.Error("Failed to delete API key", "error", err, "id", id, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete API key")
	}

	return c.NoContent(http.StatusNoContent)
}
