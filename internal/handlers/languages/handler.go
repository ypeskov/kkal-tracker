package languages

import (
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/config"

	"github.com/labstack/echo/v4"
)

// LanguageResponse represents a language in the API response
type LanguageResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Handler handles language-related requests
type Handler struct {
	logger *slog.Logger
}

// NewHandler creates a new languages handler
func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger.With("handler", "languages"),
	}
}

// GetLanguages returns the list of supported languages
func (h *Handler) GetLanguages(c echo.Context) error {
	h.logger.Debug("GetLanguages called")
	languages := make([]LanguageResponse, 0, len(config.SupportedLanguages))
	for _, lang := range config.SupportedLanguages {
		languages = append(languages, LanguageResponse{
			Code: lang.Code,
			Name: lang.Name,
		})
	}

	h.logger.Debug("GetLanguages returning languages", "languages", languages)
	return c.JSON(http.StatusOK, languages)
}

// RegisterRoutes registers the language routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetLanguages)
}
