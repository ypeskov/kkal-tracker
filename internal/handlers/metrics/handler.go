package metrics

import (
	"log/slog"
	"net/http"

	metricsservice "ypeskov/kkal-tracker/internal/services/metrics"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	metricsService *metricsservice.Service
	logger         *slog.Logger
}

func NewMetricsHandler(metricsService *metricsservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		metricsService: metricsService,
		logger:         logger,
	}
}

// GetHealthMetrics returns calculated health metrics for the current user
func (h *Handler) GetHealthMetrics(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("GetHealthMetrics called", "user_id", userID)

	metrics, err := h.metricsService.GetHealthMetrics(userID)
	if err != nil {
		h.logger.Error("Failed to get health metrics", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get health metrics")
	}

	h.logger.Debug("GetHealthMetrics returning metrics", "user_id", userID)
	return c.JSON(http.StatusOK, metrics)
}

// RegisterRoutes registers the metrics routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/metrics", h.GetHealthMetrics)
}
