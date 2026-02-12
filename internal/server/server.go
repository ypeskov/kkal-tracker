package server

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/config"
	aihandler "ypeskov/kkal-tracker/internal/handlers/ai"
	apidatahandler "ypeskov/kkal-tracker/internal/handlers/apidata"
	apikeyhandler "ypeskov/kkal-tracker/internal/handlers/apikey"
	authhandler "ypeskov/kkal-tracker/internal/handlers/auth"
	"ypeskov/kkal-tracker/internal/handlers/calories"
	exporthandler "ypeskov/kkal-tracker/internal/handlers/export"
	"ypeskov/kkal-tracker/internal/handlers/ingredients"
	languageshandler "ypeskov/kkal-tracker/internal/handlers/languages"
	metricshandler "ypeskov/kkal-tracker/internal/handlers/metrics"
	"ypeskov/kkal-tracker/internal/handlers/profile"
	reportshandler "ypeskov/kkal-tracker/internal/handlers/reports"
	"ypeskov/kkal-tracker/internal/handlers/static"
	weighthandler "ypeskov/kkal-tracker/internal/handlers/weight"
	"ypeskov/kkal-tracker/internal/middleware"
	"ypeskov/kkal-tracker/internal/repositories"
	aiservice "ypeskov/kkal-tracker/internal/services/ai"
	apikeyservice "ypeskov/kkal-tracker/internal/services/apikey"
	authservice "ypeskov/kkal-tracker/internal/services/auth"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	emailservice "ypeskov/kkal-tracker/internal/services/email"
	exportservice "ypeskov/kkal-tracker/internal/services/export"
	ingredientservice "ypeskov/kkal-tracker/internal/services/ingredient"
	metricsservice "ypeskov/kkal-tracker/internal/services/metrics"
	profileservice "ypeskov/kkal-tracker/internal/services/profile"
	reportsservice "ypeskov/kkal-tracker/internal/services/reports"
	weightservice "ypeskov/kkal-tracker/internal/services/weight"

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
	tokenRepo      repositories.ActivationTokenRepository
	calorieRepo    repositories.CalorieEntryRepository
	ingredientRepo repositories.IngredientRepository
	weightRepo     repositories.WeightHistoryRepository
	apiKeyRepo     repositories.APIKeyRepository
}

