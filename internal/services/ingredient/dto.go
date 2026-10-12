package ingredient

type CreateIngredientRequest struct {
	UserID      int
	Name        string
	KcalPer100g float64
	Fats        *float64
	Carbs       *float64
	Proteins    *float64
}

type UpdateIngredientRequest struct {
	IngredientID int
	UserID       int
	Name         string
	KcalPer100g  float64
	Fats         *float64
	Carbs        *float64
	Proteins     *float64
}

// PatchIngredientRequest changes some values of a user ingredient. Nil fields keep their current values.
type PatchIngredientRequest struct {
	IngredientID int
	UserID       int
	Name         *string
	KcalPer100g  *float64
	Fats         *float64
	Carbs        *float64
	Proteins     *float64
}
