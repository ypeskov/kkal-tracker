package apidata

// IngredientInfo is a user ingredient as seen by API clients
type IngredientInfo struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	KcalPer100g float64  `json:"kcal_per_100g"`
	Fats        *float64 `json:"fats,omitempty"`
	Carbs       *float64 `json:"carbs,omitempty"`
	Proteins    *float64 `json:"proteins,omitempty"`
}

// IngredientItem is an ingredient of the list together with its usage statistics
type IngredientItem struct {
	IngredientInfo
	// TimesUsed and LastUsed (YYYY-MM-DD) help a client choose between similar ingredients
	TimesUsed int     `json:"times_used"`
	LastUsed  *string `json:"last_used,omitempty"`
}

type IngredientsResponse struct {
	Ingredients []IngredientItem `json:"ingredients"`
}

// CreateFoodRequest is one meal: several foods eaten at the same time
type CreateFoodRequest struct {
	// MealDatetime is optional (RFC 3339), the current time is used when it is empty
	MealDatetime string            `json:"meal_datetime"`
	Items        []FoodItemRequest `json:"items" validate:"required,min=1,max=50"`
}

// FoodItemRequest refers to an existing ingredient (ingredient_id) or describes a new food (new_ingredient)
type FoodItemRequest struct {
	IngredientID  *int                  `json:"ingredient_id"`
	NewIngredient *NewIngredientRequest `json:"new_ingredient"`
	// OneOff writes a new food to the diary without adding it to the ingredient list
	OneOff bool    `json:"one_off"`
	Weight float64 `json:"weight"`
	// KcalPer100g overrides the calorie value of an existing ingredient for this entry only
	KcalPer100g *float64 `json:"kcal_per_100g"`
}

type NewIngredientRequest struct {
	Name        string   `json:"name"`
	KcalPer100g *float64 `json:"kcal_per_100g"`
	Fats        *float64 `json:"fats"`
	Carbs       *float64 `json:"carbs"`
	Proteins    *float64 `json:"proteins"`
}

type CreatedFoodEntry struct {
	FoodEntry
	// IngredientID is absent for one-off foods
	IngredientID      *int `json:"ingredient_id,omitempty"`
	IngredientCreated bool `json:"ingredient_created"`
}

type CreateFoodResponse struct {
	Entries       []CreatedFoodEntry `json:"entries"`
	TotalCalories int                `json:"total_calories"`
	// Day is the diary day (YYYY-MM-DD, UTC) the meal belongs to
	Day              string `json:"day"`
	DayTotalCalories *int   `json:"day_total_calories,omitempty"`
}

// DuplicateIngredientResponse tells the client which existing ingredient to use instead of creating a new one
type DuplicateIngredientResponse struct {
	Message            string         `json:"message"`
	ItemIndex          int            `json:"item_index"`
	ExistingIngredient IngredientInfo `json:"existing_ingredient"`
}
