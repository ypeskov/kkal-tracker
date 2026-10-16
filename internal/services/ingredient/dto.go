package ingredient

type CreateIngredientRequest struct {
	UserID      int
	Name        string
	KcalPer100g float64
	Fats        *float64
	Carbs       *float64
	Proteins    *float64
}

// AddIngredientRequest is a new ingredient of the external API. Nil fields were not sent (or sent as null):
// the name and the calories are required, unknown nutrients stay nil.
type AddIngredientRequest struct {
	UserID      int
	Name        *string
	KcalPer100g *float64
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
