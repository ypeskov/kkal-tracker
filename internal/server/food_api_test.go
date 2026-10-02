package server

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"
)

type foodAPIIngredient struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	KcalPer100g float64 `json:"kcal_per_100g"`
	TimesUsed   int     `json:"times_used"`
	LastUsed    *string `json:"last_used"`
}

type foodAPIEntry struct {
	ID                int     `json:"id"`
	Food              string  `json:"food"`
	Calories          int     `json:"calories"`
	Weight            float64 `json:"weight"`
	KcalPer100g       float64 `json:"kcal_per_100g"`
	MealDatetime      string  `json:"meal_datetime"`
	IngredientID      *int    `json:"ingredient_id"`
	IngredientCreated bool    `json:"ingredient_created"`
}

type foodAPIMeal struct {
	Entries          []foodAPIEntry `json:"entries"`
	TotalCalories    int            `json:"total_calories"`
	Day              string         `json:"day"`
	DayTotalCalories *int           `json:"day_total_calories"`
}

// signUp registers and activates a user and authorizes the client with the user's JWT and a new API key
func signUp(t *testing.T, client *smokeClient, smtpServer *fakeSMTPServer, email string) {
	t.Helper()

	const password = "smoke-password"
	client.token, client.apiKey = "", ""

	client.expect(http.StatusCreated, http.MethodPost, "/api/auth/register",
		map[string]string{"email": email, "password": password, "language_code": "en_US"}, nil)
	match := regexp.MustCompile(`/activate/([A-Za-z0-9_-]+)`).FindStringSubmatch(string(smtpServer.lastMessage(t)))
	if match == nil {
		t.Fatal("activation email does not contain an activation link")
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/auth/activate/"+match[1], nil, nil)

	var login struct {
		Token string `json:"token"`
	}
	client.expect(http.StatusOK, http.MethodPost, "/api/auth/login",
		map[string]string{"email": email, "password": password}, &login)
	client.token = login.Token

	var apiKey struct {
		Key string `json:"key"`
	}
	client.expect(http.StatusCreated, http.MethodPost, "/api/api-keys", map[string]any{"name": "agent"}, &apiKey)
	if apiKey.Key == "" {
		t.Fatal("API key was not returned")
	}
	client.apiKey = apiKey.Key
}

// addIngredient creates an ingredient through the web API and returns its ID
func addIngredient(t *testing.T, client *smokeClient, name string, kcalPer100g float64) int {
	t.Helper()

	var ingredient struct {
		ID int `json:"id"`
	}
	client.expect(http.StatusCreated, http.MethodPost, "/api/ingredients",
		map[string]any{"name": name, "kcalPer100g": kcalPer100g, "fats": 0.2, "carbs": 4.2, "proteins": 0.6}, &ingredient)
	if ingredient.ID == 0 {
		t.Fatalf("ingredient %q was not created", name)
	}
	return ingredient.ID
}

func listIngredients(t *testing.T, client *smokeClient) map[string]foodAPIIngredient {
	t.Helper()

	var response struct {
		Ingredients []foodAPIIngredient `json:"ingredients"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/v1/ingredients", nil, &response)

	byName := make(map[string]foodAPIIngredient, len(response.Ingredients))
	for _, ingredient := range response.Ingredients {
		byName[ingredient.Name] = ingredient
	}
	return byName
}

func TestFoodAPIStoresMeal(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")

	tomatoID := addIngredient(t, client, "Помидор", 20)

	ingredients := listIngredients(t, client)
	if tomato := ingredients["Помидор"]; len(ingredients) != 1 || tomato.ID != tomatoID || tomato.TimesUsed != 0 || tomato.LastUsed != nil {
		t.Fatalf("ingredients before the meal = %+v", ingredients)
	}

	// One meal: an existing ingredient, the same one with a calorie override,
	// a new ingredient and a one-off food that must stay out of the ingredient list
	meal := map[string]any{
		"meal_datetime": "2026-03-10T21:30:00+02:00",
		"items": []map[string]any{
			{"ingredient_id": tomatoID, "weight": 150},
			{"ingredient_id": tomatoID, "weight": 100, "kcal_per_100g": 25},
			{"new_ingredient": map[string]any{"name": "  Сыр   Гауда ", "kcal_per_100g": 345, "fats": 27, "proteins": 25}, "weight": 40},
			{"new_ingredient": map[string]any{"name": "Паста карбонара", "kcal_per_100g": 180}, "weight": 300, "one_off": true},
		},
	}
	var created foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food", meal, &created)

	if len(created.Entries) != 4 {
		t.Fatalf("created entries = %+v, want 4 entries", created.Entries)
	}
	wantEntries := []struct {
		food     string
		calories int
		kcal     float64
	}{
		{"Помидор", 30, 20},
		{"Помидор", 25, 25},
		{"Сыр Гауда", 138, 345},
		{"Паста карбонара", 540, 180},
	}
	for i, want := range wantEntries {
		got := created.Entries[i]
		if got.ID == 0 || got.Food != want.food || got.Calories != want.calories || got.KcalPer100g != want.kcal {
			t.Errorf("entry %d = %+v, want %+v", i, got, want)
		}
		if got.MealDatetime != "2026-03-10T19:30:00Z" {
			t.Errorf("entry %d meal_datetime = %q, want the time in UTC", i, got.MealDatetime)
		}
	}
	if created.TotalCalories != 733 || created.Day != "2026-03-10" ||
		created.DayTotalCalories == nil || *created.DayTotalCalories != 733 {
		t.Errorf("totals = %d kcal, day %s, day total %v", created.TotalCalories, created.Day, created.DayTotalCalories)
	}

	if id := created.Entries[0].IngredientID; id == nil || *id != tomatoID || created.Entries[0].IngredientCreated {
		t.Errorf("existing ingredient entry = %+v", created.Entries[0])
	}
	cheese, pasta := created.Entries[2], created.Entries[3]
	if cheese.IngredientID == nil || !cheese.IngredientCreated {
		t.Errorf("new ingredient entry = %+v, want a created ingredient", cheese)
	}
	if pasta.IngredientID != nil || pasta.IngredientCreated {
		t.Errorf("one-off entry = %+v, want no ingredient", pasta)
	}

	ingredients = listIngredients(t, client)
	if len(ingredients) != 2 {
		t.Fatalf("ingredients after the meal = %+v, want the tomato and the cheese", ingredients)
	}
	if tomato := ingredients["Помидор"]; tomato.TimesUsed != 2 || tomato.LastUsed == nil || *tomato.LastUsed != "2026-03-10" {
		t.Errorf("tomato usage = %+v", tomato)
	}
	if gouda := ingredients["Сыр Гауда"]; gouda.ID != *cheese.IngredientID || gouda.KcalPer100g != 345 || gouda.TimesUsed != 1 {
		t.Errorf("created ingredient = %+v", gouda)
	}

	// The meal is visible through the read API (with entry IDs) and in the web diary
	var data struct {
		Food []foodAPIEntry `json:"food"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/v1/data?type=food&from=2026-03-10&to=2026-03-10", nil, &data)
	if len(data.Food) != 4 || data.Food[0].ID == 0 {
		t.Errorf("food data = %+v, want 4 entries with IDs", data.Food)
	}
	var diary []foodAPIEntry
	client.expect(http.StatusOK, http.MethodGet, "/api/calories?dateFrom=2026-03-10&dateTo=2026-03-10", nil, &diary)
	if len(diary) != 4 {
		t.Errorf("web diary has %d entries, want 4", len(diary))
	}

	// Undo: delete an entry
	pastaPath := fmt.Sprintf("/api/v1/food/%d", pasta.ID)
	client.expect(http.StatusNoContent, http.MethodDelete, pastaPath, nil, nil)
	client.expect(http.StatusNotFound, http.MethodDelete, pastaPath, nil, nil)
	client.expect(http.StatusBadRequest, http.MethodDelete, "/api/v1/food/abc", nil, nil)

	// Without meal_datetime the meal is stored at the current time
	var now foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food",
		map[string]any{"items": []map[string]any{{"ingredient_id": tomatoID, "weight": 50}}}, &now)
	if today := time.Now().UTC().Format("2006-01-02"); now.Day != today {
		t.Errorf("meal without meal_datetime: day = %q, want %q", now.Day, today)
	}
}

func TestFoodAPIEditsEntry(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")

	breadID := addIngredient(t, client, "Хлеб", 240)
	baguetteID := addIngredient(t, client, "Багет", 270)

	var meal foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food", map[string]any{
		"meal_datetime": "2026-03-10T08:00:00Z",
		"items": []map[string]any{
			{"ingredient_id": breadID, "weight": 100},
			{"new_ingredient": map[string]any{"name": "Суп в кафе", "kcal_per_100g": 60}, "weight": 300, "one_off": true},
		},
	}, &meal)
	entryPath := fmt.Sprintf("/api/v1/food/%d", meal.Entries[0].ID)

	type updated struct {
		Entry            foodAPIEntry `json:"entry"`
		Day              string       `json:"day"`
		DayTotalCalories *int         `json:"day_total_calories"`
	}
	check := func(step string, got updated, food string, weight, kcal float64, calories int, datetime string, dayTotal int) {
		t.Helper()
		e := got.Entry
		if e.ID != meal.Entries[0].ID || e.Food != food || e.Weight != weight || e.KcalPer100g != kcal ||
			e.Calories != calories || e.MealDatetime != datetime {
			t.Errorf("%s: entry = %+v, want %s %g g, %g kcal/100g, %d kcal at %s", step, e, food, weight, kcal, calories, datetime)
		}
		if got.DayTotalCalories == nil || *got.DayTotalCalories != dayTotal {
			t.Errorf("%s: day total = %v, want %d", step, got.DayTotalCalories, dayTotal)
		}
	}

	// Only the weight changes: everything else, including the meal time, stays
	var result updated
	client.expect(http.StatusOK, http.MethodPut, entryPath, map[string]any{"weight": 150}, &result)
	check("weight", result, "Хлеб", 150, 240, 360, "2026-03-10T08:00:00Z", 540)

	// Another ingredient: its name and calories replace the old ones, the weight stays
	client.expect(http.StatusOK, http.MethodPut, entryPath, map[string]any{"ingredient_id": baguetteID}, &result)
	check("ingredient", result, "Багет", 150, 270, 405, "2026-03-10T08:00:00Z", 585)

	// A calorie override for this entry only
	client.expect(http.StatusOK, http.MethodPut, entryPath, map[string]any{"kcal_per_100g": 300}, &result)
	check("calories", result, "Багет", 150, 300, 450, "2026-03-10T08:00:00Z", 630)

	// Everything at once, including a move to another day
	result = updated{}
	client.expect(http.StatusOK, http.MethodPut, entryPath, map[string]any{
		"ingredient_id": breadID, "weight": 50, "kcal_per_100g": 250, "meal_datetime": "2026-03-11T01:30:00+02:00",
	}, &result)
	check("all fields", result, "Хлеб", 50, 250, 125, "2026-03-10T23:30:00Z", 305)

	// One-off entries have no ingredient but can be corrected the same way
	client.expect(http.StatusOK, http.MethodPut, fmt.Sprintf("/api/v1/food/%d", meal.Entries[1].ID),
		map[string]any{"weight": 400}, &result)
	if result.Entry.Food != "Суп в кафе" || result.Entry.Calories != 240 {
		t.Errorf("one-off entry = %+v, want the same food with 240 kcal", result.Entry)
	}

	for name, body := range map[string]map[string]any{
		"nothing to change":  {},
		"zero weight":        {"weight": 0},
		"weight too large":   {"weight": 20000},
		"calories too large": {"kcal_per_100g": 1200},
		"unknown ingredient": {"ingredient_id": 999999},
		"invalid datetime":   {"meal_datetime": "yesterday"},
	} {
		if rec := client.do(http.MethodPut, entryPath, body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want %d; body: %s", name, rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	}
	client.expect(http.StatusNotFound, http.MethodPut, "/api/v1/food/999999", map[string]any{"weight": 100}, nil)
	client.expect(http.StatusBadRequest, http.MethodPut, "/api/v1/food/abc", map[string]any{"weight": 100}, nil)

	// Rejected changes left the entry as it was, and the ingredient list was never touched
	var data struct {
		Food []foodAPIEntry `json:"food"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/v1/data?type=food&from=2026-03-10&to=2026-03-10", nil, &data)
	if len(data.Food) != 2 {
		t.Fatalf("diary = %+v, want 2 entries", data.Food)
	}
	for _, e := range data.Food {
		if e.ID == meal.Entries[0].ID && (e.Food != "Хлеб" || e.Weight != 50 || e.Calories != 125) {
			t.Errorf("entry after rejected changes = %+v", e)
		}
	}
	if ingredients := listIngredients(t, client); len(ingredients) != 2 || ingredients["Хлеб"].KcalPer100g != 240 {
		t.Errorf("ingredients = %+v, want the two original ones unchanged", ingredients)
	}
}

func TestFoodAPIRejectsInvalidMeals(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")

	salmonID := addIngredient(t, client, "Сёмга", 208)

	newFood := func(name string, kcalPer100g any) map[string]any {
		return map[string]any{"name": name, "kcal_per_100g": kcalPer100g}
	}
	valid := map[string]any{"ingredient_id": salmonID, "weight": 100}

	tests := []struct {
		name  string
		items []map[string]any
	}{
		{"no items", []map[string]any{}},
		{"neither ingredient_id nor new_ingredient", []map[string]any{{"weight": 100}}},
		{"both ingredient_id and new_ingredient", []map[string]any{{"ingredient_id": salmonID, "new_ingredient": newFood("Хлеб", 240), "weight": 100}}},
		{"unknown ingredient", []map[string]any{{"ingredient_id": 999999, "weight": 100}}},
		{"zero weight", []map[string]any{{"ingredient_id": salmonID, "weight": 0}}},
		{"weight in kilograms mistaken for grams", []map[string]any{{"ingredient_id": salmonID, "weight": 20000}}},
		{"calories of a portion instead of 100 g", []map[string]any{{"new_ingredient": newFood("Бургер", 1200), "weight": 100}}},
		{"new ingredient without calories", []map[string]any{{"new_ingredient": map[string]any{"name": "Хлеб"}, "weight": 100}}},
		{"new ingredient without name", []map[string]any{{"new_ingredient": newFood("  ", 240), "weight": 100}}},
		{"one_off with an existing ingredient", []map[string]any{{"ingredient_id": salmonID, "weight": 100, "one_off": true}}},
		{"same new ingredient twice", []map[string]any{
			{"new_ingredient": newFood("Хлеб", 240), "weight": 100},
			{"new_ingredient": newFood("хлеб", 240), "weight": 50},
		}},
		// The valid items of a rejected meal must not be stored either
		{"valid items followed by an invalid one", []map[string]any{
			valid,
			{"new_ingredient": newFood("Хлеб", 240), "weight": 100},
			{"ingredient_id": salmonID, "weight": -5},
		}},
	}
	for _, tt := range tests {
		rec := client.do(http.MethodPost, "/api/v1/food", map[string]any{"meal_datetime": "2026-03-10T12:00:00Z", "items": tt.items})
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want %d; body: %s", tt.name, rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	}

	client.expect(http.StatusBadRequest, http.MethodPost, "/api/v1/food",
		map[string]any{"meal_datetime": "yesterday", "items": []map[string]any{valid}}, nil)

	// A new ingredient whose name differs from an existing one only by case, spacing or "ё" is a conflict:
	// the response points to the ingredient to use instead
	var conflict struct {
		ItemIndex          int               `json:"item_index"`
		ExistingIngredient foodAPIIngredient `json:"existing_ingredient"`
	}
	client.expect(http.StatusConflict, http.MethodPost, "/api/v1/food", map[string]any{
		"meal_datetime": "2026-03-10T12:00:00Z",
		"items":         []map[string]any{valid, {"new_ingredient": newFood(" семга ", 200), "weight": 100}},
	}, &conflict)
	if conflict.ItemIndex != 1 || conflict.ExistingIngredient.ID != salmonID || conflict.ExistingIngredient.Name != "Сёмга" {
		t.Errorf("conflict = %+v, want item 1 pointing to the existing ingredient", conflict)
	}

	// Nothing was stored by the rejected meals
	if ingredients := listIngredients(t, client); len(ingredients) != 1 {
		t.Errorf("ingredients after rejected meals = %+v, want only the original one", ingredients)
	}
	var diary []foodAPIEntry
	client.expect(http.StatusOK, http.MethodGet, "/api/calories?dateFrom=2026-03-10&dateTo=2026-03-10", nil, &diary)
	if len(diary) != 0 {
		t.Errorf("diary after rejected meals = %+v, want no entries", diary)
	}
}

func TestFoodAPIIsolatesUsers(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)

	signUp(t, client, smtpServer, "first@example.com")
	firstIngredientID := addIngredient(t, client, "Tomato", 20)
	var meal foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food",
		map[string]any{"items": []map[string]any{{"ingredient_id": firstIngredientID, "weight": 100}}}, &meal)
	firstToken, firstAPIKey := client.token, client.apiKey

	signUp(t, client, smtpServer, "second@example.com")

	// The second user can neither use the ingredients nor delete the entries of the first one
	client.expect(http.StatusBadRequest, http.MethodPost, "/api/v1/food",
		map[string]any{"items": []map[string]any{{"ingredient_id": firstIngredientID, "weight": 100}}}, nil)
	client.expect(http.StatusNotFound, http.MethodDelete, fmt.Sprintf("/api/v1/food/%d", meal.Entries[0].ID), nil, nil)
	client.expect(http.StatusNotFound, http.MethodPut, fmt.Sprintf("/api/v1/food/%d", meal.Entries[0].ID),
		map[string]any{"weight": 1}, nil)
	if ingredients := listIngredients(t, client); len(ingredients) != 0 {
		t.Errorf("second user sees ingredients of the first one: %+v", ingredients)
	}

	// The same ingredient name is not a conflict across users, and the diary of the first user
	// does not count as usage of the second user's ingredient
	addIngredient(t, client, "Tomato", 18)
	if tomato := listIngredients(t, client)["Tomato"]; tomato.ID == firstIngredientID || tomato.TimesUsed != 0 || tomato.LastUsed != nil {
		t.Errorf("second user's ingredient = %+v, want an own unused ingredient", tomato)
	}
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food", map[string]any{
		"items": []map[string]any{{"new_ingredient": map[string]any{"name": "Cucumber", "kcal_per_100g": 10}, "weight": 100}},
	}, nil)

	// The first user still has exactly their own entry and ingredient
	secondToken, secondAPIKey := client.token, client.apiKey
	client.token, client.apiKey = firstToken, firstAPIKey
	today := time.Now().UTC().Format("2006-01-02")
	var data struct {
		Food []foodAPIEntry `json:"food"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/v1/data?type=food&from="+today+"&to="+today, nil, &data)
	if len(data.Food) != 1 || data.Food[0].ID != meal.Entries[0].ID || data.Food[0].Weight != 100 {
		t.Errorf("first user's diary = %+v, want only the unchanged entry they created", data.Food)
	}
	if ingredients := listIngredients(t, client); len(ingredients) != 1 || ingredients["Tomato"].TimesUsed != 1 {
		t.Errorf("first user's ingredients = %+v, want only their own tomato used once", ingredients)
	}
	client.token, client.apiKey = secondToken, secondAPIKey

	client.apiKey = ""
	client.expect(http.StatusUnauthorized, http.MethodPost, "/api/v1/food",
		map[string]any{"items": []map[string]any{{"ingredient_id": firstIngredientID, "weight": 100}}}, nil)
	client.expect(http.StatusUnauthorized, http.MethodGet, "/api/v1/ingredients", nil, nil)
}
