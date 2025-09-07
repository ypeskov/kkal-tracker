package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RegisterStaticRoutes(e *echo.Echo, staticFiles embed.FS) {
	// Get the embedded filesystem for dist
	distFS, err := fs.Sub(staticFiles, "dist")
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