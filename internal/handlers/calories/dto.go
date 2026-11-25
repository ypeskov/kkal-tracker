package calories

// CreateEntryRequest is a diary entry of the website (create and update). Fiber (grams per 100 g) is the only
// nutrient that is range-checked: 0 is a value, an absent field means unknown.
type CreateEntryRequest struct {
	Food         string   `json:"food" validate:"required"`
	Calories     int      `json:"calories" validate:"required,min=1"`
	Weight       float64  `json:"weight" validate:"required,min=0.1"`
	KcalPer100g  float64  `json:"kcalPer100g" validate:"required,min=0.1"`
	Fats         *float64 `json:"fats,omitempty"`
	Carbs        *float64 `json:"carbs,omitempty"`
	Proteins     *float64 `json:"proteins,omitempty"`
	Fiber        *float64 `json:"fiber,omitempty" validate:"omitempty,min=0,max=100"`
	MealDatetime string   `json:"meal_datetime" validate:"required"`
}
