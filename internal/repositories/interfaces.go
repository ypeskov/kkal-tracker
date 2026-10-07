package repositories

import (
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

// UserRepository defines the contract for user data access
type UserRepository interface {
	Create(email, passwordHash string) (*models.User, error)
	CreateWithLanguage(email, passwordHash, languageCode string, isActive bool) (*models.User, error)
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	UpdateProfile(userID int, firstName, lastName *string, email string, age *int, height *float64, gender *string, weight *float64, language string, activityLevel *string) error
	AddWeightEntry(userID int, weight float64) error
	ActivateUser(userID int) error
	Delete(userID int) error
	// SetWeightGoal starts a goal: startedAt and initialWeight are the point the progress is measured from
	SetWeightGoal(userID int, targetWeight float64, targetDate *string, startedAt time.Time, initialWeight float64) error
	// UpdateWeightGoal changes the target of an existing goal, keeping its start date and initial weight
	UpdateWeightGoal(userID int, targetWeight float64, targetDate *string) error
	ClearWeightGoal(userID int) error
}

// CalorieEntryRepository defines the contract for calorie entry data access
type CalorieEntryRepository interface {
	Create(userID int, food string, calories int, weight float64, kcalPer100g float64,
		fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error)
	// CreateMeal stores all entries of one meal, and the new ingredients they bring, in a single transaction
	CreateMeal(userID int, entries []models.NewMealEntry, mealDatetime time.Time) ([]models.CreatedMealEntry, error)
	GetByID(id int) (*models.CalorieEntry, error)
	GetByUserID(userID int) ([]*models.CalorieEntry, error)
	GetByUserIDAndDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error)
	Update(id, userID int, food string, calories int, weight float64, kcalPer100g float64,
		fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error)
	Delete(id, userID int) error
}

// WeightHistoryRepository defines the contract for weight history data access
type WeightHistoryRepository interface {
	GetByUserID(userID int) ([]*models.WeightHistory, error)
	GetByUserIDAndDateRange(userID int, dateFrom, dateTo string) ([]*models.WeightHistory, error)
	GetLatestByUserID(userID int) (*models.WeightHistory, error)
	// GetLatestByUserIDOnOrBefore returns the last entry recorded on the given day (YYYY-MM-DD) or earlier, nil if there is none
	GetLatestByUserIDOnOrBefore(userID int, date string) (*models.WeightHistory, error)
	Create(userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error)
	Update(id, userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error)
	Delete(id, userID int) error
}

// IngredientRepository defines the contract for ingredient data access
type IngredientRepository interface {
	// User ingredients
	GetAllUserIngredients(userID int) ([]*models.UserIngredient, error)
	GetUserIngredientsWithUsage(userID int) ([]*models.UserIngredientUsage, error)
	GetUserIngredientByName(userID int, name string) (*models.UserIngredient, error)
	GetUserIngredientByID(userID int, ingredientID int) (*models.UserIngredient, error)
	CreateOrUpdateUserIngredient(userID int, name string, kcalPer100g float64,
		fats, carbs, proteins *float64) (*models.UserIngredient, error)
	CreateUserIngredient(userID int, name string, kcalPer100g float64,
		fats, carbs, proteins *float64) (*models.UserIngredient, error)
	UpdateUserIngredient(userID int, ingredientID int, name string, kcalPer100g float64,
		fats, carbs, proteins *float64) (*models.UserIngredient, error)
	DeleteUserIngredient(userID int, ingredientID int) error
	CopyGlobalIngredientsToUser(userID int, languageCode string) error

	// CreateGlobalIngredient Global ingredients (admin)
	CreateGlobalIngredient(kcalPer100g float64, fats, carbs, proteins *float64,
		names map[string]string) (*models.GlobalIngredient, error)
	GetGlobalIngredientByID(id int) (*models.GlobalIngredient, error)
}

// APIKeyRepository defines the contract for API key data access
type APIKeyRepository interface {
	Create(userID int, name, keyHash, keyPrefix string, expiresAt *time.Time) (*models.APIKey, error)
	GetByKeyHash(keyHash string) (*models.APIKey, error)
	GetByUserID(userID int) ([]*models.APIKey, error)
	Revoke(id, userID int) error
	Delete(id, userID int) error
}
