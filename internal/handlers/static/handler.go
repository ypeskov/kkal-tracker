package static

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handler struct {
	staticFiles embed.FS
}

func NewHandler(staticFiles embed.FS) *Handler {
	return &Handler{
		staticFiles: staticFiles,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// Get the embedded filesystem for dist
	distFS, err := fs.Sub(h.staticFiles, "dist")
	if err != nil {
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
}