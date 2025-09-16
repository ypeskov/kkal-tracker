package server

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/config"
	authhandler "ypeskov/kkal-tracker/internal/handlers/auth"
	"ypeskov/kkal-tracker/internal/handlers/calories"
	"ypeskov/kkal-tracker/internal/handlers/ingredients"
	"ypeskov/kkal-tracker/internal/handlers/profile"
	"ypeskov/kkal-tracker/internal/handlers/static"
	"ypeskov/kkal-tracker/internal/middleware"
	"ypeskov/kkal-tracker/internal/repositories"
	authservice "ypeskov/kkal-tracker/internal/services/auth"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	ingredientservice "ypeskov/kkal-tracker/internal/services/ingredient"
	profileservice "ypeskov/kkal-tracker/internal/services/profile"

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
		s.userRepo = repositories.NewUserRepository(s.db, s.logger, repositories.DialectSQLite)
		s.calorieRepo = repositories.NewCalorieEntryRepository(s.db, s.logger, repositories.DialectSQLite)
		s.ingredientRepo = repositories.NewIngredientRepository(s.db, s.logger, repositories.DialectSQLite)
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

	authService := authservice.NewService(s.userRepo, jwtService, s.logger)
	calorieService := calorieservice.New(s.calorieRepo, s.ingredientRepo, s.logger)
	ingredientService := ingredientservice.NewService(s.ingredientRepo, s.logger)
	profileService := profileservice.NewService(s.db, s.userRepo, s.logger)

	authHandler := authhandler.NewHandler(authService, s.logger)
	calorieHandler := calories.NewHandler(calorieService, s.logger)
	ingredientHandler := ingredients.NewHandler(ingredientService, s.logger)
	profileHandler := profile.NewProfileHandler(profileService, s.logger)

	apiGroup := e.Group("/api")

	authGroup := apiGroup.Group("/auth")
	authHandler.RegisterRoutes(authGroup, authMiddleware)

	caloriesGroup := apiGroup.Group("/calories", authMiddleware.RequireAuth)
	calorieHandler.RegisterRoutes(caloriesGroup)

	ingredientsGroup := apiGroup.Group("/ingredients", authMiddleware.RequireAuth)
	ingredientHandler.RegisterRoutes(ingredientsGroup)

	// Profile routes require authentication
	profileGroup := apiGroup.Group("", authMiddleware.RequireAuth)
	profileHandler.RegisterRoutes(profileGroup)

	staticHandler := static.NewHandler(s.staticFiles)
	staticHandler.RegisterRoutes(e)

	s.logger.Info("Server configured", "port", s.config.Port, "database_type", s.config.DatabaseType)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.Port),
		Handler: e,
	}
}
