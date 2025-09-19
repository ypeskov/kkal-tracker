package ingredients

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	ingredientservice "ypeskov/kkal-tracker/internal/services/ingredient"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	ingredientService *ingredientservice.Service
	logger            *slog.Logger
}

func NewHandler(ingredientService *ingredientservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		ingredientService: ingredientService,
		logger:            logger,
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
	h.logger.Debug("GetAllIngredients called", "user_id", userID)

	ingredients, err := h.ingredientService.GetAllIngredients(userID)
	if err != nil {
		h.logger.Error("failed to get user ingredients", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("GetAllIngredients returning ingredients", "user_id", userID, "count", len(ingredients))
	return c.JSON(http.StatusOK, ingredients)
}

// GetIngredientByID Get a single user ingredient by ID
func (h *Handler) GetIngredientByID(c echo.Context) error {
	userID := c.Get("user_id").(int)
	ingredientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Debug("GetIngredientByID failed - invalid ID", "user_id", userID, "id_param", c.Param("id"), "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ingredient ID")
	}
	h.logger.Debug("GetIngredientByID called", "user_id", userID, "ingredient_id", ingredientID)

	ingredient, err := h.ingredientService.GetIngredientByID(userID, ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingredient not found")
		}
		h.logger.Error("Failed to get user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("GetIngredientByID returning ingredient", "user_id", userID, "ingredient_id", ingredientID, "name", ingredient.Name)
	return c.JSON(http.StatusOK, ingredient)
}

// CreateIngredient Create a new user ingredient
func (h *Handler) CreateIngredient(c echo.Context) error {
	userID := c.Get("user_id").(int)
	h.logger.Debug("CreateIngredient called", "user_id", userID)

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Debug("CreateIngredient failed - invalid request body", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.Debug("CreateIngredient failed - validation error", "user_id", userID, "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	serviceReq := &ingredientservice.CreateIngredientRequest{
		UserID:      userID,
		Name:        req.Name,
		KcalPer100g: req.KcalPer100g,
		Fats:        req.Fats,
		Carbs:       req.Carbs,
		Proteins:    req.Proteins,
	}

	ingredient, err := h.ingredientService.CreateIngredient(serviceReq)
	if err != nil {
		h.logger.Error("Failed to create user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("CreateIngredient successful", "user_id", userID, "ingredient_id", ingredient.ID, "name", ingredient.Name)
	return c.JSON(http.StatusCreated, ingredient)
}

// UpdateIngredient Update an existing user ingredient
func (h *Handler) UpdateIngredient(c echo.Context) error {
	userID := c.Get("user_id").(int)
	ingredientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Debug("UpdateIngredient failed - invalid ID", "user_id", userID, "id_param", c.Param("id"), "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ingredient ID")
	}
	h.logger.Debug("UpdateIngredient called", "user_id", userID, "ingredient_id", ingredientID)

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Debug("UpdateIngredient failed - invalid request body", "user_id", userID, "ingredient_id", ingredientID, "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.Debug("UpdateIngredient failed - validation error", "user_id", userID, "ingredient_id", ingredientID, "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	serviceReq := &ingredientservice.UpdateIngredientRequest{
		IngredientID: ingredientID,
		UserID:       userID,
		Name:         req.Name,
		KcalPer100g:  req.KcalPer100g,
		Fats:         req.Fats,
		Carbs:        req.Carbs,
		Proteins:     req.Proteins,
	}

	ingredient, err := h.ingredientService.UpdateIngredient(serviceReq)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingredient not found")
		}
		h.logger.Error("Failed to update user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("UpdateIngredient successful", "user_id", userID, "ingredient_id", ingredientID, "name", ingredient.Name)
	return c.JSON(http.StatusOK, ingredient)
}

// DeleteIngredient Delete a user ingredient
func (h *Handler) DeleteIngredient(c echo.Context) error {
	userID := c.Get("user_id").(int)
	ingredientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Debug("DeleteIngredient failed - invalid ID", "user_id", userID, "id_param", c.Param("id"), "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ingredient ID")
	}
	h.logger.Debug("DeleteIngredient called", "user_id", userID, "ingredient_id", ingredientID)

	err = h.ingredientService.DeleteIngredient(userID, ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Ingredient not found")
		}
		h.logger.Error("Failed to delete user ingredient", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
	}

	h.logger.Debug("DeleteIngredient successful", "user_id", userID, "ingredient_id", ingredientID)
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("", h.GetAllIngredients)
	g.GET("/:id", h.GetIngredientByID)
	g.POST("", h.CreateIngredient)
	g.PUT("/:id", h.UpdateIngredient)
	g.DELETE("/:id", h.DeleteIngredient)
}
