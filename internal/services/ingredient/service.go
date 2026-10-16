package ingredient

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"ypeskov/kkal-tracker/internal/food"
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

// GetIngredientsWithUsage returns all user ingredients with their usage statistics in the food diary
func (s *Service) GetIngredientsWithUsage(userID int) ([]*models.UserIngredientUsage, error) {
	s.logger.Debug("GetIngredientsWithUsage called", "user_id", userID)

	ingredients, err := s.ingredientRepo.GetUserIngredientsWithUsage(userID)
	if err != nil {
		s.logger.Error("Failed to get user ingredients with usage", "error", err, "user_id", userID)
		return nil, err
	}

	s.logger.Debug("GetIngredientsWithUsage completed successfully", "user_id", userID, "count", len(ingredients))
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

	return s.createIngredient(req.UserID, req.Name, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins)
}

// AddIngredient creates an ingredient for the external API. Unlike CreateIngredient of the website, the values are
// checked against the sanity limits of package food and the name is cleaned, like a new ingredient of a meal.
// The result carries the usage statistics the ingredient list shows for it: the statistics are matched by name, so
// diary entries that already have exactly this name count.
func (s *Service) AddIngredient(req *AddIngredientRequest) (*models.UserIngredientUsage, error) {
	s.logger.Debug("AddIngredient called", "user_id", req.UserID)

	name, err := validateAddRequest(req)
	if err != nil {
		s.logger.Debug("AddIngredient failed - validation error", "user_id", req.UserID, "error", err)
		return nil, err
	}

	ingredient, err := s.createIngredient(req.UserID, name, *req.KcalPer100g, req.Fats, req.Carbs, req.Proteins)
	if err != nil {
		return nil, err
	}

	return s.withUsage(ingredient), nil
}

// withUsage returns a new ingredient with its row of the ingredient list. The ingredient is already stored, so
// a failure to read the statistics is logged and the ingredient is returned without them instead of as an error:
// an error would make the client retry a create that has succeeded.
func (s *Service) withUsage(ingredient *models.UserIngredient) *models.UserIngredientUsage {
	ingredients, err := s.ingredientRepo.GetUserIngredientsWithUsage(ingredient.UserID)
	if err != nil {
		s.logger.Error("Failed to get usage of the created ingredient", "error", err,
			"user_id", ingredient.UserID, "ingredient_id", ingredient.ID)
		return &models.UserIngredientUsage{UserIngredient: *ingredient}
	}
	for _, row := range ingredients {
		if row.ID == ingredient.ID {
			return row
		}
	}
	// Deleted by a concurrent request right after the create
	return &models.UserIngredientUsage{UserIngredient: *ingredient}
}

// createIngredient is the one create path of the ingredient list, for the website and the external API:
// the duplicate-name rule, the write, and the explanation of a failed write
func (s *Service) createIngredient(userID int, name string, kcalPer100g float64,
	fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	if err := s.checkDuplicateName(userID, name, nil); err != nil {
		return nil, err
	}

	ingredient, err := s.ingredientRepo.CreateUserIngredient(userID, name, kcalPer100g, fats, carbs, proteins)
	if err != nil {
		if duplicateErr := s.duplicateAfterWriteError(userID, name, 0, err); duplicateErr != nil {
			return nil, duplicateErr
		}
		s.logger.Error("Failed to create user ingredient", "error", err, "user_id", userID, "name", name)
		return nil, err
	}

	s.logger.Info("User ingredient created", "user_id", userID, "ingredient_id", ingredient.ID,
		"name", ingredient.Name, "kcal_per_100g", ingredient.KcalPer100g)
	return ingredient, nil
}

// UpdateIngredient replaces all values of a user ingredient (nil nutrients are cleared).
// A missing or foreign ingredient is sql.ErrNoRows.
func (s *Service) UpdateIngredient(req *UpdateIngredientRequest) (*models.UserIngredient, error) {
	s.logger.Debug("UpdateIngredient called", "user_id", req.UserID, "ingredient_id", req.IngredientID, "name", req.Name, "kcal_per_100g", req.KcalPer100g)

	if err := s.validateUpdateRequest(req); err != nil {
		s.logger.Debug("UpdateIngredient failed - validation error", "user_id", req.UserID, "ingredient_id", req.IngredientID, "error", err)
		return nil, err
	}

	current, err := s.getOwnIngredient(req.UserID, req.IngredientID)
	if err != nil {
		return nil, err
	}

	return s.saveIngredient(current, req.Name, req.KcalPer100g, req.Fats, req.Carbs, req.Proteins)
}

