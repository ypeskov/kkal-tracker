package api

import (
	"errors"
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/middleware"
	"ypeskov/kkal-tracker/internal/services"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *services.AuthService
	logger      *slog.Logger
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6"`
	LanguageCode string `json:"language_code" validate:"required"`
}

type ResponseUser struct {
	Email string `json:"email"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  ResponseUser `json:"user"`
}

func NewAuthHandler(authService *services.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	user, token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
		}
		h.logger.Error("Login failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	responseUser := ResponseUser{
		Email: user.Email,
	}
	h.logger.Debug("%+v", "responseUser", responseUser)

	return c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  responseUser,
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	user, token, err := h.authService.Register(req.Email, req.Password, req.LanguageCode)
	if err != nil {
		if err.Error() == "user already exists" {
			return echo.NewHTTPError(http.StatusConflict, "User already exists")
		}
		h.logger.Error("Registration failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	responseUser := ResponseUser{
		Email: user.Email,
	}

	return c.JSON(http.StatusCreated, &LoginResponse{
		Token: token,
		User:  responseUser,
	})
}

func (h *AuthHandler) GetCurrentUser(c echo.Context) error {
	userID := c.Get("user_id").(int)

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		h.logger.Error("Failed to get current user", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) RegisterRoutes(g *echo.Group, authMiddleware *middleware.AuthMiddleware) {
	g.POST("/login", h.Login)
	g.POST("/register", h.Register)
	g.GET("/me", h.GetCurrentUser, authMiddleware.RequireAuth)
}
