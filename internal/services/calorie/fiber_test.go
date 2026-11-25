package calorie

import (
	"errors"
	"testing"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

const fiberLimitMessage = "fiber must be between 0 and 100 (grams per 100 g)"

// Oats (fibre 10.6) next to the ingredients of newTestService, whose fibre is unknown
func newFiberTestService() (*Service, *fakeCalorieRepo, *fakeIngredientRepo) {
	s, calories, ingredients := newTestService()
	ingredients.ingredients[96] = &models.UserIngredient{ID: 96, UserID: testUserID, Name: "Овсянка", KcalPer100g: 370, Fiber: ptr(10.6)}
	return s, calories, ingredients
}

func TestCreateMealCopiesAndOverridesFiber(t *testing.T) {
	s, calories, ingredients := newFiberTestService()

	result, err := s.CreateMeal(&CreateMealRequest{UserID: testUserID, Items: []MealItem{
		{IngredientID: intPtr(96), Weight: 50},
		{IngredientID: intPtr(96), Weight: 50, Fiber: ptr(8)},
		{IngredientID: intPtr(94), Weight: 100},
		{IngredientID: intPtr(94), Weight: 100, Fiber: ptr(0)},
		{IngredientID: intPtr(94), Weight: 100, Fiber: ptr(100)},
		{NewIngredient: &NewIngredient{Name: "Груша", KcalPer100g: 57, Fiber: ptr(3.1)}, Weight: 100},
		{NewIngredient: &NewIngredient{Name: "Салат", KcalPer100g: 40, Fiber: ptr(0)}, Weight: 100, OneOff: true},
		{NewIngredient: &NewIngredient{Name: "Вода", KcalPer100g: 0}, Weight: 100, OneOff: true},
	}})
	if err != nil {
		t.Fatalf("CreateMeal: %v", err)
	}

	want := []any{10.6, 8.0, nil, 0.0, 100.0, 3.1, 0.0, nil}
	for i, w := range want {
		if got := value(result.Entries[i].Entry.Fiber); got != w {
			t.Errorf("entry %d: fiber = %v, want %v", i, got, w)
		}
	}
	// The fibre of a new food is stored with the food: the repository saves it on the entry and on the ingredient
	stored := calories.meals[0]
	if value(stored[5].Fiber) != 3.1 || !stored[5].SaveAsIngredient || value(stored[6].Fiber) != 0.0 || stored[6].SaveAsIngredient {
		t.Errorf("new foods passed to the repository = %+v, %+v", stored[5], stored[6])
	}
	// Calories do not depend on fibre
	if result.TotalCalories != 185+185+250+250+250+57+40 {
		t.Errorf("total calories = %d", result.TotalCalories)
	}
	// An override never changes the ingredient
	if value(ingredients.ingredients[96].Fiber) != 10.6 || ingredients.ingredients[94].Fiber != nil {
		t.Errorf("ingredients changed: %+v, %+v", ingredients.ingredients[96], ingredients.ingredients[94])
	}
	checkIngredientsUnchanged(t, ingredients)
}

func TestCreateMealRejectsInvalidFiber(t *testing.T) {
	tests := []struct {
		item MealItem
		want string
	}{
		{MealItem{IngredientID: intPtr(96), Weight: 100, Fiber: ptr(101)}, "items[1]: " + fiberLimitMessage},
		{MealItem{IngredientID: intPtr(96), Weight: 100, Fiber: ptr(-0.1)}, "items[1]: " + fiberLimitMessage},
		{MealItem{NewIngredient: &NewIngredient{Name: "X", KcalPer100g: 100, Fiber: ptr(101)}, Weight: 100},
			"items[1]: new_ingredient." + fiberLimitMessage},
		{MealItem{NewIngredient: &NewIngredient{Name: "X", KcalPer100g: 100, Fiber: ptr(-1)}, Weight: 100, OneOff: true},
			"items[1]: new_ingredient." + fiberLimitMessage},
		{MealItem{NewIngredient: &NewIngredient{Name: "X", KcalPer100g: 100}, Weight: 100, Fiber: ptr(3)},
			"items[1]: fiber of a new food belongs inside new_ingredient"},
		{MealItem{NewIngredient: &NewIngredient{Name: "X", KcalPer100g: 100}, Weight: 100, Fiber: ptr(0), OneOff: true},
			"items[1]: fiber of a new food belongs inside new_ingredient"},
		// The order of the checks is fats, carbs, proteins, fiber
		{MealItem{IngredientID: intPtr(96), Weight: 100, Proteins: ptr(101), Fiber: ptr(101)},
			"items[1]: proteins must be between 0 and 100 (grams per 100 g)"},
		{MealItem{NewIngredient: &NewIngredient{Name: "X", KcalPer100g: 100}, Weight: 100, Proteins: ptr(1), Fiber: ptr(1)},
			"items[1]: proteins of a new food belongs inside new_ingredient"},
	}
	for _, tt := range tests {
		s, calories, _ := newFiberTestService()
		_, err := s.CreateMeal(&CreateMealRequest{UserID: testUserID, Items: []MealItem{
			{IngredientID: intPtr(96), Weight: 50}, tt.item,
		}})

		var itemErr *MealItemError
		if !errors.As(err, &itemErr) || itemErr.Index != 1 || err.Error() != tt.want {
			t.Errorf("error = %v, want %q", err, tt.want)
		}
		if len(calories.meals) != 0 {
			t.Errorf("%s: the meal was stored", tt.want)
		}
	}
}

func TestUpdateMealEntryFiber(t *testing.T) {
	tests := []struct {
		name string
		// start is the fibre of the entry before the change
		start *float64
		req   UpdateMealEntryRequest
		food  string
		want  any
	}{
		{"fiber only", nil, UpdateMealEntryRequest{Fiber: ptr(4)}, "Хлеб", 4.0},
		{"zero is a value", ptr(4), UpdateMealEntryRequest{Fiber: ptr(0)}, "Хлеб", 0.0},
		{"the upper limit is inclusive", nil, UpdateMealEntryRequest{Fiber: ptr(100)}, "Хлеб", 100.0},
		{"weight keeps the fiber", ptr(4), UpdateMealEntryRequest{Weight: ptr(120)}, "Хлеб", 4.0},
		{"a macro keeps the fiber", ptr(4), UpdateMealEntryRequest{Fats: ptr(2)}, "Хлеб", 4.0},
		{"ingredient copies the fiber", ptr(4), UpdateMealEntryRequest{IngredientID: intPtr(96)}, "Овсянка", 10.6},
		{"override on top of the ingredient", nil, UpdateMealEntryRequest{IngredientID: intPtr(96), Fiber: ptr(7)}, "Овсянка", 7.0},
		{"ingredient with unknown fiber clears it", ptr(4), UpdateMealEntryRequest{IngredientID: intPtr(95)}, "Масло", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, calories, _ := newFiberTestService()
			calories.entries[1].Fiber = tt.start
			req := tt.req
			req.EntryID, req.UserID = 1, testUserID

			result, err := s.UpdateMealEntry(&req)
			if err != nil {
				t.Fatalf("UpdateMealEntry: %v", err)
			}
			if e := result.Entry; e.Food != tt.food || value(e.Fiber) != tt.want {
				t.Errorf("entry = %s with fiber %v, want %s with fiber %v", e.Food, value(e.Fiber), tt.food, tt.want)
			}
			// The macros of the bread are untouched unless they were sent or the ingredient changed
			if e := result.Entry; tt.req.IngredientID == nil && tt.req.Fats == nil &&
				(value(e.Fats) != 1.0 || value(e.Carbs) != 50.0 || value(e.Proteins) != 8.0 || e.KcalPer100g != 250) {
				t.Errorf("other values changed: %+v", e)
			}
		})
	}

	for _, tt := range []struct {
		req  UpdateMealEntryRequest
		want string
	}{
		{UpdateMealEntryRequest{Fiber: ptr(-0.1)}, fiberLimitMessage},
		{UpdateMealEntryRequest{Fiber: ptr(100.01), Weight: ptr(50)}, fiberLimitMessage},
		{UpdateMealEntryRequest{IngredientID: intPtr(96), Fiber: ptr(150)}, fiberLimitMessage},
	} {
		s, calories, _ := newFiberTestService()
		req := tt.req
		req.EntryID, req.UserID = 1, testUserID

		_, err := s.UpdateMealEntry(&req)
		var validationErr *EntryValidationError
		if !errors.As(err, &validationErr) || err.Error() != tt.want {
			t.Errorf("error = %v, want %q", err, tt.want)
		}
		if calories.updates != 0 {
			t.Errorf("%+v: the entry was updated", tt.req)
		}
	}
}

