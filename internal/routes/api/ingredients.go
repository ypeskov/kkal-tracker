package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"ypeskov/kkal-tracker/internal/models"

	"github.com/labstack/echo/v4"
)

type IngredientHandler struct {
	ingredientRepo *models.IngredientRepository
	logger         *slog.Logger
}

func NewIngredientHandler(ingredientRepo *models.IngredientRepository, logger *slog.Logger) *IngredientHandler {
	return &IngredientHandler{
		ingredientRepo: ingredientRepo,
		logger:         logger,
	}
}

// GetAllIngredients Get all user ingredients for session storage caching
func (h *IngredientHandler) GetAllIngredients(c echo.Context) error {
	userID := c.Get("user_id").(int)

	ingredients, err := h.ingredientRepo.GetAllUserIngredients(userID)
	if err != nil {
		h.logger.Error("Failed to get user ingredients", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	return c.JSON(http.StatusOK, ingredients)
}

// SearchIngredients Search user ingredients for autocomplete (fallback if session storage fails)
func (h *IngredientHandler) SearchIngredients(c echo.Context) error {
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

func (h *IngredientHandler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetAllIngredients)
	g.GET("/search", h.SearchIngredients)
}