// setupRepositories configures repositories based on the database type
func (s *Server) setupRepositories() error {
	switch s.config.DatabaseType {
	case "sqlite":
		s.userRepo = repositories.NewUserRepository(s.db, s.logger, repositories.DialectSQLite)
		s.tokenRepo = repositories.NewActivationTokenRepository(s.db, repositories.DialectSQLite, s.logger)
		s.calorieRepo = repositories.NewCalorieEntryRepository(s.db, s.logger, repositories.DialectSQLite)
		s.ingredientRepo = repositories.NewIngredientRepository(s.db, s.logger, repositories.DialectSQLite)
		s.weightRepo = repositories.NewWeightHistoryRepository(s.db, s.logger, repositories.DialectSQLite)
		s.apiKeyRepo = repositories.NewAPIKeyRepository(s.db, repositories.DialectSQLite, s.logger)
		s.logger.Debug("Configured SQLite repositories")
	case "postgres":
		s.userRepo = repositories.NewUserRepository(s.db, s.logger, repositories.DialectPostgres)
		s.tokenRepo = repositories.NewActivationTokenRepository(s.db, repositories.DialectPostgres, s.logger)
		s.calorieRepo = repositories.NewCalorieEntryRepository(s.db, s.logger, repositories.DialectPostgres)
		s.ingredientRepo = repositories.NewIngredientRepository(s.db, s.logger, repositories.DialectPostgres)
		s.weightRepo = repositories.NewWeightHistoryRepository(s.db, s.logger, repositories.DialectPostgres)
		s.apiKeyRepo = repositories.NewAPIKeyRepository(s.db, repositories.DialectPostgres, s.logger)
		s.logger.Debug("Configured PostgreSQL repositories")
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
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{s.config.AppURL},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization, "X-API-Key"},
	}))
	e.Use(echomiddleware.SecureWithConfig(echomiddleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		HSTSMaxAge:         31536000,
	}))

	jwtService := auth.NewJWTService(s.config.JWTSecret)
	authMiddleware := middleware.NewAuthMiddleware(jwtService, s.logger)

	// Initialize email service for activation emails
	emailService := emailservice.New(s.config, s.logger)

	// Initialize auth service with all dependencies
	authService := authservice.New(s.userRepo, s.tokenRepo, jwtService, emailService, s.logger)
	calorieService := calorieservice.New(s.calorieRepo, s.ingredientRepo, s.logger)
	ingredientService := ingredientservice.New(s.ingredientRepo, s.logger)
	profileService := profileservice.New(s.db, s.userRepo, s.weightRepo, s.logger)
	weightService := weightservice.New(s.weightRepo, s.logger)
	metricsService := metricsservice.New(s.userRepo, s.weightRepo, s.logger)
	reportsService := reportsservice.New(calorieService, weightService, s.logger)
	aiSvc := aiservice.New(s.config, s.logger)
	exportSvc := exportservice.New(calorieService, weightService, emailService, s.logger)
	apiKeySvc := apikeyservice.New(s.apiKeyRepo, s.logger)

	authHandler := authhandler.NewHandler(authService, s.logger)
	calorieHandler := calories.New(calorieService, s.logger)
	ingredientHandler := ingredients.NewHandler(ingredientService, s.logger)
	profileHandler := profile.NewProfileHandler(profileService, s.logger)
	weightHandler := weighthandler.NewHandler(weightService, s.logger)
	metricsHandler := metricshandler.NewMetricsHandler(metricsService, s.logger)
	reportsHandler := reportshandler.New(reportsService, s.logger)
	aiHandler := aihandler.New(aiSvc, calorieService, weightService, s.userRepo, s.logger)
	exportHandler := exporthandler.New(exportSvc, s.userRepo, s.logger)
	apiKeyHandler := apikeyhandler.New(apiKeySvc, s.logger)
	apiKeyMiddleware := middleware.NewAPIKeyMiddleware(apiKeySvc, s.logger)
	apiDataHandler := apidatahandler.New(calorieService, weightService, s.logger)

	apiGroup := e.Group("/api")

	// Public routes (no auth required)
	languagesHandler := languageshandler.NewHandler(s.logger)
	languagesGroup := apiGroup.Group("/languages")
	languagesHandler.RegisterRoutes(languagesGroup)

	// Auth routes with rate limiting (5 requests per second to prevent brute force)
	// NOTE: In-memory rate limiter is per-instance. If scaling to multiple replicas,
	// switch to a distributed store (e.g., Redis) for effective rate limiting.
	authRateLimiter := echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(5))
	authGroup := apiGroup.Group("/auth", authRateLimiter)
	authHandler.RegisterRoutes(authGroup, authMiddleware)

	caloriesGroup := apiGroup.Group("/calories", authMiddleware.RequireAuth)
	calorieHandler.RegisterRoutes(caloriesGroup)

	ingredientsGroup := apiGroup.Group("/ingredients", authMiddleware.RequireAuth)
	ingredientHandler.RegisterRoutes(ingredientsGroup)

	// Profile routes require authentication
	profileGroup := apiGroup.Group("", authMiddleware.RequireAuth)
	profileHandler.RegisterRoutes(profileGroup)

	// Weight history routes require authentication
	weightGroup := apiGroup.Group("/weight", authMiddleware.RequireAuth)
	weightHandler.RegisterRoutes(weightGroup)

	// Health metrics routes require authentication
	metricsGroup := apiGroup.Group("", authMiddleware.RequireAuth)
	metricsHandler.RegisterRoutes(metricsGroup)

	// Reports routes require authentication
	reportsGroup := apiGroup.Group("/reports", authMiddleware.RequireAuth)
	reportsHandler.RegisterRoutes(reportsGroup)

	// AI routes require authentication and strict rate limiting
	// Rate: 2 requests per minute (1 request every 30 seconds) to control AI costs
	aiRateLimiter := echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(2.0/60.0))
	aiGroup := apiGroup.Group("/ai", authMiddleware.RequireAuth, aiRateLimiter)
	aiHandler.RegisterRoutes(aiGroup)

	// Export routes require authentication
	exportGroup := apiGroup.Group("/export", authMiddleware.RequireAuth)
	exportHandler.RegisterRoutes(exportGroup)

	// API key management routes (JWT auth - user manages their keys)
	apiKeysGroup := apiGroup.Group("/api-keys", authMiddleware.RequireAuth)
	apiKeyHandler.RegisterRoutes(apiKeysGroup)

	// External data API (API key auth, rate limited: 60 req/min)
	v1RateLimiter := echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(1))
	v1Group := apiGroup.Group("/v1", apiKeyMiddleware.RequireAPIKey, v1RateLimiter)
	apiDataHandler.RegisterRoutes(v1Group)

	staticHandler := static.New(s.staticFiles, s.logger)
	staticHandler.RegisterRoutes(e)

	s.logger.Info("Server configured", "port", s.config.Port, "database_type", s.config.DatabaseType)

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.Port),
		Handler: e,
	}
}
