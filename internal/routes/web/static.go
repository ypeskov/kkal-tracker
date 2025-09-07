package web

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RegisterStaticRoutes(e *echo.Echo) {
	// Serve built frontend files from web/dist
	e.Static("/", "web/dist")
	
	// SPA fallback - serve index.html for any unmatched routes
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "web/dist",
		HTML5:  true,
		Browse: false,
	}))
}