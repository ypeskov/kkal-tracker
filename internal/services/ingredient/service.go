package ingredient

import (
	"log/slog"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

type Service struct {
	ingredientRepo repositories.IngredientRepository
	logger         *slog.Logger
}

func New(ingredientRepo repositories.IngredientRepository, logger *slog.Logger) *Service {
	return &Service{
		ingredientRepo: ingredientRepo,
		logger:         logger.With("service", "ingredient"),
	}
}

func (s *Service) GetAllIngredients(userID int) ([]*models.UserIngredient, error) {
	s.logger.Debug("GetAllIngredients called", "user_id", userID)

	ingredients, err := s.ingredientRepo.GetAllUserIngredients(userID)
	if err != nil {
		s.logger.Error("Failed to get user ingredients", "error", err, "user_id", userID)
		return nil, err
	}

	s.logger.Debug("GetAllIngredients completed successfully", "user_id", userID, "count", len(ingredients))
	return ingredients, nil
}

func (s *Service) SearchIngredients(req *SearchIngredientsRequest) ([]*models.UserIngredient, error) {
	s.logger.Debug("SearchIngredients called", "user_id", req.UserID, "query", req.Query, "limit", req.Limit)

	if req.Query == "" {
		s.logger.Debug("SearchIngredients failed - empty query", "user_id", req.UserID)
		return nil, ErrEmptyQuery
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	ingredients, err := s.ingredientRepo.SearchUserIngredients(req.UserID, req.Query, req.Limit)
	if err != nil {
		s.logger.Error("Failed to search user ingredients", "error", err, "user_id", req.UserID, "query", req.Query)
		return nil, err
	}

	s.logger.Debug("SearchIngredients completed successfully", "user_id", req.UserID, "query", req.Query, "limit", req.Limit, "results", len(ingredients))
	return ingredients, nil
}

func (s *Service) GetIngredientByID(userID, ingredientID int) (*models.UserIngredient, error) {
	s.logger.Debug("GetIngredientByID called", "user_id", userID, "ingredient_id", ingredientID)

	ingredient, err := s.ingredientRepo.GetUserIngredientByID(userID, ingredientID)
	if err != nil {
		s.logger.Error("Failed to get user ingredient by ID", "error", err, "user_id", userID, "ingredient_id", ingredientID)
		return nil, err
	}

	s.logger.Debug("GetIngredientByID completed successfully", "user_id", userID, "ingredient_id", ingredientID, "name", ingredient.Name)
	return ingredient, nil
}

func (s *Service) CreateIngredient(req *CreateIngredientRequest) (*models.UserIngredient, error) {
	s.logger.Debug("CreateIngredient called", "user_id", req.UserID, "name", req.Name, "kcal_per_100g", req.KcalPer100g)

	if err := s.validateCreateRequest(req); err != nil {
		s.logger.Debug("CreateIngredient failed - validation error", "user_id", req.UserID, "error", err)
		return nil, err
	}

	ingredient, err := s.ingredientRepo.CreateUserIngredient(
		req.UserID, req.Name, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins,
	)
	if err != nil {
		s.logger.Error("Failed to create user ingredient", "error", err, "user_id", req.UserID, "name", req.Name)
		return nil, err
	}

	s.logger.Debug("CreateIngredient completed successfully", "user_id", req.UserID, "ingredient_id", ingredient.ID, "name", req.Name)
	return ingredient, nil
}

func (s *Service) UpdateIngredient(req *UpdateIngredientRequest) (*models.UserIngredient, error) {
	s.logger.Debug("UpdateIngredient called", "user_id", req.UserID, "ingredient_id", req.IngredientID, "name", req.Name, "kcal_per_100g", req.KcalPer100g)

	if err := s.validateUpdateRequest(req); err != nil {
		s.logger.Debug("UpdateIngredient failed - validation error", "user_id", req.UserID, "ingredient_id", req.IngredientID, "error", err)
		return nil, err
	}

	ingredient, err := s.ingredientRepo.UpdateUserIngredient(
		req.UserID, req.IngredientID, req.Name, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins,
	)
	if err != nil {
		s.logger.Error("Failed to update user ingredient", "error", err, "user_id", req.UserID, "ingredient_id", req.IngredientID)
		return nil, err
	}

	s.logger.Debug("UpdateIngredient completed successfully", "user_id", req.UserID, "ingredient_id", req.IngredientID, "name", req.Name)
	return ingredient, nil
}

func (s *Service) DeleteIngredient(userID, ingredientID int) error {
	s.logger.Debug("DeleteIngredient called", "user_id", userID, "ingredient_id", ingredientID)

	err := s.ingredientRepo.DeleteUserIngredient(userID, ingredientID)
	if err != nil {
		s.logger.Error("Failed to delete user ingredient", "error", err, "user_id", userID, "ingredient_id", ingredientID)
		return err
	}

	s.logger.Debug("DeleteIngredient completed successfully", "user_id", userID, "ingredient_id", ingredientID)
	return nil
}

func (s *Service) validateCreateRequest(req *CreateIngredientRequest) error {
	if req.Name == "" {
		return ErrInvalidName
	}
	if req.KcalPer100g < 0 {
		return ErrInvalidKcalPer100g
	}
	if req.Fats != nil && *req.Fats < 0 {
		return ErrInvalidNutritionValue
	}
	if req.Carbs != nil && *req.Carbs < 0 {
		return ErrInvalidNutritionValue
	}
	if req.Proteins != nil && *req.Proteins < 0 {
		return ErrInvalidNutritionValue
	}
	return nil
}

func (s *Service) validateUpdateRequest(req *UpdateIngredientRequest) error {
	if req.IngredientID <= 0 {
		return ErrInvalidIngredientID
	}
	if req.Name == "" {
		return ErrInvalidName
	}
	if req.KcalPer100g < 0 {
		return ErrInvalidKcalPer100g
	}
	if req.Fats != nil && *req.Fats < 0 {
		return ErrInvalidNutritionValue
	}
	if req.Carbs != nil && *req.Carbs < 0 {
		return ErrInvalidNutritionValue
	}
	if req.Proteins != nil && *req.Proteins < 0 {
		return ErrInvalidNutritionValue
	}
	return nil
}
