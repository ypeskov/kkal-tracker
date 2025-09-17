package calories

type CreateEntryRequest struct {
	Food         string   `json:"food" validate:"required"`
	Calories     int      `json:"calories" validate:"required,min=1"`
	Weight       float64  `json:"weight" validate:"required,min=0.1"`
	KcalPer100g  float64  `json:"kcalPer100g" validate:"required,min=0.1"`
	Fats         *float64 `json:"fats,omitempty"`
	Carbs        *float64 `json:"carbs,omitempty"`
	Proteins     *float64 `json:"proteins,omitempty"`
	MealDatetime string   `json:"meal_datetime" validate:"required"`
}
