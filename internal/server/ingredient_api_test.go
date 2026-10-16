package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// invisibleSpace is a zero-width space: it must not make two ingredient names different
const invisibleSpace = string(rune(0x200b))

const duplicateNameMessage = "an ingredient with this name already exists"

// apiIngredient is an ingredient in the shape of the external API (/api/v1)
type apiIngredient struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	KcalPer100g float64  `json:"kcal_per_100g"`
	Fats        *float64 `json:"fats"`
	Carbs       *float64 `json:"carbs"`
	Proteins    *float64 `json:"proteins"`
}

// webIngredient is an ingredient in the shape of the website API (/api/ingredients)
type webIngredient struct {
	ID          int      `json:"id"`
	UserID      int      `json:"user_id"`
	Name        string   `json:"name"`
	KcalPer100g float64  `json:"kcalPer100g"`
	Fats        *float64 `json:"fats"`
	Carbs       *float64 `json:"carbs"`
	Proteins    *float64 `json:"proteins"`
}

func optional(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func (i apiIngredient) String() string {
	return fmt.Sprintf("#%d %q %g kcal, fats %v, carbs %v, proteins %v",
		i.ID, i.Name, i.KcalPer100g, optional(i.Fats), optional(i.Carbs), optional(i.Proteins))
}

func (i apiIngredient) is(name string, kcal float64, fats, carbs, proteins any) bool {
	return i.Name == name && i.KcalPer100g == kcal &&
		optional(i.Fats) == fats && optional(i.Carbs) == carbs && optional(i.Proteins) == proteins
}

// createWebIngredient creates an ingredient through the website API with the given fields
func createWebIngredient(t *testing.T, client *smokeClient, body map[string]any) webIngredient {
	t.Helper()
	var ingredient webIngredient
	client.expect(http.StatusCreated, http.MethodPost, "/api/ingredients", body, &ingredient)
	return ingredient
}

// addBread creates the ingredient used by most cases: Хлеб, 250 kcal, fats 1, carbs 50, proteins 8
func addBread(t *testing.T, client *smokeClient) int {
	t.Helper()
	return createWebIngredient(t, client,
		map[string]any{"name": "Хлеб", "kcalPer100g": 250, "fats": 1, "carbs": 50, "proteins": 8}).ID
}

func ingredientPath(id any) string {
	return fmt.Sprintf("/api/v1/ingredients/%v", id)
}

func editIngredient(t *testing.T, client *smokeClient, id int, body any) apiIngredient {
	t.Helper()
	var response struct {
		Ingredient apiIngredient `json:"ingredient"`
	}
	client.expect(http.StatusOK, http.MethodPut, ingredientPath(id), body, &response)
	return response.Ingredient
}

// v1Ingredients returns the ingredient list of the external API by ID
func v1Ingredients(t *testing.T, client *smokeClient) map[int]apiIngredient {
	t.Helper()
	var response struct {
		Ingredients []apiIngredient `json:"ingredients"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/v1/ingredients", nil, &response)
	byID := make(map[int]apiIngredient, len(response.Ingredients))
	for _, ingredient := range response.Ingredients {
		byID[ingredient.ID] = ingredient
	}
	return byID
}

func webIngredients(t *testing.T, client *smokeClient) map[int]webIngredient {
	t.Helper()
	var ingredients []webIngredient
	client.expect(http.StatusOK, http.MethodGet, "/api/ingredients", nil, &ingredients)
	byID := make(map[int]webIngredient, len(ingredients))
	for _, ingredient := range ingredients {
		byID[ingredient.ID] = ingredient
	}
	return byID
}

// doRaw sends a request with a body as is, e.g. malformed JSON
func doRaw(client *smokeClient, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "203.0.113.7:1234"
	req.Header.Set("Content-Type", "application/json")
	if client.token != "" {
		req.Header.Set("Authorization", "Bearer "+client.token)
	}
	if client.apiKey != "" {
		req.Header.Set("X-API-Key", client.apiKey)
	}
	rec := httptest.NewRecorder()
	client.handler.ServeHTTP(rec, req)
	return rec
}

// expectError checks the status of an error response, that its message contains every part,
// and that it does not reveal internal error text
func expectError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantParts ...string) {
	t.Helper()
	body := rec.Body.String()
	if rec.Code != wantStatus {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, wantStatus, body)
		return
	}
	var response struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("error response is not JSON: %s", body)
		return
	}
	for _, part := range wantParts {
		if !strings.Contains(response.Message, part) {
			t.Errorf("message %q does not contain %q", response.Message, part)
		}
	}
	lower := strings.ToLower(body)
	for _, internal := range []string{"sql", "unique", "constraint", "no rows", "strconv", "json:", "unmarshal"} {
		if strings.Contains(lower, internal) {
			t.Errorf("error response reveals internal error text (%q): %s", internal, body)
		}
	}
}

// conflict decodes a 409 answer: the message and the raw existing ingredient
func conflict(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("conflict response is not JSON: %s", rec.Body.String())
	}
	if response["message"] != duplicateNameMessage {
		t.Errorf("message = %v, want %q", response["message"], duplicateNameMessage)
	}
	if _, ok := response["item_index"]; ok {
		t.Errorf("conflict outside POST /food must not have item_index: %v", response)
	}
	existing, ok := response["existing_ingredient"].(map[string]any)
	if !ok {
		t.Fatalf("conflict without existing_ingredient: %s", rec.Body.String())
	}
	return existing
}

// expectConflict checks a 409 answer that points to the ingredient with the given ID and name
func expectConflict(t *testing.T, rec *httptest.ResponseRecorder, id int, name, kcalKey string) {
	t.Helper()
	existing := conflict(t, rec)
	if existing["id"] != float64(id) || existing["name"] != name {
		t.Errorf("existing_ingredient = %v, want #%d %q", existing, id, name)
	}
	if _, ok := existing[kcalKey]; !ok {
		t.Errorf("existing_ingredient = %v, want the %s field of this API", existing, kcalKey)
	}
}

func TestIngredientAPIEditsIngredient(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	breadID := addBread(t, client)

	// Only the calories change; the response has the external API shape
	rec := client.expect(http.StatusOK, http.MethodPut, ingredientPath(breadID), map[string]any{"kcal_per_100g": 260}, nil)
	want := fmt.Sprintf(`{"ingredient":{"id":%d,"name":"Хлеб","kcal_per_100g":260,"fats":1,"carbs":50,"proteins":8}}`, breadID)
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Errorf("response = %s, want %s", got, want)
	}
	if bread := v1Ingredients(t, client)[breadID]; !bread.is("Хлеб", 260, 1.0, 50.0, 8.0) {
		t.Errorf("external API list: %v", bread)
	}
	if bread := webIngredients(t, client)[breadID]; bread.KcalPer100g != 260 || bread.Name != "Хлеб" {
		t.Errorf("website list: %+v", bread)
	}

	// The name is cleaned, null is the same as not sent
	got := editIngredient(t, client, breadID, map[string]any{"name": "  Хлеб" + invisibleSpace + "   ржаной ", "fats": nil})
	if !got.is("Хлеб ржаной", 260, 1.0, 50.0, 8.0) {
		t.Errorf("after rename: %v", got)
	}

	got = editIngredient(t, client, breadID, map[string]any{
		"name": "Батон", "kcal_per_100g": 270, "fats": 1.2, "carbs": 48, "proteins": 8.5,
	})
	if !got.is("Батон", 270, 1.2, 48.0, 8.5) {
		t.Errorf("after changing all fields: %v", got)
	}

	// Boundary values are allowed
	got = editIngredient(t, client, breadID, map[string]any{"kcal_per_100g": 0, "fats": 0, "carbs": 0, "proteins": 0})
	if !got.is("Батон", 0, 0.0, 0.0, 0.0) {
		t.Errorf("lower boundaries: %v", got)
	}
	got = editIngredient(t, client, breadID, map[string]any{"kcal_per_100g": 900, "fats": 100, "carbs": 100, "proteins": 100})
	if !got.is("Батон", 900, 100.0, 100.0, 100.0) {
		t.Errorf("upper boundaries: %v", got)
	}
	longName := strings.Repeat("я", 200)
	if got = editIngredient(t, client, breadID, map[string]any{"name": longName}); got.Name != longName {
		t.Errorf("200-character name: %v", got)
	}

	// A value the website allows but the API does not is kept when it is not sent
	butterID := createWebIngredient(t, client, map[string]any{"name": "Масло", "kcalPer100g": 1000}).ID
	got = editIngredient(t, client, butterID, map[string]any{"name": "Масло сливочное"})
	if !got.is("Масло сливочное", 1000, nil, nil, nil) {
		t.Errorf("rename of an ingredient with 1000 kcal: %v", got)
	}

	// The route accepts only API keys: a website session is not enough
	client.apiKey = ""
	client.expect(http.StatusUnauthorized, http.MethodPut, ingredientPath(breadID), map[string]any{"kcal_per_100g": 1}, nil)
	client.expect(http.StatusUnauthorized, http.MethodDelete, ingredientPath(breadID), nil, nil)
	if bread := webIngredients(t, client)[breadID]; bread.KcalPer100g != 900 {
		t.Errorf("ingredient changed by a request without an API key: %+v", bread)
	}
}

func TestIngredientAPIRejectsInvalidValues(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	breadID := addBread(t, client)

	tests := []struct {
		body      map[string]any
		wantParts []string
	}{
		{map[string]any{"kcal_per_100g": -1}, []string{"kcal_per_100g", "900"}},
		{map[string]any{"kcal_per_100g": 900.1}, []string{"kcal_per_100g", "900"}},
		{map[string]any{"fats": -0.1}, []string{"fats", "100"}},
		{map[string]any{"fats": 100.1}, []string{"fats", "100"}},
		{map[string]any{"carbs": -0.1}, []string{"carbs", "100"}},
		{map[string]any{"carbs": 100.1}, []string{"carbs", "100"}},
		{map[string]any{"proteins": -0.1}, []string{"proteins", "100"}},
		{map[string]any{"proteins": 100.1}, []string{"proteins", "100"}},
		{map[string]any{"name": "Батон", "kcal_per_100g": 1000}, []string{"kcal_per_100g"}},
	}
	for _, tt := range tests {
		expectError(t, client.do(http.MethodPut, ingredientPath(breadID), tt.body), http.StatusBadRequest, tt.wantParts...)
	}

	if bread := v1Ingredients(t, client)[breadID]; !bread.is("Хлеб", 250, 1.0, 50.0, 8.0) {
		t.Errorf("ingredient after rejected changes: %v", bread)
	}
}

func TestIngredientAPIRejectsInvalidRequests(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	breadID := addBread(t, client)
	path := ingredientPath(breadID)

	for _, name := range []string{"", "   ", invisibleSpace + " \t"} {
		expectError(t, client.do(http.MethodPut, path, map[string]any{"name": name}), http.StatusBadRequest, "name")
	}
	expectError(t, client.do(http.MethodPut, path, map[string]any{"name": strings.Repeat("я", 201)}),
		http.StatusBadRequest, "name", "200")

	for _, body := range []map[string]any{{}, {"name": nil, "kcal_per_100g": nil, "fats": nil}} {
		expectError(t, client.do(http.MethodPut, path, body), http.StatusBadRequest,
			"nothing to change", "name", "kcal_per_100g", "fats", "carbs", "proteins")
	}

	expectError(t, doRaw(client, http.MethodPut, path, `{"kcal_per_100g": 260`), http.StatusBadRequest, "Invalid request body")
	expectError(t, client.do(http.MethodPut, path, map[string]any{"kcal_per_100g": "abc"}), http.StatusBadRequest, "Invalid request body")
	expectError(t, client.do(http.MethodPut, ingredientPath("abc"), map[string]any{"kcal_per_100g": 260}),
		http.StatusBadRequest, "Invalid ingredient ID")
	expectError(t, client.do(http.MethodPut, ingredientPath(999999), map[string]any{"kcal_per_100g": 260}),
		http.StatusNotFound, "Ingredient not found")

	if bread := v1Ingredients(t, client)[breadID]; !bread.is("Хлеб", 250, 1.0, 50.0, 8.0) {
		t.Errorf("ingredient after rejected requests: %v", bread)
	}
}

func TestIngredientAPIRejectsDuplicateNames(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)

	// Another user's ingredient names do not conflict
	signUp(t, client, smtpServer, "second@example.com")
	createWebIngredient(t, client, map[string]any{"name": "Овсянка", "kcalPer100g": 370})

	signUp(t, client, smtpServer, "first@example.com")
	buckwheatID := createWebIngredient(t, client, map[string]any{"name": "Гречка", "kcalPer100g": 330, "proteins": 12.6}).ID
	hedgehogID := createWebIngredient(t, client, map[string]any{"name": "Ёжик", "kcalPer100g": 100}).ID
	breadID := addBread(t, client)

	for _, name := range []string{"гречка", " Гречка ", "ГРЕЧКА"} {
		rec := client.do(http.MethodPut, ingredientPath(breadID), map[string]any{"name": name})
		expectConflict(t, rec, buckwheatID, "Гречка", "kcal_per_100g")
		if existing := conflict(t, rec); existing["kcal_per_100g"] != 330.0 || existing["proteins"] != 12.6 {
			t.Errorf("existing_ingredient = %v, want its values", existing)
		}
	}
	expectConflict(t, client.do(http.MethodPut, ingredientPath(breadID), map[string]any{"name": "ежик", "kcal_per_100g": 1}),
		hedgehogID, "Ёжик", "kcal_per_100g")

	// Variants of the own name are not a conflict
	for _, name := range []string{"хлеб", "ХЛЕБ"} {
		if got := editIngredient(t, client, breadID, map[string]any{"name": name}); got.Name != name {
			t.Errorf("rename to %q: %v", name, got)
		}
	}
	if bread := v1Ingredients(t, client)[breadID]; !bread.is("ХЛЕБ", 250, 1.0, 50.0, 8.0) {
		t.Errorf("ingredient after conflicts: %v", bread)
	}

	if got := editIngredient(t, client, breadID, map[string]any{"name": "Овсянка"}); got.Name != "Овсянка" {
		t.Errorf("rename to another user's name: %v", got)
	}
}

func TestIngredientAPIDeletesIngredient(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "first@example.com")
	breadID := addBread(t, client)
	firstToken, firstAPIKey := client.token, client.apiKey

	// Another user can neither change nor delete it; both look like a missing ingredient
	signUp(t, client, smtpServer, "second@example.com")
	expectError(t, client.do(http.MethodPut, ingredientPath(breadID), map[string]any{"kcal_per_100g": 1, "name": "Взлом"}),
		http.StatusNotFound, "Ingredient not found")
	expectError(t, client.do(http.MethodDelete, ingredientPath(breadID), nil), http.StatusNotFound, "Ingredient not found")

	// Invalid and revoked keys
	var revoked struct {
		ID  int    `json:"id"`
		Key string `json:"key"`
	}
	client.expect(http.StatusCreated, http.MethodPost, "/api/api-keys", map[string]any{"name": "old"}, &revoked)
	client.expect(http.StatusOK, http.MethodPost, fmt.Sprintf("/api/api-keys/%d/revoke", revoked.ID), nil, nil)
	client.token = ""
	for _, key := range []string{"invalid-key", revoked.Key} {
		client.apiKey = key
		client.expect(http.StatusUnauthorized, http.MethodPut, ingredientPath(breadID), map[string]any{"kcal_per_100g": 1}, nil)
		client.expect(http.StatusUnauthorized, http.MethodDelete, ingredientPath(breadID), nil, nil)
	}

	client.token, client.apiKey = firstToken, firstAPIKey
	if bread := v1Ingredients(t, client)[breadID]; !bread.is("Хлеб", 250, 1.0, 50.0, 8.0) {
		t.Fatalf("ingredient after the other user's requests: %v", bread)
	}

	rec := client.expect(http.StatusNoContent, http.MethodDelete, ingredientPath(breadID), nil, nil)
	if rec.Body.Len() != 0 {
		t.Errorf("delete response body = %q, want empty", rec.Body.String())
	}
	expectError(t, client.do(http.MethodDelete, ingredientPath(breadID), nil), http.StatusNotFound, "Ingredient not found")
	expectError(t, client.do(http.MethodDelete, ingredientPath("abc"), nil), http.StatusBadRequest, "Invalid ingredient ID")

	if _, ok := v1Ingredients(t, client)[breadID]; ok {
		t.Error("deleted ingredient is still in the external API list")
	}
	if _, ok := webIngredients(t, client)[breadID]; ok {
		t.Error("deleted ingredient is still in the website list")
	}
}

func TestIngredientAPIKeepsDiaryEntries(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	breadID := addBread(t, client)

	var meal foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food", map[string]any{
		"meal_datetime": "2026-03-10T08:00:00Z",
		"items":         []map[string]any{{"ingredient_id": breadID, "weight": 100}},
	}, &meal)
	entryID := meal.Entries[0].ID

	diaryEntry := func() (foodAPIEntry, int) {
		t.Helper()
		var data struct {
			Food []foodAPIEntry `json:"food"`
		}
		client.expect(http.StatusOK, http.MethodGet, "/api/v1/data?type=food&from=2026-03-10&to=2026-03-10", nil, &data)
		if len(data.Food) != 1 || data.Food[0].ID != entryID {
			t.Fatalf("diary = %+v, want the one entry", data.Food)
		}
		return data.Food[0], data.Food[0].Calories
	}

	// Editing the ingredient does not change the logged entry
	editIngredient(t, client, breadID, map[string]any{"kcal_per_100g": 300, "name": "Хлеб ржаной"})
	if entry, _ := diaryEntry(); entry.Food != "Хлеб" || entry.KcalPer100g != 250 || entry.Calories != 250 {
		t.Errorf("entry after the ingredient edit = %+v, want the old copy", entry)
	}

	// Switching the entry to the same ingredient copies the corrected values
	var updated struct {
		Entry            foodAPIEntry `json:"entry"`
		DayTotalCalories *int         `json:"day_total_calories"`
	}
	client.expect(http.StatusOK, http.MethodPut, fmt.Sprintf("/api/v1/food/%d", entryID), map[string]any{"ingredient_id": breadID}, &updated)
	if e := updated.Entry; e.Food != "Хлеб ржаной" || e.KcalPer100g != 300 || e.Calories != 300 || e.Weight != 100 ||
		e.MealDatetime != "2026-03-10T08:00:00Z" {
		t.Errorf("entry after re-sync = %+v", e)
	}

	// Deleting the ingredient keeps the entry, it still counts and can still be edited
	client.expect(http.StatusNoContent, http.MethodDelete, ingredientPath(breadID), nil, nil)
	if entry, calories := diaryEntry(); entry.Food != "Хлеб ржаной" || calories != 300 {
		t.Errorf("entry after the ingredient was deleted = %+v", entry)
	}
	client.expect(http.StatusOK, http.MethodPut, fmt.Sprintf("/api/v1/food/%d", entryID), map[string]any{"weight": 50}, &updated)
	if updated.Entry.Calories != 150 || updated.DayTotalCalories == nil || *updated.DayTotalCalories != 150 {
		t.Errorf("entry after a weight change = %+v, day total %v", updated.Entry, updated.DayTotalCalories)
	}

	// The deleted ingredient cannot be logged any more
	expectError(t, client.do(http.MethodPost, "/api/v1/food", map[string]any{
		"meal_datetime": "2026-03-10T12:00:00Z",
		"items":         []map[string]any{{"ingredient_id": breadID, "weight": 100}},
	}), http.StatusBadRequest, fmt.Sprintf("items[0]: ingredient %d not found", breadID))
	if _, calories := diaryEntry(); calories != 150 {
		t.Errorf("rejected meal changed the diary: %d kcal", calories)
	}
}

func TestIngredientsWebRejectsDuplicateNames(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)

	signUp(t, client, smtpServer, "second@example.com")
	createWebIngredient(t, client, map[string]any{"name": "Овсянка", "kcalPer100g": 370})
	createWebIngredient(t, client, map[string]any{"name": "Перловка", "kcalPer100g": 320})

	signUp(t, client, smtpServer, "first@example.com")
	buckwheat := createWebIngredient(t, client, map[string]any{"name": "Гречка", "kcalPer100g": 330, "proteins": 12.6})
	hedgehogID := createWebIngredient(t, client, map[string]any{"name": "Ёжик", "kcalPer100g": 100}).ID
	breadID := addBread(t, client)
	before := len(webIngredients(t, client))

	// Exact and near-duplicates on create; the existing ingredient comes in the website shape
	for _, name := range []string{"Гречка", "гречка", " Гречка ", "ГРЕЧКА", "Греч" + invisibleSpace + "ка"} {
		rec := client.do(http.MethodPost, "/api/ingredients", map[string]any{"name": name, "kcalPer100g": 330})
		expectConflict(t, rec, buckwheat.ID, "Гречка", "kcalPer100g")
		if existing := conflict(t, rec); existing["user_id"] != float64(buckwheat.UserID) || existing["proteins"] != 12.6 {
			t.Errorf("existing_ingredient = %v, want the stored ingredient", existing)
		}
	}
	expectConflict(t, client.do(http.MethodPost, "/api/ingredients", map[string]any{"name": "ежик", "kcalPer100g": 100}),
		hedgehogID, "Ёжик", "kcalPer100g")
	if after := len(webIngredients(t, client)); after != before {
		t.Errorf("ingredients = %d, want %d: a duplicate was created", after, before)
	}

	// Renames
	breadPath := fmt.Sprintf("/api/ingredients/%d", breadID)
	for _, name := range []string{"Гречка", "гречка"} {
		expectConflict(t, client.do(http.MethodPut, breadPath, map[string]any{"name": name, "kcalPer100g": 250}),
			buckwheat.ID, "Гречка", "kcalPer100g")
	}
	var bread webIngredient
	client.expect(http.StatusOK, http.MethodGet, breadPath, nil, &bread)
	if bread.Name != "Хлеб" || bread.KcalPer100g != 250 || optional(bread.Fats) != 1.0 {
		t.Errorf("ingredient after rejected renames = %+v", bread)
	}
	client.expect(http.StatusOK, http.MethodPut, breadPath, map[string]any{"name": "ХЛЕБ", "kcalPer100g": 260}, &bread)
	if bread.Name != "ХЛЕБ" || bread.KcalPer100g != 260 {
		t.Errorf("own-name variant = %+v", bread)
	}

	// Near-duplicates from before the rule: the diary still creates ingredients by exact name
	createWebIngredient(t, client, map[string]any{"name": "Рис", "kcalPer100g": 130})
	client.expect(http.StatusCreated, http.MethodPost, "/api/calories", map[string]any{
		"food": "рис", "calories": 130, "weight": 100, "kcalPer100g": 130, "meal_datetime": "2026-03-10T12:00:00Z",
	}, nil)
	var lowerRice webIngredient
	for _, ingredient := range webIngredients(t, client) {
		if ingredient.Name == "рис" {
			lowerRice = ingredient
		}
	}
	if lowerRice.ID == 0 {
		t.Fatal("the diary did not create the near-duplicate ingredient")
	}
	client.expect(http.StatusOK, http.MethodPut, fmt.Sprintf("/api/ingredients/%d", lowerRice.ID),
		map[string]any{"name": "рис", "kcalPer100g": 120}, nil)
	existing := conflict(t, client.do(http.MethodPut, breadPath, map[string]any{"name": "РИС", "kcalPer100g": 250}))
	if name := existing["name"]; name != "Рис" && name != "рис" {
		t.Errorf("existing_ingredient = %v, want one of the rice ingredients", existing)
	}

	// Names of another user do not conflict
	createWebIngredient(t, client, map[string]any{"name": "Овсянка", "kcalPer100g": 370})
	client.expect(http.StatusOK, http.MethodPut, breadPath, map[string]any{"name": "Перловка", "kcalPer100g": 320}, nil)
}

func TestIngredientsWebKeepsOtherBehavior(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "first@example.com")

	// The name is stored as sent and there are no upper limits on the website
	created := createWebIngredient(t, client, map[string]any{"name": " Масло ", "kcalPer100g": 1000, "fats": 120})
	if created.Name != " Масло " || created.KcalPer100g != 1000 || optional(created.Fats) != 120.0 {
		t.Errorf("created = %+v", created)
	}
	path := fmt.Sprintf("/api/ingredients/%d", created.ID)

	// Full replace: nutrients that are not sent are cleared
	var updated webIngredient
	client.expect(http.StatusOK, http.MethodPut, path, map[string]any{"name": "Масло", "kcalPer100g": 750}, &updated)
	if updated.Name != "Масло" || updated.KcalPer100g != 750 || updated.Fats != nil {
		t.Errorf("updated = %+v, want the nutrients cleared", updated)
	}

	expectError(t, client.do(http.MethodPost, "/api/ingredients", map[string]any{"name": "Хлеб", "kcalPer100g": -1}), http.StatusBadRequest)
	expectError(t, client.do(http.MethodPut, path, map[string]any{"name": "Масло", "kcalPer100g": -1}), http.StatusBadRequest)
	expectError(t, client.do(http.MethodPut, "/api/ingredients/999999", map[string]any{"name": "Масло", "kcalPer100g": 1}),
		http.StatusNotFound, "Ingredient not found")

	firstToken := client.token
	signUp(t, client, smtpServer, "second@example.com")
	expectError(t, client.do(http.MethodPut, path, map[string]any{"name": "Взлом", "kcalPer100g": 1}),
		http.StatusNotFound, "Ingredient not found")
	expectError(t, client.do(http.MethodDelete, path, nil), http.StatusNotFound, "Ingredient not found")

	client.token = firstToken
	client.expect(http.StatusNoContent, http.MethodDelete, path, nil, nil)
	client.expect(http.StatusNotFound, http.MethodDelete, path, nil, nil)
}

// Every test below stays under the burst of the api-v1 rate limiter: all requests of a test come from one address

const v1IngredientsPath = "/api/v1/ingredients"

// createV1Ingredient creates an ingredient through the external API and returns it with the raw response body
func createV1Ingredient(t *testing.T, client *smokeClient, body any) (apiIngredient, string) {
	t.Helper()
	var created apiIngredient
	rec := client.expect(http.StatusCreated, http.MethodPost, v1IngredientsPath, body, &created)
	return created, strings.TrimSpace(rec.Body.String())
}

// v1IngredientRows returns the raw rows of the external API ingredient list by ID
func v1IngredientRows(t *testing.T, client *smokeClient) map[int]string {
	t.Helper()
	var response struct {
		Ingredients []json.RawMessage `json:"ingredients"`
	}
	client.expect(http.StatusOK, http.MethodGet, v1IngredientsPath, nil, &response)
	rows := make(map[int]string, len(response.Ingredients))
	for _, raw := range response.Ingredients {
		var row struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw, &row); err != nil {
			t.Fatalf("invalid ingredient row %s: %v", raw, err)
		}
		rows[row.ID] = string(raw)
	}
	return rows
}

func TestIngredientAPICreatesIngredient(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")

	// The response is the row of the list: the same keys in the same order, without last_used for an unused name
	created, body := createV1Ingredient(t, client,
		map[string]any{"name": "Хлебцы ржаные", "kcal_per_100g": 310, "fats": 2, "carbs": 60, "proteins": 10})
	if want := fmt.Sprintf(`{"id":%d,"name":"Хлебцы ржаные","kcal_per_100g":310,"fats":2,"carbs":60,"proteins":10,"times_used":0}`,
		created.ID); body != want {
		t.Errorf("response = %s, want %s", body, want)
	}

	// Macros that are not sent, or sent as null, stay unknown and are absent
	sauce, sauceBody := createV1Ingredient(t, client, map[string]any{"name": "Соус", "kcal_per_100g": 120})
	if want := fmt.Sprintf(`{"id":%d,"name":"Соус","kcal_per_100g":120,"times_used":0}`, sauce.ID); sauceBody != want {
		t.Errorf("response without macros = %s, want %s", sauceBody, want)
	}
	nulls, nullsBody := createV1Ingredient(t, client,
		map[string]any{"name": "Кетчуп", "kcal_per_100g": 100, "fats": nil, "carbs": nil, "proteins": nil})
	if want := fmt.Sprintf(`{"id":%d,"name":"Кетчуп","kcal_per_100g":100,"times_used":0}`, nulls.ID); nullsBody != want {
		t.Errorf("response with null macros = %s, want %s", nullsBody, want)
	}

	// The name is cleaned before it is stored
	if cleaned, _ := createV1Ingredient(t, client, map[string]any{"name": "  Хлебцы" + invisibleSpace + "   с отрубями ", "kcal_per_100g": 300}); cleaned.Name != "Хлебцы с отрубями" {
		t.Errorf("name = %q, want the cleaned name", cleaned.Name)
	}

	// Boundary values; unknown fields are ignored, the ingredient belongs to the key's user
	lower, _ := createV1Ingredient(t, client, map[string]any{"name": "Вода", "kcal_per_100g": 0, "fats": 0, "carbs": 0, "proteins": 0})
	upper, _ := createV1Ingredient(t, client, map[string]any{"name": "Жир", "kcal_per_100g": 900, "fats": 100, "carbs": 100, "proteins": 100})
	longName := strings.Repeat("я", 200)
	long, _ := createV1Ingredient(t, client, map[string]any{"name": longName, "kcal_per_100g": 1, "user_id": 999999, "times_used": 5})
	if !lower.is("Вода", 0, 0.0, 0.0, 0.0) || !upper.is("Жир", 900, 100.0, 100.0, 100.0) || !long.is(longName, 1, nil, nil, nil) {
		t.Errorf("boundaries: %v, %v, %v", lower, upper, long)
	}

	// The list rows equal the responses
	rows := v1IngredientRows(t, client)
	for id, want := range map[int]string{created.ID: body, sauce.ID: sauceBody, nulls.ID: nullsBody} {
		if rows[id] != want {
			t.Errorf("list row = %s, want the response %s", rows[id], want)
		}
	}
	if !strings.Contains(rows[long.ID], `"times_used":0`) {
		t.Errorf("row with ignored fields = %s", rows[long.ID])
	}

	// The website lists it with the same values
	web := webIngredients(t, client)[created.ID]
	if web.Name != "Хлебцы ржаные" || web.KcalPer100g != 310 ||
		optional(web.Fats) != 2.0 || optional(web.Carbs) != 60.0 || optional(web.Proteins) != 10.0 {
		t.Errorf("website list: %+v", web)
	}
	if web := webIngredients(t, client)[long.ID]; web.Name != longName {
		t.Errorf("ingredient with a user_id in the body is not in the key user's list: %+v", web)
	}

	// The new ingredient can be logged right away
	var meal foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food",
		map[string]any{"items": []map[string]any{{"ingredient_id": created.ID, "weight": 100}}}, &meal)
	if e := meal.Entries[0]; e.Calories != 310 || e.Food != "Хлебцы ржаные" || e.IngredientCreated {
		t.Errorf("entry = %+v, want 310 kcal of the created ingredient", e)
	}
}

func TestIngredientAPICreateRejectsInvalidValues(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	before := v1IngredientRows(t, client)

	tests := []struct {
		field string
		value any
		want  string
	}{
		{"kcal_per_100g", -1, "kcal_per_100g must be between 0 and 900"},
		{"kcal_per_100g", 900.1, "kcal_per_100g must be between 0 and 900"},
		{"fats", -0.1, "fats must be between 0 and 100 (grams per 100 g)"},
		{"fats", 100.1, "fats must be between 0 and 100 (grams per 100 g)"},
		{"carbs", -0.1, "carbs must be between 0 and 100 (grams per 100 g)"},
		{"carbs", 100.1, "carbs must be between 0 and 100 (grams per 100 g)"},
		{"proteins", -0.1, "proteins must be between 0 and 100 (grams per 100 g)"},
		{"proteins", 100.1, "proteins must be between 0 and 100 (grams per 100 g)"},
		{"kcal_per_100g", nil, "kcal_per_100g is required"},
	}
	for _, tt := range tests {
		body := map[string]any{"name": "Хлебцы", "kcal_per_100g": 310, "fats": 2, "carbs": 60, "proteins": 10}
		body[tt.field] = tt.value
		expectError(t, client.do(http.MethodPost, v1IngredientsPath, body), http.StatusBadRequest, tt.want)
	}
	// kcal_per_100g not sent at all
	expectError(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "Хлебцы", "fats": 2}),
		http.StatusBadRequest, "kcal_per_100g is required")

	if after := v1IngredientRows(t, client); len(after) != len(before) {
		t.Errorf("ingredients = %d, want %d: a rejected ingredient was created", len(after), len(before))
	}
}

func TestIngredientAPICreateRejectsInvalidRequests(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	before := v1IngredientRows(t, client)

	for _, body := range []map[string]any{
		{"kcal_per_100g": 310},
		{"name": nil, "kcal_per_100g": 310},
		{"name": "", "kcal_per_100g": 310},
		{"name": "   ", "kcal_per_100g": 310},
		{"name": invisibleSpace + " \t" + invisibleSpace, "kcal_per_100g": 310},
	} {
		expectError(t, client.do(http.MethodPost, v1IngredientsPath, body), http.StatusBadRequest, "name is required")
	}
	expectError(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": strings.Repeat("я", 201), "kcal_per_100g": 310}),
		http.StatusBadRequest, "name must be at most 200 characters")

	expectError(t, doRaw(client, http.MethodPost, v1IngredientsPath, `{"name": "Хлебцы", "kcal_per_100g": 310`),
		http.StatusBadRequest, "Invalid request body")
	expectError(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "Хлебцы", "kcal_per_100g": "abc"}),
		http.StatusBadRequest, "Invalid request body")
	expectError(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": 5, "kcal_per_100g": 310}),
		http.StatusBadRequest, "Invalid request body")

	if after := v1IngredientRows(t, client); len(after) != len(before) {
		t.Errorf("ingredients = %d, want %d: a rejected ingredient was created", len(after), len(before))
	}
}

func TestIngredientAPICreateRejectsDuplicateNames(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)

	signUp(t, client, smtpServer, "second@example.com")
	otherOats := createWebIngredient(t, client, map[string]any{"name": "Овсянка", "kcalPer100g": 370})
	otherToken, otherAPIKey := client.token, client.apiKey

	signUp(t, client, smtpServer, "first@example.com")
	buckwheatID := createWebIngredient(t, client, map[string]any{"name": "Гречка", "kcalPer100g": 330, "proteins": 12.6}).ID
	hedgehogID := createWebIngredient(t, client, map[string]any{"name": "Ёжик", "kcalPer100g": 100}).ID
	before := v1IngredientRows(t, client)

	// The existing ingredient comes in the external API shape with its own values
	rec := client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "Гречка", "kcal_per_100g": 340})
	want := fmt.Sprintf(`{"message":"%s","existing_ingredient":{"id":%d,"name":"Гречка","kcal_per_100g":330,"proteins":12.6}}`,
		duplicateNameMessage, buckwheatID)
	if got := strings.TrimSpace(rec.Body.String()); rec.Code != http.StatusConflict || got != want {
		t.Errorf("response = %d %s, want 409 %s", rec.Code, got, want)
	}
	for _, name := range []string{"гречка", " Гречка ", "ГРЕЧКА", "Греч" + invisibleSpace + "ка"} {
		expectConflict(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": name, "kcal_per_100g": 340}),
			buckwheatID, "Гречка", "kcal_per_100g")
	}
	expectConflict(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "ежик", "kcal_per_100g": 1}),
		hedgehogID, "Ёжик", "kcal_per_100g")

	after := v1IngredientRows(t, client)
	if len(after) != len(before) || after[buckwheatID] != before[buckwheatID] {
		t.Errorf("list after conflicts: %v, want %v", after, before)
	}

	// Another user's name is not a conflict, and nothing of that user is revealed
	oats, oatsBody := createV1Ingredient(t, client, map[string]any{"name": "Овсянка", "kcal_per_100g": 350})
	if oats.ID == otherOats.ID || !oats.is("Овсянка", 350, nil, nil, nil) {
		t.Errorf("created = %s, want a new ingredient of the user", oatsBody)
	}
	client.token, client.apiKey = otherToken, otherAPIKey
	if others := v1Ingredients(t, client); len(others) != 1 || !others[otherOats.ID].is("Овсянка", 370, nil, nil, nil) {
		t.Errorf("the other user's list = %v, want only the own Овсянка", others)
	}
}

func TestIngredientAPICreateKeepsDiary(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	breadID := addBread(t, client)

	var meal foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food",
		map[string]any{"items": []map[string]any{{"ingredient_id": breadID, "weight": 120}}}, &meal)
	if meal.DayTotalCalories == nil || *meal.DayTotalCalories != 300 {
		t.Fatalf("day total after the meal = %v, want 300", meal.DayTotalCalories)
	}
	dataPath := fmt.Sprintf("/api/v1/data?type=food&from=%s&to=%s", meal.Day, meal.Day)
	diary := func() string {
		t.Helper()
		return strings.TrimSpace(client.expect(http.StatusOK, http.MethodGet, dataPath, nil, nil).Body.String())
	}
	before := diary()

	createV1Ingredient(t, client, map[string]any{"name": "Хлебцы ржаные", "kcal_per_100g": 310})
	expectError(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "Чипсы", "kcal_per_100g": 1000}),
		http.StatusBadRequest, "kcal_per_100g")
	conflict(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "хлеб", "kcal_per_100g": 250}))

	after := diary()
	if after != before {
		t.Errorf("diary changed: %s, want %s", after, before)
	}
	var data struct {
		Food []foodAPIEntry `json:"food"`
	}
	if err := json.Unmarshal([]byte(after), &data); err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, entry := range data.Food {
		total += entry.Calories
		if entry.Food != "Хлеб" {
			t.Errorf("unexpected diary entry %+v", entry)
		}
	}
	if len(data.Food) != 1 || total != 300 {
		t.Errorf("diary = %+v, want the one meal with 300 kcal", data.Food)
	}
}

// The statistics are matched by name: entries that already have exactly the new name count, as in the list
func TestIngredientAPICreateShowsUsageOfTheName(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")

	for _, datetime := range []string{"2026-10-10T08:00:00Z", "2026-10-12T08:00:00Z"} {
		client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food", map[string]any{
			"meal_datetime": datetime,
			"items": []map[string]any{{"new_ingredient": map[string]any{"name": "Кофе раф", "kcal_per_100g": 80},
				"weight": 250, "one_off": true}},
		}, nil)
	}

	created, body := createV1Ingredient(t, client, map[string]any{"name": "Кофе раф", "kcal_per_100g": 80})
	if want := fmt.Sprintf(`{"id":%d,"name":"Кофе раф","kcal_per_100g":80,"times_used":2,"last_used":"2026-10-12"}`,
		created.ID); body != want {
		t.Errorf("response = %s, want %s", body, want)
	}
	if row := v1IngredientRows(t, client)[created.ID]; row != body {
		t.Errorf("list row = %s, want the response %s", row, body)
	}
}

func TestIngredientAPICreateRequiresAPIKey(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "food@example.com")
	token, apiKey := client.token, client.apiKey
	before := len(webIngredients(t, client))

	var revoked struct {
		ID  int    `json:"id"`
		Key string `json:"key"`
	}
	client.expect(http.StatusCreated, http.MethodPost, "/api/api-keys", map[string]any{"name": "old"}, &revoked)
	client.expect(http.StatusOK, http.MethodPost, fmt.Sprintf("/api/api-keys/%d/revoke", revoked.ID), nil, nil)

	body := map[string]any{"name": "Хлебцы ржаные", "kcal_per_100g": 310}
	for _, auth := range []struct{ token, apiKey string }{
		{token, ""},         // a website session only
		{"", ""},            // nothing
		{"", "invalid-key"}, // an unknown key
		{"", revoked.Key},   // a revoked key
	} {
		client.token, client.apiKey = auth.token, auth.apiKey
		client.expect(http.StatusUnauthorized, http.MethodPost, v1IngredientsPath, body, nil)
	}

	client.token, client.apiKey = token, apiKey
	if after := len(webIngredients(t, client)); after != before {
		t.Errorf("ingredients = %d, want %d: created without a valid API key", after, before)
	}
}
