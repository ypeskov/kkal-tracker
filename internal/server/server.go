package server

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/middleware"
	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/routes/api"
	"ypeskov/kkal-tracker/internal/routes/web"
	"ypeskov/kkal-tracker/internal/services"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func New(cfg *config.Config, logger *slog.Logger, db *sql.DB, staticFiles embed.FS) *http.Server {
	e := echo.New()
	//e.HideBanner = true

	//e.Use(middleware.Logger(logger))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	jwtService := auth.NewJWTService(cfg.JWTSecret)
	authMiddleware := middleware.NewAuthMiddleware(jwtService, logger)

	userRepo := models.NewUserRepository(db)
	calorieRepo := models.NewCalorieEntryRepository(db, logger)
	ingredientRepo := models.NewIngredientRepository(db)

	authService := services.NewAuthService(userRepo, ingredientRepo, jwtService, logger)
	calorieService := services.NewCalorieService(calorieRepo, ingredientRepo, logger)

	authHandler := api.NewAuthHandler(authService, logger)
	calorieHandler := api.NewCalorieHandler(calorieService, logger)
	ingredientHandler := api.NewIngredientHandler(ingredientRepo, logger)

	apiGroup := e.Group("/api")

	authGroup := apiGroup.Group("/auth")
	authHandler.RegisterRoutes(authGroup, authMiddleware)

	caloriesGroup := apiGroup.Group("/calories", authMiddleware.RequireAuth)
	calorieHandler.RegisterRoutes(caloriesGroup)

	ingredientsGroup := apiGroup.Group("/ingredients", authMiddleware.RequireAuth)
	ingredientHandler.RegisterRoutes(ingredientsGroup)

	web.RegisterStaticRoutes(e, staticFiles)

	logger.Info("Server configured", "port", cfg.Port)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: e,
	}
}
