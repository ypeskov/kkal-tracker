package calorie

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

const testUserID = 7

func ptr(v float64) *float64 { return &v }

func value(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

// fakeCalorieRepo keeps diary entries in memory and records the writes.
// Methods the service does not use come from the embedded nil interface and panic when called.
type fakeCalorieRepo struct {
	repositories.CalorieEntryRepository

	entries map[int]*models.CalorieEntry
	// meals are the entries passed to CreateMeal, updates the number of Update calls
	meals   [][]models.NewMealEntry
	updates int
}

func (r *fakeCalorieRepo) CreateMeal(userID int, entries []models.NewMealEntry, mealDatetime time.Time) ([]models.CreatedMealEntry, error) {
	r.meals = append(r.meals, entries)
	created := make([]models.CreatedMealEntry, len(entries))
	for i, e := range entries {
		created[i] = models.CreatedMealEntry{Entry: &models.CalorieEntry{
			ID: 100 + i, UserID: userID, Food: e.Food, Calories: e.Calories, Weight: e.Weight, KcalPer100g: e.KcalPer100g,
			Fats: e.Fats, Carbs: e.Carbs, Proteins: e.Proteins, MealDatetime: mealDatetime,
		}}
	}
	return created, nil
}

func (r *fakeCalorieRepo) GetByID(id int) (*models.CalorieEntry, error) {
	entry, ok := r.entries[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	copied := *entry
	return &copied, nil
}

func (r *fakeCalorieRepo) Update(id, userID int, food string, calories int, weight float64, kcalPer100g float64,
	fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error) {
	r.updates++
	entry := &models.CalorieEntry{ID: id, UserID: userID, Food: food, Calories: calories, Weight: weight,
		KcalPer100g: kcalPer100g, Fats: fats, Carbs: carbs, Proteins: proteins, MealDatetime: mealDatetime}
	r.entries[id] = entry
	return entry, nil
}

func (r *fakeCalorieRepo) GetByUserIDAndDateRange(int, string, string) ([]*models.CalorieEntry, error) {
	return nil, nil
}

// fakeIngredientRepo returns the ingredients of the test user and fails the test on any write
type fakeIngredientRepo struct {
	repositories.IngredientRepository

	ingredients map[int]*models.UserIngredient
}

func (r *fakeIngredientRepo) GetAllUserIngredients(int) ([]*models.UserIngredient, error) {
	var all []*models.UserIngredient
	for _, ingredient := range r.ingredients {
		all = append(all, ingredient)
	}
	return all, nil
}

func (r *fakeIngredientRepo) GetUserIngredientByID(userID, ingredientID int) (*models.UserIngredient, error) {
	ingredient, ok := r.ingredients[ingredientID]
	if !ok || ingredient.UserID != userID {
		return nil, sql.ErrNoRows
	}
	return ingredient, nil
}

// Bread (250 kcal, F 1, C 50, P 8) and butter (748 kcal, F 82, unknown carbs, P 0.5)
func newTestService() (*Service, *fakeCalorieRepo, *fakeIngredientRepo) {
	calories := &fakeCalorieRepo{entries: map[int]*models.CalorieEntry{
		1: {ID: 1, UserID: testUserID, Food: "Хлеб", Calories: 250, Weight: 100, KcalPer100g: 250,
			Fats: ptr(1), Carbs: ptr(50), Proteins: ptr(8), MealDatetime: time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC)},
	}}
	ingredients := &fakeIngredientRepo{ingredients: map[int]*models.UserIngredient{
		94: {ID: 94, UserID: testUserID, Name: "Хлеб", KcalPer100g: 250, Fats: ptr(1), Carbs: ptr(50), Proteins: ptr(8)},
		95: {ID: 95, UserID: testUserID, Name: "Масло", KcalPer100g: 748, Fats: ptr(82), Proteins: ptr(0.5)},
	}}
	return New(calories, ingredients, slog.New(slog.NewTextHandler(io.Discard, nil))), calories, ingredients
}

func checkIngredientsUnchanged(t *testing.T, ingredients *fakeIngredientRepo) {
	t.Helper()
	bread, butter := ingredients.ingredients[94], ingredients.ingredients[95]
	if value(bread.Fats) != 1.0 || value(bread.Carbs) != 50.0 || value(bread.Proteins) != 8.0 || bread.KcalPer100g != 250 {
		t.Errorf("bread changed: %+v", bread)
	}
	if value(butter.Fats) != 82.0 || butter.Carbs != nil || value(butter.Proteins) != 0.5 {
		t.Errorf("butter changed: %+v", butter)
	}
}

func intPtr(v int) *int { return &v }

func TestCreateMealOverridesNutrientsOfAnEntry(t *testing.T) {
	s, calories, ingredients := newTestService()

	result, err := s.CreateMeal(&CreateMealRequest{UserID: testUserID, Items: []MealItem{
		{IngredientID: intPtr(94), Weight: 100, KcalPer100g: ptr(260), Fats: ptr(3), Carbs: ptr(48), Proteins: ptr(9)},
		{IngredientID: intPtr(94), Weight: 100, Fats: ptr(3)},
		{IngredientID: intPtr(95), Weight: 10, Carbs: ptr(0)},
		{IngredientID: intPtr(94), Weight: 100, Fats: ptr(0), Carbs: ptr(100)},
	}})
	if err != nil {
		t.Fatalf("CreateMeal: %v", err)
	}

	want := []struct {
		kcal                  float64
		fats, carbs, proteins any
		calories              int
	}{
		{260, 3.0, 48.0, 9.0, 260},
		{250, 3.0, 50.0, 8.0, 250},
		{748, 82.0, 0.0, 0.5, 75},
		{250, 0.0, 100.0, 8.0, 250},
	}
	for i, w := range want {
		e := result.Entries[i].Entry
		if e.KcalPer100g != w.kcal || value(e.Fats) != w.fats || value(e.Carbs) != w.carbs || value(e.Proteins) != w.proteins ||
			e.Calories != w.calories {
			t.Errorf("entry %d = %g kcal/100g, F %v, C %v, P %v, %d kcal; want %+v",
				i, e.KcalPer100g, value(e.Fats), value(e.Carbs), value(e.Proteins), e.Calories, w)
		}
	}
	if len(calories.meals) != 1 {
		t.Errorf("CreateMeal calls = %d, want 1", len(calories.meals))
	}
	checkIngredientsUnchanged(t, ingredients)
}

func TestCreateMealRejectsInvalidNutrientOverrides(t *testing.T) {
	newFood := &NewIngredient{Name: "Батон", KcalPer100g: 260}
	tests := []struct {
		item MealItem
		want string
	}{
		{MealItem{IngredientID: intPtr(94), Weight: 100, Fats: ptr(-1)}, "items[1]: fats must be between 0 and 100 (grams per 100 g)"},
		{MealItem{IngredientID: intPtr(94), Weight: 100, Carbs: ptr(100.01)}, "items[1]: carbs must be between 0 and 100 (grams per 100 g)"},
		{MealItem{IngredientID: intPtr(94), Weight: 100, Fats: ptr(3), Proteins: ptr(100.01)},
			"items[1]: proteins must be between 0 and 100 (grams per 100 g)"},
		{MealItem{NewIngredient: newFood, Weight: 100, Fats: ptr(3)}, "items[1]: fats of a new food belongs inside new_ingredient"},
		{MealItem{NewIngredient: newFood, Weight: 100, Carbs: ptr(3), OneOff: true}, "items[1]: carbs of a new food belongs inside new_ingredient"},
		{MealItem{NewIngredient: newFood, Weight: 100, Proteins: ptr(3)}, "items[1]: proteins of a new food belongs inside new_ingredient"},
	}
	for _, tt := range tests {
		s, calories, ingredients := newTestService()
		_, err := s.CreateMeal(&CreateMealRequest{UserID: testUserID, Items: []MealItem{
			{IngredientID: intPtr(94), Weight: 50}, tt.item,
		}})

		var itemErr *MealItemError
		if !errors.As(err, &itemErr) || itemErr.Index != 1 || err.Error() != tt.want {
			t.Errorf("error = %v, want %q", err, tt.want)
		}
		if len(calories.meals) != 0 {
			t.Errorf("%s: the meal was stored", tt.want)
		}
		checkIngredientsUnchanged(t, ingredients)
	}
}

func TestUpdateMealEntryOverridesNutrients(t *testing.T) {
	tests := []struct {
		name                  string
		req                   UpdateMealEntryRequest
		food                  string
		kcal                  float64
		fats, carbs, proteins any
		calories              int
	}{
		{"one macro", UpdateMealEntryRequest{Proteins: ptr(20)}, "Хлеб", 250, 1.0, 50.0, 20.0, 250},
		{"all macros", UpdateMealEntryRequest{Fats: ptr(3), Carbs: ptr(48), Proteins: ptr(9)}, "Хлеб", 250, 3.0, 48.0, 9.0, 250},
		{"ingredient only", UpdateMealEntryRequest{IngredientID: intPtr(95)}, "Масло", 748, 82.0, nil, 0.5, 748},
		{"ingredient and an override", UpdateMealEntryRequest{IngredientID: intPtr(95), Carbs: ptr(0)}, "Масло", 748, 82.0, 0.0, 0.5, 748},
		{"ingredient, kcal and an override", UpdateMealEntryRequest{IngredientID: intPtr(95), KcalPer100g: ptr(700), Fats: ptr(80)},
			"Масло", 700, 80.0, nil, 0.5, 700},
		{"limits are inclusive", UpdateMealEntryRequest{Fats: ptr(0), Carbs: ptr(100)}, "Хлеб", 250, 0.0, 100.0, 8.0, 250},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, calories, ingredients := newTestService()
			req := tt.req
			req.EntryID, req.UserID = 1, testUserID

			result, err := s.UpdateMealEntry(&req)
			if err != nil {
				t.Fatalf("UpdateMealEntry: %v", err)
			}
			e := result.Entry
			if e.Food != tt.food || e.KcalPer100g != tt.kcal || value(e.Fats) != tt.fats || value(e.Carbs) != tt.carbs ||
				value(e.Proteins) != tt.proteins || e.Calories != tt.calories || e.Weight != 100 {
				t.Errorf("entry = %s %g g, %g kcal/100g, F %v, C %v, P %v, %d kcal", e.Food, e.Weight, e.KcalPer100g,
					value(e.Fats), value(e.Carbs), value(e.Proteins), e.Calories)
			}
			if result.Day != "2026-03-10" || calories.updates != 1 {
				t.Errorf("day = %q, updates = %d; want 2026-03-10 and one update", result.Day, calories.updates)
			}
			checkIngredientsUnchanged(t, ingredients)
		})
	}
}

