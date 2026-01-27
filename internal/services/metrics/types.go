package metrics

// HealthMetrics represents calculated health metrics for a user
type HealthMetrics struct {
	BMI          *float64 `json:"bmi,omitempty"`           // Body Mass Index
	BMICategory  string   `json:"bmi_category,omitempty"`  // BMI category (underweight, normal, overweight, obese)
	BMR          *float64 `json:"bmr,omitempty"`           // Basal Metabolic Rate (calories/day)
	TDEE         *float64 `json:"tdee,omitempty"`          // Total Daily Energy Expenditure (optional)
	HealthStatus string   `json:"health_status,omitempty"` // Overall health status message
}

// ActivityLevel represents different activity levels for TDEE calculation
type ActivityLevel string

const (
	ActivitySedentary     ActivityLevel = "sedentary"      // Little or no exercise
	ActivityLightlyActive ActivityLevel = "lightly_active" // Exercise 1-3 times/week
	ActivityModerate      ActivityLevel = "moderate"       // Exercise 3-5 times/week
	ActivityVeryActive    ActivityLevel = "very_active"    // Exercise 6-7 times/week
	ActivityExtraActive   ActivityLevel = "extra_active"   // Very hard exercise & physical job
)

// ActivityMultipliers maps activity levels to their multipliers for TDEE calculation
var ActivityMultipliers = map[ActivityLevel]float64{
	ActivitySedentary:     1.2,
	ActivityLightlyActive: 1.375,
	ActivityModerate:      1.55,
	ActivityVeryActive:    1.725,
	ActivityExtraActive:   1.9,
}