// The website entry API checks the range of the fibre only; the check runs before anything is read or written
// (the fakes panic on the calls CreateEntry would make next)
func TestWebsiteEntriesValidateFiber(t *testing.T) {
	when := time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)

	for _, fiber := range []float64{-1, -0.01, 100.01, 101} {
		s, calories, _ := newTestService()

		_, err := s.CreateEntry(&CreateEntryRequest{UserID: testUserID, Food: "Овсянка", Calories: 370, Weight: 100,
			KcalPer100g: 370, Fiber: ptr(fiber), MealDatetime: when})
		var validationErr *EntryValidationError
		if !errors.As(err, &validationErr) || err.Error() != fiberLimitMessage {
			t.Errorf("CreateEntry with fiber %g: error = %v, want %q", fiber, err, fiberLimitMessage)
		}

		_, err = s.UpdateEntry(&UpdateEntryRequest{EntryID: 1, UserID: testUserID, Food: "Хлеб", Calories: 250, Weight: 100,
			KcalPer100g: 250, Fiber: ptr(fiber), MealDatetime: when})
		if !errors.As(err, &validationErr) || err.Error() != fiberLimitMessage {
			t.Errorf("UpdateEntry with fiber %g: error = %v, want %q", fiber, err, fiberLimitMessage)
		}
		if calories.updates != 0 {
			t.Errorf("fiber %g: the entry was updated", fiber)
		}
	}

	// UpdateEntry replaces all values: the limits are inclusive, nil clears the fibre, the macros are not range-checked
	for _, tt := range []struct {
		fiber *float64
		want  any
	}{{ptr(0), 0.0}, {ptr(100), 100.0}, {ptr(6), 6.0}, {nil, nil}} {
		s, calories, _ := newTestService()
		calories.entries[1].Fiber = ptr(4)

		entry, err := s.UpdateEntry(&UpdateEntryRequest{EntryID: 1, UserID: testUserID, Food: "Хлеб", Calories: 250, Weight: 100,
			KcalPer100g: 250, Fats: ptr(150), Fiber: tt.fiber, MealDatetime: when})
		if err != nil {
			t.Fatalf("UpdateEntry with fiber %v: %v", value(tt.fiber), err)
		}
		if value(entry.Fiber) != tt.want || value(entry.Fats) != 150.0 {
			t.Errorf("UpdateEntry with fiber %v: entry fiber = %v, fats = %v", value(tt.fiber), value(entry.Fiber), value(entry.Fats))
		}
	}
}
