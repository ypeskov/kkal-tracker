package reports

import (
	"log/slog"
	"net/http"

	reportsservice "ypeskov/kkal-tracker/internal/services/reports"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	reportsService *reportsservice.Service
	logger         *slog.Logger
}

func New(reportsService *reportsservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		reportsService: reportsService,
		logger:         logger.With("handler", "reports"),
	}
}

// GetReportData returns combined weight and calorie data for the specified date range
func (h *Handler) GetReportData(c echo.Context) error {
	userID := c.Get("user_id").(int)
	dateFrom := c.QueryParam("from")
	dateTo := c.QueryParam("to")

	h.logger.Debug("GetReportData called", "user_id", userID, "from", dateFrom, "to", dateTo)

	// Delegate to service for business logic
	reportData, err := h.reportsService.GetAggregatedMetrics(userID, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("failed to get aggregated metrics", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, userFacingErrorMessage)
	}

	h.logger.Debug("GetReportData end")
	return c.JSON(http.StatusOK, reportData)
}

// RegisterRoutes registers all report-related routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/data", h.GetReportData)
}