// PatchIngredient changes only the sent values of a user ingredient. Sent values are checked against the sanity
// limits of package food, the name is cleaned like the name of a new ingredient of a meal. Values that are not sent
// are kept as they are, even outside the limits. A missing or foreign ingredient is sql.ErrNoRows.
func (s *Service) PatchIngredient(req *PatchIngredientRequest) (*models.UserIngredient, error) {
	s.logger.Debug("PatchIngredient called", "user_id", req.UserID, "ingredient_id", req.IngredientID)

	name, err := validatePatchRequest(req)
	if err != nil {
		s.logger.Debug("PatchIngredient failed - validation error", "user_id", req.UserID, "ingredient_id", req.IngredientID, "error", err)
		return nil, err
	}

	current, err := s.getOwnIngredient(req.UserID, req.IngredientID)
	if err != nil {
		return nil, err
	}

	if req.Name == nil {
		name = current.Name
	}
	kcalPer100g := current.KcalPer100g
	if req.KcalPer100g != nil {
		kcalPer100g = *req.KcalPer100g
	}
	fats, carbs, proteins := current.Fats, current.Carbs, current.Proteins
	if req.Fats != nil {
		fats = req.Fats
	}
	if req.Carbs != nil {
		carbs = req.Carbs
	}
	if req.Proteins != nil {
		proteins = req.Proteins
	}

	return s.saveIngredient(current, name, kcalPer100g, fats, carbs, proteins)
}

// getOwnIngredient reads the ingredient that is about to change. A missing or foreign ingredient is sql.ErrNoRows.
func (s *Service) getOwnIngredient(userID, ingredientID int) (*models.UserIngredient, error) {
	ingredient, err := s.ingredientRepo.GetUserIngredientByID(userID, ingredientID)
	if errors.Is(err, sql.ErrNoRows) {
		s.logger.Debug("Ingredient not found", "user_id", userID, "ingredient_id", ingredientID)
		return nil, sql.ErrNoRows
	}
	if err != nil {
		s.logger.Error("Failed to get user ingredient by ID", "error", err, "user_id", userID, "ingredient_id", ingredientID)
		return nil, err
	}
	return ingredient, nil
}

// saveIngredient writes the new values of an existing ingredient after the duplicate-name check
func (s *Service) saveIngredient(current *models.UserIngredient, name string, kcalPer100g float64,
	fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	if err := s.checkDuplicateName(current.UserID, name, current); err != nil {
		return nil, err
	}

	ingredient, err := s.ingredientRepo.UpdateUserIngredient(current.UserID, current.ID, name, kcalPer100g, fats, carbs, proteins)
	if errors.Is(err, sql.ErrNoRows) {
		// Deleted between the read and the write
		return nil, sql.ErrNoRows
	}
	if err != nil {
		if duplicateErr := s.duplicateAfterWriteError(current.UserID, name, current.ID, err); duplicateErr != nil {
			return nil, duplicateErr
		}
		s.logger.Error("Failed to update user ingredient", "error", err, "user_id", current.UserID, "ingredient_id", current.ID)
		return nil, err
	}

	s.logger.Info("User ingredient updated", "user_id", current.UserID, "ingredient_id", current.ID,
		"name", ingredient.Name, "kcal_per_100g", ingredient.KcalPer100g)
	return ingredient, nil
}

// checkDuplicateName is the duplicate-name rule of the ingredient list: a name must not match (normalized) the name
// of another ingredient of the same user. current is nil for a new ingredient. An existing ingredient is checked only
// when its normalized name changes, so that a case-only rename, or editing an ingredient that already has
// a near-duplicate from before the rule, is allowed.
func (s *Service) checkDuplicateName(userID int, name string, current *models.UserIngredient) error {
	excludeID := 0
	if current != nil {
		if food.NormalizeName(current.Name) == food.NormalizeName(name) {
			return nil
		}
		excludeID = current.ID
	}

	normalized := food.NormalizeName(name)
	existing, err := s.findOtherIngredient(userID, excludeID, func(other string) bool {
		return food.NormalizeName(other) == normalized
	})
	if err != nil {
		return err
	}
	if existing != nil {
		s.logger.Debug("Duplicate ingredient name", "user_id", userID, "ingredient_id", excludeID, "existing_ingredient_id", existing.ID)
		return &DuplicateIngredientError{Existing: existing}
	}
	return nil
}

