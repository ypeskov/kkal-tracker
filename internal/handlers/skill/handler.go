package skill

import (
	"fmt"
	"log/slog"
	"net/http"

	skillservice "ypeskov/kkal-tracker/internal/services/skill"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	skillService skillservice.Servicer
	logger       *slog.Logger
}

func New(skillService skillservice.Servicer, logger *slog.Logger) *Handler {
	return &Handler{
		skillService: skillService,
		logger:       logger.With("handler", "skill"),
	}
}

// RegisterRoutes registers the skill routes. POST: every download creates an API key.
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/food-log/download", h.DownloadFoodLog)
}

// DownloadFoodLog returns the food logging skill as a zip with a new API key of the user inside
func (h *Handler) DownloadFoodLog(c *echo.Context) error {
	userID := c.Get("user_id").(int)

	download, err := h.skillService.DownloadFoodLog(userID)
	if err != nil {
		h.logger.Error("Failed to prepare the food logging skill", "error", err, "user_id", userID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to prepare the skill")
	}

	header := c.Response().Header()
	header.Set(echo.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, download.FileName))
	// The body holds a secret: no browser or proxy cache may keep it
	header.Set(echo.HeaderCacheControl, "no-store")
	return c.Blob(http.StatusOK, "application/zip", download.Content)
}