func TestUpdateMealEntryRejectsInvalidNutrients(t *testing.T) {
	tests := []struct {
		req  UpdateMealEntryRequest
		want string
	}{
		{UpdateMealEntryRequest{}, "nothing to change: send weight, ingredient_id, kcal_per_100g, fats, carbs, proteins or meal_datetime"},
		{UpdateMealEntryRequest{Carbs: ptr(101)}, "carbs must be between 0 and 100 (grams per 100 g)"},
		{UpdateMealEntryRequest{Fats: ptr(-0.5), Weight: ptr(50)}, "fats must be between 0 and 100 (grams per 100 g)"},
		{UpdateMealEntryRequest{IngredientID: intPtr(95), Proteins: ptr(100.01)}, "proteins must be between 0 and 100 (grams per 100 g)"},
	}
	for _, tt := range tests {
		s, calories, ingredients := newTestService()
		req := tt.req
		req.EntryID, req.UserID = 1, testUserID

		_, err := s.UpdateMealEntry(&req)
		var validationErr *EntryValidationError
		if !errors.As(err, &validationErr) || err.Error() != tt.want {
			t.Errorf("error = %v, want %q", err, tt.want)
		}
		if calories.updates != 0 {
			t.Errorf("%s: the entry was updated", tt.want)
		}
		checkIngredientsUnchanged(t, ingredients)
	}
}
