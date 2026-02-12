package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"ypeskov/kkal-tracker/internal/middleware"
	authservice "ypeskov/kkal-tracker/internal/services/auth"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	authService *authservice.Service
	logger      *slog.Logger
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=6,max=72"`
	LanguageCode string `json:"language_code" validate:"required"`
}

type ResponseUser struct {
	Email string `json:"email"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  ResponseUser `json:"user"`
}

func NewHandler(authService *authservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

func (h *Handler) Login(c echo.Context) error {
	h.logger.Debug("Login called")

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Debug("Login failed - invalid request body", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	h.logger.Debug("Login attempt", "email", req.Email)

	user, token, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
		}
		if errors.Is(err, authservice.ErrUserNotActivated) {
			return echo.NewHTTPError(http.StatusForbidden, "Account not activated. Please check your email.")
		}
		h.logger.Error("Login failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	responseUser := ResponseUser{
		Email: user.Email,
	}

	h.logger.Debug("Login successful", "email", user.Email)
	return c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  responseUser,
	})
}

func (h *Handler) Register(c echo.Context) error {
	h.logger.Debug("Register called")

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Debug("Register failed - invalid request body", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	h.logger.Debug("Registration attempt", "email", req.Email, "language", req.LanguageCode)

	// Web registration - skipActivation=false (requires email activation)
	user, _, err := h.authService.Register(req.Email, req.Password, req.LanguageCode, false)
	if err != nil {
		if errors.Is(err, authservice.ErrUserAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "User already exists")
		}
		if errors.Is(err, authservice.ErrEmailSendFailed) {
			h.logger.Error("Registration failed - email sending failed", "error", err)
			return echo.NewHTTPError(http.StatusServiceUnavailable, "Failed to send activation email. Please try again later or contact support.")
		}
		h.logger.Error("Registration failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("Registration successful, activation email sent", "email", user.Email, "language", req.LanguageCode)

	// Return success message without token (user must activate first)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Registration successful. Please check your email to activate your account.",
		"email":   user.Email,
	})
}

func (h *Handler) GetCurrentUser(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("GetCurrentUser called", "user_id", userID)

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		if errors.Is(err, authservice.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		h.logger.Error("Failed to get current user", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("GetCurrentUser successful", "user_id", userID, "email", user.Email)
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) Activate(c echo.Context) error {
	token := c.Param("token")

	if len(token) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid activation token")
	}

	h.logger.Debug("Activate called", "token", token[:8]+"...")

	err := h.authService.ActivateUser(token)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidToken) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired activation token")
		}
		h.logger.Error("Activation failed", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("Activation successful")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account activated successfully. You can now log in.",
	})
}

func (h *Handler) RegisterRoutes(g *echo.Group, authMiddleware *middleware.AuthMiddleware) {
	g.POST("/login", h.Login)
	g.POST("/register", h.Register)
	g.GET("/activate/:token", h.Activate)
	g.GET("/me", h.GetCurrentUser, authMiddleware.RequireAuth)
}
