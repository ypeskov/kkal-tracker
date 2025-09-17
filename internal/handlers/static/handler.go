package static

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handler struct {
	staticFiles embed.FS
	logger      *slog.Logger
}

func New(staticFiles embed.FS, logger *slog.Logger) *Handler {
	return &Handler{
		staticFiles: staticFiles,
		logger:      logger.With("handler", "static"),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	h.logger.Debug("RegisterRoutes called - setting up static file serving")

	// Get the embedded filesystem for dist
	distFS, err := fs.Sub(h.staticFiles, "dist")
	if err != nil {
		h.logger.Error("Failed to create dist filesystem", "error", err)
		panic(err)
	}

	// Serve built frontend files from embedded filesystem
	e.StaticFS("/", distFS)

	// SPA fallback - serve index.html for any unmatched routes
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: http.FS(distFS),
		HTML5:      true,
		Browse:     false,
	}))

	h.logger.Debug("Static file serving configured successfully")
}
