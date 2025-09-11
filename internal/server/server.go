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
	"ypeskov/kkal-tracker/internal/repositories"
	"ypeskov/kkal-tracker/internal/repositories/sqlite"
	authhandler "ypeskov/kkal-tracker/internal/handlers/auth"
	"ypeskov/kkal-tracker/internal/handlers/calories"
	"ypeskov/kkal-tracker/internal/handlers/ingredients"
	"ypeskov/kkal-tracker/internal/routes/web"
	"ypeskov/kkal-tracker/internal/services"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// Server holds all the dependencies and repositories
type Server struct {
	config         *config.Config
	logger         *slog.Logger
	db             *sql.DB
	staticFiles    embed.FS
	userRepo       repositories.UserRepository
	calorieRepo    repositories.CalorieEntryRepository
	ingredientRepo repositories.IngredientRepository
}

// setupRepositories configures repositories based on the database type
func (s *Server) setupRepositories() error {
	switch s.config.DatabaseType {
	case "sqlite":
		s.userRepo = sqlite.NewUserRepository(s.db, s.logger)
		s.calorieRepo = sqlite.NewCalorieEntryRepository(s.db, s.logger)
		s.ingredientRepo = sqlite.NewIngredientRepository(s.db, s.logger)
		s.logger.Debug("Configured SQLite repositories")
	case "postgres":
		// TODO: Implement PostgreSQL repositories when needed
		// s.userRepo = postgres.NewUserRepository(s.db, s.logger)
		// s.calorieRepo = postgres.NewCalorieEntryRepository(s.db, s.logger)
		// s.ingredientRepo = postgres.NewIngredientRepository(s.db, s.logger)
		return fmt.Errorf("postgres repositories not yet implemented")
	default:
		return fmt.Errorf("unsupported database type: %s", s.config.DatabaseType)
	}

	return nil
}

// New creates and configures a new server instance
func New(cfg *config.Config, logger *slog.Logger, db *sql.DB, staticFiles embed.FS) (*Server, error) {
	server := &Server{
		config:      cfg,
		logger:      logger,
		db:          db,
		staticFiles: staticFiles,
	}

	// Setup repositories based on config
	if err := server.setupRepositories(); err != nil {
		return nil, fmt.Errorf("failed to setup repositories: %w", err)
	}

	return server, nil
}

// Start configures and starts the HTTP server
func (s *Server) Start() *http.Server {
	e := echo.New()
	//e.HideBanner = true

	// Register custom validator
	e.Validator = middleware.NewValidator()

	//e.Use(middleware.Logger(s.logger))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	jwtService := auth.NewJWTService(s.config.JWTSecret)
	authMiddleware := middleware.NewAuthMiddleware(jwtService, s.logger)

	authService := services.NewAuthService(s.userRepo, s.ingredientRepo, jwtService, s.logger)
	calorieService := services.NewCalorieService(s.calorieRepo, s.ingredientRepo, s.logger)

	authHandler := authhandler.NewHandler(authService, s.logger)
	calorieHandler := calories.NewHandler(calorieService, s.logger)
	ingredientHandler := ingredients.NewHandler(s.ingredientRepo, s.logger)

	apiGroup := e.Group("/api")

	authGroup := apiGroup.Group("/auth")
	authHandler.RegisterRoutes(authGroup, authMiddleware)

	caloriesGroup := apiGroup.Group("/calories", authMiddleware.RequireAuth)
	calorieHandler.RegisterRoutes(caloriesGroup)

	ingredientsGroup := apiGroup.Group("/ingredients", authMiddleware.RequireAuth)
	ingredientHandler.RegisterRoutes(ingredientsGroup)

	web.RegisterStaticRoutes(e, s.staticFiles)

	s.logger.Info("Server configured", "port", s.config.Port, "database_type", s.config.DatabaseType)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.Port),
		Handler: e,
	}
}

