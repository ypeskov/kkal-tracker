package static

import (
	"embed"
	"io/fs"
	"log/slog"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
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

	// Serve built frontend files from the embedded filesystem, with SPA fallback:
	// index.html is served for any path that matches neither a file nor a route.
	// There must be no catch-all static route next to this middleware: the fallback
	// is applied only to router-level 404s, so that API 404 responses stay untouched.
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: distFS,
		HTML5:      true,
		Browse:     false,
	}))

	h.logger.Debug("Static file serving configured successfully")
}