// duplicateAfterWriteError explains a failed write: when another ingredient has exactly the written name (it was
// added concurrently after the check, or it is a near-duplicate from before the rule), the failure is the unique index
// on the name, which compares the exact text, and is reported as a duplicate. Returns nil otherwise: a near-duplicate
// cannot make the write fail, so any other failure stays an internal error.
func (s *Service) duplicateAfterWriteError(userID int, name string, excludeID int, writeErr error) error {
	existing, err := s.findOtherIngredient(userID, excludeID, func(other string) bool { return other == name })
	if err != nil || existing == nil {
		return nil
	}
	s.logger.Debug("Ingredient write failed on a duplicate name", "error", writeErr, "user_id", userID,
		"ingredient_id", excludeID, "existing_ingredient_id", existing.ID)
	return &DuplicateIngredientError{Existing: existing}
}

// findOtherIngredient returns an ingredient of the user other than excludeID whose name matches, or nil
func (s *Service) findOtherIngredient(userID, excludeID int, matches func(name string) bool) (*models.UserIngredient, error) {
	ingredients, err := s.ingredientRepo.GetAllUserIngredients(userID)
	if err != nil {
		s.logger.Error("Failed to get user ingredients", "error", err, "user_id", userID)
		return nil, err
	}

	for _, ingredient := range ingredients {
		if ingredient.ID != excludeID && matches(ingredient.Name) {
			return ingredient, nil
		}
	}
	return nil, nil
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

// validatePatchRequest checks the sent values and returns the cleaned name (empty when the name is not sent)
func validatePatchRequest(req *PatchIngredientRequest) (string, error) {
	if req.Name == nil && req.KcalPer100g == nil && req.Fats == nil && req.Carbs == nil && req.Proteins == nil {
		return "", &ValidationError{Message: "nothing to change: send name, kcal_per_100g, fats, carbs or proteins"}
	}

	var name string
	if req.Name != nil {
		var err error
		if name, err = cleanName(*req.Name, "name must not be empty"); err != nil {
			return "", err
		}
	}
	if err := validateValues(req.KcalPer100g, req.Fats, req.Carbs, req.Proteins); err != nil {
		return "", err
	}
	return name, nil
}

// validateAddRequest checks a new ingredient of the external API and returns its cleaned name
func validateAddRequest(req *AddIngredientRequest) (string, error) {
	const nameRequired = "name is required"
	if req.Name == nil {
		return "", &ValidationError{Message: nameRequired}
	}
	name, err := cleanName(*req.Name, nameRequired)
	if err != nil {
		return "", err
	}
	if req.KcalPer100g == nil {
		return "", &ValidationError{Message: "kcal_per_100g is required"}
	}
	if err := validateValues(req.KcalPer100g, req.Fats, req.Carbs, req.Proteins); err != nil {
		return "", err
	}
	return name, nil
}

// cleanName cleans a sent ingredient name and checks its length; emptyMessage is the error of a name that is empty
// after cleaning
func cleanName(raw, emptyMessage string) (string, error) {
	name := food.CleanName(raw)
	if name == "" {
		return "", &ValidationError{Message: emptyMessage}
	}
	if food.NameTooLong(name) {
		return "", &ValidationError{Message: fmt.Sprintf("name must be at most %d characters", food.MaxNameLength)}
	}
	return name, nil
}

// validateValues checks the sent (non-nil) calorie and nutrient values against the sanity limits of package food
func validateValues(kcalPer100g, fats, carbs, proteins *float64) error {
	if kcalPer100g != nil && !food.ValidKcalPer100g(*kcalPer100g) {
		return &ValidationError{Message: fmt.Sprintf("kcal_per_100g must be between 0 and %g", food.MaxKcalPer100g)}
	}
	nutrients := []struct {
		field string
		value *float64
	}{{"fats", fats}, {"carbs", carbs}, {"proteins", proteins}}
	for _, n := range nutrients {
		if n.value != nil && !food.ValidNutrientPer100g(*n.value) {
			return &ValidationError{Message: fmt.Sprintf("%s must be between 0 and %g (grams per 100 g)", n.field, food.MaxNutrientPer100g)}
		}
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
