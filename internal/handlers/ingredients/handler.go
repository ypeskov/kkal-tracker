package ingredients

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"ypeskov/kkal-tracker/internal/repositories"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	ingredientRepo repositories.IngredientRepository
	logger         *slog.Logger
}

func NewHandler(ingredientRepo repositories.IngredientRepository, logger *slog.Logger) *Handler {
	return &Handler{
		ingredientRepo: ingredientRepo,
		logger:         logger,
	}
}

type CreateRequest struct {
	Name        string   `json:"name" validate:"required"`
	KcalPer100g float64  `json:"kcalPer100g" validate:"required,min=0"`
	Fats        *float64 `json:"fats,omitempty" validate:"omitempty,min=0"`
	Carbs       *float64 `json:"carbs,omitempty" validate:"omitempty,min=0"`
	Proteins    *float64 `json:"proteins,omitempty" validate:"omitempty,min=0"`
}

type UpdateRequest struct {
	Name        string   `json:"name" validate:"required"`
	KcalPer100g float64  `json:"kcalPer100g" validate:"required,min=0"`
	Fats        *float64 `json:"fats,omitempty" validate:"omitempty,min=0"`
	Carbs       *float64 `json:"carbs,omitempty" validate:"omitempty,min=0"`
	Proteins    *float64 `json:"proteins,omitempty" validate:"omitempty,min=0"`
}

// GetAllIngredients Get all user ingredients for session storage caching
func (h *Handler) GetAllIngredients(c echo.Context) error {
	userID := c.Get("user_id").(int)

	ingredients, err := h.ingredientRepo.GetAllUserIngredients(userID)
	if err != nil {
		h.logger.Error("Failed to get user ingredients", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, ingredients)
}

// SearchIngredients Search user ingredients for autocomplete (fallback if session storage fails)
func (h *Handler) SearchIngredients(c echo.Context) error {
	userID := c.Get("user_id").(int)
	query := c.QueryParam("q")
	limitParam := c.QueryParam("limit")

	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Query parameter 'q' is required")
	}

	// Default limit to 10
	limit := 10
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	ingredients, err := h.ingredientRepo.SearchUserIngredients(userID, query, limit)
	if err != nil {
		h.logger.Error("Failed to search user ingredients", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, ingredients)
}

// GetIngredientByID Get a single user ingredient by ID
func (h *Handler) GetIngredientByID(c echo.Context) error {
	userID := c.Get("user_id").(int)
	ingredientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ingredient ID")
	}

	ingredient, err := h.ingredientRepo.GetUserIngredientByID(userID, ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingredient not found")
		}
		h.logger.Error("Failed to get user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, ingredient)
}

// CreateIngredient Create a new user ingredient
func (h *Handler) CreateIngredient(c echo.Context) error {
	userID := c.Get("user_id").(int)
	
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ingredient, err := h.ingredientRepo.CreateUserIngredient(
		userID, req.Name, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins,
	)
	if err != nil {
		h.logger.Error("Failed to create user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusCreated, ingredient)
}

// UpdateIngredient Update an existing user ingredient
func (h *Handler) UpdateIngredient(c echo.Context) error {
	userID := c.Get("user_id").(int)
	ingredientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ingredient ID")
	}

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ingredient, err := h.ingredientRepo.UpdateUserIngredient(
		userID, ingredientID, req.Name, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingredient not found")
		}
		h.logger.Error("Failed to update user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, ingredient)
}

// DeleteIngredient Delete a user ingredient
func (h *Handler) DeleteIngredient(c echo.Context) error {
	userID := c.Get("user_id").(int)
	ingredientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ingredient ID")
	}

	err = h.ingredientRepo.DeleteUserIngredient(userID, ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingredient not found")
		}
		h.logger.Error("Failed to delete user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetAllIngredients)
	g.GET("/search", h.SearchIngredients)
	g.GET("/:id", h.GetIngredientByID)
	g.POST("", h.CreateIngredient)
	g.PUT("/:id", h.UpdateIngredient)
	g.DELETE("/:id", h.DeleteIngredient)
}