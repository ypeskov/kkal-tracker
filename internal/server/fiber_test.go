package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

const fiberLimitMessage = "fiber must be between 0 and 100 (grams per 100 g)"

// hasFiberKey reports whether a raw JSON object has the "fiber" key at its top level
func hasFiberKey(t *testing.T, raw string) bool {
	t.Helper()
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		t.Fatalf("not a JSON object: %s", raw)
	}
	_, ok := object["fiber"]
	return ok
}

// dayEntries returns the diary entries of a day through the external API, raw, by ID
func dayEntries(t *testing.T, client *smokeClient, day string) map[int]string {
	t.Helper()
	var data struct {
		Food []json.RawMessage `json:"food"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/v1/data?type=food&from="+day+"&to="+day, nil, &data)
	entries := make(map[int]string, len(data.Food))
	for _, raw := range data.Food {
		var entry struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw, &entry); err != nil {
			t.Fatalf("invalid entry %s: %v", raw, err)
		}
		entries[entry.ID] = string(raw)
	}
	return entries
}

// webEntry is a diary entry in the shape of the website API (/api/calories)
type webEntry struct {
	ID          int      `json:"id"`
	Food        string   `json:"food"`
	Calories    int      `json:"calories"`
	Weight      float64  `json:"weight"`
	KcalPer100g float64  `json:"kcalPer100g"`
	Fats        *float64 `json:"fats"`
	Carbs       *float64 `json:"carbs"`
	Proteins    *float64 `json:"proteins"`
	Fiber       *float64 `json:"fiber"`
}

// webEntries returns the diary entries of a day through the website API by ID. The tests read the stored state
// with it where the shape does not matter: these requests do not count against the burst of the api-v1 rate limiter.
func webEntries(t *testing.T, client *smokeClient, day string) map[int]webEntry {
	t.Helper()
	var entries []webEntry
	client.expect(http.StatusOK, http.MethodGet, "/api/calories?dateFrom="+day+"&dateTo="+day, nil, &entries)
	byID := make(map[int]webEntry, len(entries))
	for _, entry := range entries {
		byID[entry.ID] = entry
	}
	return byID
}

// Every test below stays under the burst of the api-v1 rate limiter: all requests of a test come from one address

func TestFiberIngredientAPICreate(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "fiber@example.com")

	// A value is returned and listed, after proteins
	oats, body := createV1Ingredient(t, client, map[string]any{"name": "Овсянка", "kcal_per_100g": 370, "fiber": 10.6})
	if want := fmt.Sprintf(`{"id":%d,"name":"Овсянка","kcal_per_100g":370,"fiber":10.6,"times_used":0}`, oats.ID); body != want {
		t.Errorf("response = %s, want %s", body, want)
	}
	full, fullBody := createV1Ingredient(t, client,
		map[string]any{"name": "Хлебцы", "kcal_per_100g": 310, "fats": 2, "carbs": 60, "proteins": 10, "fiber": 14})
	if want := fmt.Sprintf(`{"id":%d,"name":"Хлебцы","kcal_per_100g":310,"fats":2,"carbs":60,"proteins":10,"fiber":14,"times_used":0}`,
		full.ID); fullBody != want {
		t.Errorf("response = %s, want %s", fullBody, want)
	}

	// Zero is a value, not unknown
	meat, meatBody := createV1Ingredient(t, client, map[string]any{"name": "Мясо", "kcal_per_100g": 200, "fiber": 0})
	if want := fmt.Sprintf(`{"id":%d,"name":"Мясо","kcal_per_100g":200,"fiber":0,"times_used":0}`, meat.ID); meatBody != want {
		t.Errorf("response with fiber 0 = %s, want %s", meatBody, want)
	}

	// Not sent or null is unknown: no key at all
	sauce, sauceBody := createV1Ingredient(t, client, map[string]any{"name": "Соус", "kcal_per_100g": 120})
	null, nullBody := createV1Ingredient(t, client, map[string]any{"name": "Кетчуп", "kcal_per_100g": 100, "fiber": nil})
	if hasFiberKey(t, sauceBody) || hasFiberKey(t, nullBody) {
		t.Errorf("responses without fiber = %s, %s; want no fiber key", sauceBody, nullBody)
	}

	// The upper limit is inclusive
	if bran, _ := createV1Ingredient(t, client, map[string]any{"name": "Отруби", "kcal_per_100g": 200, "fiber": 100}); optional(bran.Fiber) != 100.0 {
		t.Errorf("fiber 100: %v", optional(bran.Fiber))
	}

	rows := v1IngredientRows(t, client)
	for id, want := range map[int]string{oats.ID: body, full.ID: fullBody, meat.ID: meatBody, sauce.ID: sauceBody, null.ID: nullBody} {
		if rows[id] != want {
			t.Errorf("list row = %s, want the response %s", rows[id], want)
		}
	}

	// Out of range: nothing is created
	before := len(rows)
	for _, value := range []any{-1, 100.1, -0.01} {
		rec := client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "Плохой", "kcal_per_100g": 100, "fiber": value})
		expectError(t, rec, http.StatusBadRequest, fiberLimitMessage)
	}
	if after := len(webIngredients(t, client)); after != before {
		t.Errorf("ingredients after rejected creates = %d, want %d", after, before)
	}

	// The validation order is fats, carbs, proteins, fiber
	rec := client.do(http.MethodPost, v1IngredientsPath,
		map[string]any{"name": "Плохой", "kcal_per_100g": 100, "proteins": 101, "fiber": 101})
	expectError(t, rec, http.StatusBadRequest, "proteins must be between 0 and 100")

}

func TestFiberIngredientAPIEdit(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "fiber@example.com")

	createV1Ingredient(t, client, map[string]any{"name": "Овсянка", "kcal_per_100g": 370, "fiber": 10.6})
	createWebIngredient(t, client, map[string]any{"name": "Соус", "kcalPer100g": 120})
	bread, _ := createV1Ingredient(t, client,
		map[string]any{"name": "Хлеб", "kcal_per_100g": 250, "fats": 1, "carbs": 50, "proteins": 8})
	if edited := editIngredient(t, client, bread.ID, map[string]any{"fiber": 2.5}); !edited.is("Хлеб", 250, 1.0, 50.0, 8.0) ||
		optional(edited.Fiber) != 2.5 {
		t.Errorf("fiber only: %v, fiber %v; want only the fiber changed to 2.5", edited, optional(edited.Fiber))
	}
	if edited := editIngredient(t, client, bread.ID, map[string]any{"kcal_per_100g": 260}); !edited.is("Хлеб", 260, 1.0, 50.0, 8.0) ||
		optional(edited.Fiber) != 2.5 {
		t.Errorf("kcal only: %v, fiber %v; want the fiber kept", edited, optional(edited.Fiber))
	}

	path := ingredientPath(bread.ID)
	expectError(t, doRaw(client, http.MethodPut, path, `{"fiber": null}`), http.StatusBadRequest,
		"nothing to change: send name, kcal_per_100g, fats, carbs, proteins or fiber")
	expectError(t, doRaw(client, http.MethodPut, path, `{"fiber": 150}`), http.StatusBadRequest, fiberLimitMessage)
	expectError(t, doRaw(client, http.MethodPut, path, `{"kcal_per_100g": 300, "fiber": -1}`), http.StatusBadRequest, fiberLimitMessage)
	if kept := webIngredients(t, client)[bread.ID]; kept.Name != "Хлеб" || kept.KcalPer100g != 260 || optional(kept.Fiber) != 2.5 {
		t.Errorf("after rejected edits: %+v; want nothing changed", kept)
	}

	// null next to another change keeps the fibre; it cannot be reset to unknown, only set to 0
	var response struct {
		Ingredient apiIngredient `json:"ingredient"`
	}
	if rec := doRaw(client, http.MethodPut, path, `{"fiber": null, "fats": 2}`); rec.Code != http.StatusOK {
		t.Fatalf("null with fats: status = %d; body: %s", rec.Code, rec.Body.String())
	} else if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || optional(response.Ingredient.Fiber) != 2.5 ||
		optional(response.Ingredient.Fats) != 2.0 {
		t.Errorf("null with fats: %s; want fiber 2.5 kept and fats 2", rec.Body.String())
	}
	rec := doRaw(client, http.MethodPut, path, `{"fiber": 0}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"fiber":0`) {
		t.Errorf("fiber 0: status = %d, body %s; want 200 with the fiber 0", rec.Code, rec.Body.String())
	}
	if listed := v1IngredientRows(t, client)[bread.ID]; !strings.Contains(listed, `"fiber":0,`) {
		t.Errorf("list row after fiber 0 = %s", listed)
	}

	// 409 answers show the fibre of the existing ingredient
	existing := conflict(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "овсянка", "kcal_per_100g": 1}))
	if existing["fiber"] != 10.6 {
		t.Errorf("existing_ingredient of a create = %v, want fiber 10.6", existing)
	}
	existing = conflict(t, client.do(http.MethodPut, path, map[string]any{"name": "Овсянка"}))
	if existing["fiber"] != 10.6 {
		t.Errorf("existing_ingredient of a rename = %v, want fiber 10.6", existing)
	}
	duplicate := client.do(http.MethodPost, "/api/v1/food", map[string]any{"items": []map[string]any{
		{"new_ingredient": map[string]any{"name": "ОВСЯНКА", "kcal_per_100g": 1}, "weight": 10}}})
	var duplicateResponse struct {
		Existing map[string]any `json:"existing_ingredient"`
	}
	if err := json.Unmarshal(duplicate.Body.Bytes(), &duplicateResponse); err != nil || duplicate.Code != http.StatusConflict ||
		duplicateResponse.Existing["fiber"] != 10.6 {
		t.Errorf("409 of a meal: status %d, body %s; want existing_ingredient with fiber 10.6", duplicate.Code, duplicate.Body.String())
	}
	if _, ok := conflict(t, client.do(http.MethodPost, v1IngredientsPath, map[string]any{"name": "соус", "kcal_per_100g": 1}))["fiber"]; ok {
		t.Error("existing_ingredient without fibre has a fiber key")
	}
}

func TestFiberFoodAPICreate(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "fiber@example.com")

	oats, _ := createV1Ingredient(t, client, map[string]any{"name": "Овсянка", "kcal_per_100g": 370, "fiber": 10.6})
	// An ingredient created before the feature: its fibre is unknown
	breadID := addBread(t, client)

	const day = "2026-03-10"
	post := func(items ...map[string]any) *foodAPIMeal {
		t.Helper()
		var meal foodAPIMeal
		client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food",
			map[string]any{"meal_datetime": day + "T08:00:00Z", "items": items}, &meal)
		return &meal
	}

	meal := post(
		// The entry copies the fibre of the ingredient
		map[string]any{"ingredient_id": oats.ID, "weight": 50},
		// An override is for this entry only
		map[string]any{"ingredient_id": oats.ID, "weight": 50, "fiber": 8},
		// Unknown fibre of the ingredient stays unknown
		map[string]any{"ingredient_id": breadID, "weight": 100},
		// A new food: the entry and the ingredient get the value
		map[string]any{"new_ingredient": map[string]any{"name": "Груша", "kcal_per_100g": 100, "fiber": 3}, "weight": 100},
		// A one-off food: the entry only
		map[string]any{"new_ingredient": map[string]any{"name": "Салат в кафе", "kcal_per_100g": 100, "fiber": 3}, "weight": 100, "one_off": true},
		// An override of a value the ingredient does not have, and the inclusive limits
		map[string]any{"ingredient_id": breadID, "weight": 100, "fiber": 0},
		map[string]any{"ingredient_id": breadID, "weight": 100, "fiber": 100},
		// A new food without fibre
		map[string]any{"new_ingredient": map[string]any{"name": "Вода", "kcal_per_100g": 0}, "weight": 100, "one_off": true},
	)
	wantFiber := []any{10.6, 8.0, nil, 3.0, 3.0, 0.0, 100.0, nil}
	stored := dayEntries(t, client, day)
	web := webEntries(t, client, day)
	for i, want := range wantFiber {
		entry := meal.Entries[i]
		if optional(entry.Fiber) != want {
			t.Errorf("created entry %d: fiber = %v, want %v", i, optional(entry.Fiber), want)
		}
		if got := optional(web[entry.ID].Fiber); got != want {
			t.Errorf("stored entry %d: fiber = %v, want %v", i, got, want)
		}
		// An unknown value has no key at all
		if hasFiberKey(t, stored[entry.ID]) != (want != nil) {
			t.Errorf("stored entry %d = %s, fiber key present = %v", i, stored[entry.ID], want == nil)
		}
	}
	// Calories do not depend on fibre: 185 + 185 + 250 + 100 + 100 + 250 + 250 + 0
	if meal.TotalCalories != 1320 || meal.DayTotalCalories == nil || *meal.DayTotalCalories != 1320 {
		t.Errorf("totals = %d / %v, want 1320 / 1320", meal.TotalCalories, meal.DayTotalCalories)
	}

	ingredients := webIngredients(t, client)
	if got := optional(ingredients[oats.ID].Fiber); got != 10.6 {
		t.Errorf("fibre of the ingredient after an override = %v, want 10.6", got)
	}
	if ingredients[breadID].Fiber != nil {
		t.Errorf("fibre of the bread = %v, want unknown", *ingredients[breadID].Fiber)
	}
	pear := ingredients[*meal.Entries[3].IngredientID]
	if !meal.Entries[3].IngredientCreated || pear.Name != "Груша" || optional(pear.Fiber) != 3.0 {
		t.Errorf("new ingredient = %+v; want Груша with fiber 3", pear)
	}
	if meal.Entries[4].IngredientID != nil {
		t.Errorf("one-off entry = %+v, want no ingredient", meal.Entries[4])
	}
	for _, ingredient := range ingredients {
		if ingredient.Name == "Салат в кафе" || ingredient.Name == "Вода" {
			t.Errorf("one-off food %q is in the ingredient list", ingredient.Name)
		}
	}

	// Rejected meals: the item is named and nothing of the meal is stored
	entriesBefore, ingredientsBefore := len(stored), len(ingredients)
	for _, tt := range []struct {
		item map[string]any
		want string
	}{
		{map[string]any{"new_ingredient": map[string]any{"name": "X", "kcal_per_100g": 100}, "weight": 100, "fiber": 3},
			"items[1]: fiber of a new food belongs inside new_ingredient"},
		{map[string]any{"new_ingredient": map[string]any{"name": "X", "kcal_per_100g": 100}, "weight": 100, "fiber": 3, "one_off": true},
			"items[1]: fiber of a new food belongs inside new_ingredient"},
		{map[string]any{"ingredient_id": oats.ID, "weight": 100, "fiber": 101}, "items[1]: " + fiberLimitMessage},
		{map[string]any{"ingredient_id": oats.ID, "weight": 100, "fiber": -0.1}, "items[1]: " + fiberLimitMessage},
		{map[string]any{"new_ingredient": map[string]any{"name": "X", "kcal_per_100g": 100, "fiber": 101}, "weight": 100},
			"items[1]: new_ingredient." + fiberLimitMessage},
		{map[string]any{"new_ingredient": map[string]any{"name": "X", "kcal_per_100g": 100, "fiber": -1}, "weight": 100, "one_off": true},
			"items[1]: new_ingredient." + fiberLimitMessage},
	} {
		rec := client.do(http.MethodPost, "/api/v1/food", map[string]any{
			"meal_datetime": day + "T12:00:00Z",
			"items": []map[string]any{
				{"new_ingredient": map[string]any{"name": "Первый", "kcal_per_100g": 100, "fiber": 1}, "weight": 50}, tt.item},
		})
		expectError(t, rec, http.StatusBadRequest, tt.want)
	}
	if got := len(webEntries(t, client, day)); got != entriesBefore {
		t.Errorf("entries after rejected meals = %d, want %d", got, entriesBefore)
	}
	if got := len(webIngredients(t, client)); got != ingredientsBefore {
		t.Errorf("ingredients after rejected meals = %d, want %d", got, ingredientsBefore)
	}
}

func TestFiberFoodAPIEdit(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "fiber@example.com")

	oats, _ := createV1Ingredient(t, client, map[string]any{"name": "Овсянка", "kcal_per_100g": 370, "fiber": 10.6})
	breadID := addBread(t, client)

	const day = "2026-03-10"
	var meal foodAPIMeal
	client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food", map[string]any{
		"meal_datetime": day + "T08:00:00Z",
		"items":         []map[string]any{{"ingredient_id": oats.ID, "weight": 50}, {"ingredient_id": breadID, "weight": 100}},
	}, &meal)
	id := meal.Entries[1].ID // 100 g of bread, fibre unknown
	entryPath := fmt.Sprintf("/api/v1/food/%d", id)
	type updated struct {
		Entry            foodAPIEntry `json:"entry"`
		DayTotalCalories *int         `json:"day_total_calories"`
	}
	edit := func(step string, body any, wantFiber any) foodAPIEntry {
		t.Helper()
		var result updated
		client.expect(http.StatusOK, http.MethodPut, entryPath, body, &result)
		if optional(result.Entry.Fiber) != wantFiber {
			t.Errorf("%s: fiber = %v, want %v", step, optional(result.Entry.Fiber), wantFiber)
		}
		if got := optional(webEntries(t, client, day)[id].Fiber); got != wantFiber {
			t.Errorf("%s: stored fiber = %v, want %v", step, got, wantFiber)
		}
		return result.Entry
	}

	if e := edit("fiber only", map[string]any{"fiber": 4}, 4.0); !e.hasValues("Хлеб", 250, 1.0, 50.0, 8.0, 250) ||
		e.Weight != 100 || e.MealDatetime != meal.Entries[1].MealDatetime {
		t.Errorf("fiber only changed other values: %v at %s", e, e.MealDatetime)
	}
	if e := edit("weight only", map[string]any{"weight": 120}, 4.0); e.Weight != 120 || e.Calories != 300 {
		t.Errorf("weight only: %v", e)
	}
	// Another ingredient replaces the fibre, an explicit value goes on top
	if e := edit("ingredient only", map[string]any{"ingredient_id": oats.ID}, 10.6); e.Food != "Овсянка" || e.Fats != nil {
		t.Errorf("ingredient only: %v", e)
	}
	edit("ingredient and fiber", map[string]any{"ingredient_id": oats.ID, "fiber": 7}, 7.0)
	// An ingredient with unknown fibre makes the fibre of the entry unknown, like the macros
	edit("ingredient with unknown fiber", map[string]any{"ingredient_id": breadID}, nil)
	if hasFiberKey(t, dayEntries(t, client, day)[id]) {
		t.Error("entry switched to an ingredient without fibre still has a fiber key")
	}
	edit("fiber zero", map[string]any{"fiber": 0}, 0.0)
	edit("fiber again", map[string]any{"fiber": 4}, 4.0)

	before := webEntries(t, client, day)[id]
	for _, tt := range []struct{ body, want string }{
		{`{"fiber": null}`, "nothing to change: send weight, ingredient_id, kcal_per_100g, fats, carbs, proteins, fiber or meal_datetime"},
		{`{"fiber": -0.1}`, fiberLimitMessage},
		{`{"fiber": 100.5, "weight": 10}`, fiberLimitMessage},
		{`{"ingredient_id": ` + fmt.Sprint(oats.ID) + `, "fiber": 101}`, fiberLimitMessage},
	} {
		expectError(t, doRaw(client, http.MethodPut, entryPath, tt.body), http.StatusBadRequest, tt.want)
	}
	if after := webEntries(t, client, day)[id]; after.Food != before.Food || after.Weight != before.Weight ||
		after.Calories != before.Calories || optional(after.Fiber) != 4.0 {
		t.Errorf("entry after rejected changes = %+v, want %+v with fiber 4", after, before)
	}
	// null next to another change keeps the value
	edit("null with weight", json.RawMessage(`{"fiber": null, "weight": 100}`), 4.0)

	// Editing the ingredient afterwards does not change the entries logged before
	editIngredient(t, client, oats.ID, map[string]any{"fiber": 12})
	if got := optional(fetchEntry(t, client, day, meal.Entries[0].ID).Fiber); got != 10.6 {
		t.Errorf("logged entry after an ingredient edit: fiber = %v, want 10.6", got)
	}
	// The ingredient with unknown fibre is still unknown
	if bread := webIngredients(t, client)[breadID]; bread.Fiber != nil {
		t.Errorf("fibre of the bread = %v, want unknown", *bread.Fiber)
	}
}

func TestFiberWebIngredients(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "fiber@example.com")

	var created webIngredient
	rec := client.expect(http.StatusCreated, http.MethodPost, "/api/ingredients",
		map[string]any{"name": "N", "kcalPer100g": 100, "fiber": 5}, &created)
	if optional(created.Fiber) != 5.0 || !strings.Contains(rec.Body.String(), `"fiber":5`) {
		t.Errorf("created = %s, want fiber 5", rec.Body.String())
	}
	path := fmt.Sprintf("/api/ingredients/%d", created.ID)
	var single webIngredient
	client.expect(http.StatusOK, http.MethodGet, path, nil, &single)
	if optional(single.Fiber) != 5.0 || optional(webIngredients(t, client)[created.ID].Fiber) != 5.0 {
		t.Errorf("GET by id and the list: %v, %v; want fiber 5", optional(single.Fiber), optional(webIngredients(t, client)[created.ID].Fiber))
	}
	// The same value through the external API
	if got := optional(v1Ingredients(t, client)[created.ID].Fiber); got != 5.0 {
		t.Errorf("/api/v1/ingredients: fiber = %v, want 5", got)
	}

	// Without fibre the request works as before and the response has no key
	rec = client.expect(http.StatusCreated, http.MethodPost, "/api/ingredients",
		map[string]any{"name": "Без клетчатки", "kcalPer100g": 100, "fats": 1}, nil)
	if hasFiberKey(t, rec.Body.String()) {
		t.Errorf("created without fiber = %s, want no fiber key", rec.Body.String())
	}

	// The limits are inclusive
	zero := createWebIngredient(t, client, map[string]any{"name": "Ноль", "kcalPer100g": 100, "fiber": 0})
	hundred := createWebIngredient(t, client, map[string]any{"name": "Сто", "kcalPer100g": 100, "fiber": 100})
	if optional(zero.Fiber) != 0.0 || optional(hundred.Fiber) != 100.0 {
		t.Errorf("limits: %v, %v; want 0 and 100", optional(zero.Fiber), optional(hundred.Fiber))
	}

	// PUT replaces all values: 0 is stored, an absent fibre is cleared like the macros
	var updated webIngredient
	rec = client.expect(http.StatusOK, http.MethodPut, path, map[string]any{"name": "N", "kcalPer100g": 100, "fiber": 0}, &updated)
	if optional(updated.Fiber) != 0.0 || !strings.Contains(rec.Body.String(), `"fiber":0`) {
		t.Errorf("PUT with fiber 0 = %s", rec.Body.String())
	}
	client.expect(http.StatusOK, http.MethodPut, path, map[string]any{"name": "N", "kcalPer100g": 100, "fiber": 7.5}, &updated)
	if optional(updated.Fiber) != 7.5 {
		t.Errorf("PUT with fiber 7.5: %v", optional(updated.Fiber))
	}

	// Out of range: 400 and nothing stored or changed
	count := len(webIngredients(t, client))
	for _, value := range []any{-1, 101, 100.01} {
		expectError(t, client.do(http.MethodPost, "/api/ingredients", map[string]any{"name": "Плохой", "kcalPer100g": 100, "fiber": value}),
			http.StatusBadRequest)
		expectError(t, client.do(http.MethodPut, path, map[string]any{"name": "Другой", "kcalPer100g": 200, "fiber": value}),
			http.StatusBadRequest)
	}
	if after := webIngredients(t, client); len(after) != count || after[created.ID].Name != "N" ||
		after[created.ID].KcalPer100g != 100 || optional(after[created.ID].Fiber) != 7.5 {
		t.Errorf("after rejected requests: %d ingredients, %+v; want %d and the ingredient unchanged", len(after), after[created.ID], count)
	}

	rec = client.expect(http.StatusOK, http.MethodPut, path, map[string]any{"name": "N", "kcalPer100g": 100}, &updated)
	if hasFiberKey(t, rec.Body.String()) || webIngredients(t, client)[created.ID].Fiber != nil {
		t.Errorf("PUT without fiber = %s, want the fiber cleared", rec.Body.String())
	}

	// The validation of the macros is as before: no upper limit, negative values rejected
	if fat := createWebIngredient(t, client, map[string]any{"name": "Жир", "kcalPer100g": 100, "fats": 150}); optional(fat.Fats) != 150.0 {
		t.Errorf("fats 150 on the website: %v, want it accepted", optional(fat.Fats))
	}
	expectError(t, client.do(http.MethodPost, "/api/ingredients", map[string]any{"name": "Минус", "kcalPer100g": 100, "fats": -1}),
		http.StatusBadRequest)
}

func TestFiberWebCalories(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "fiber@example.com")

	request := func(food string, extra map[string]any) map[string]any {
		body := map[string]any{"food": food, "calories": 150, "weight": 100.0, "kcalPer100g": 150.0,
			"meal_datetime": "2026-03-10T08:30:00Z"}
		for key, value := range extra {
			body[key] = value
		}
		return body
	}
	create := func(food string, extra map[string]any) (webEntry, string) {
		t.Helper()
		var result struct {
			Entry                webEntry `json:"entry"`
			NewIngredientCreated bool     `json:"new_ingredient_created"`
		}
		rec := client.expect(http.StatusCreated, http.MethodPost, "/api/calories", request(food, extra), &result)
		return result.Entry, rec.Body.String()
	}
	listed := func() map[int]webEntry {
		t.Helper()
		return webEntries(t, client, "2026-03-10")
	}
	ingredientFiber := func(name string) (any, bool) {
		t.Helper()
		for _, ingredient := range webIngredients(t, client) {
			if ingredient.Name == name {
				return optional(ingredient.Fiber), true
			}
		}
		return nil, false
	}

	// The entry and the ingredient created for a new name get the fibre
	oats, _ := create("Овсянка новая", map[string]any{"fiber": 3})
	if optional(oats.Fiber) != 3.0 {
		t.Errorf("entry: fiber = %v, want 3", optional(oats.Fiber))
	}
	if fiber, ok := ingredientFiber("Овсянка новая"); !ok || fiber != 3.0 {
		t.Errorf("ingredient created for the new name: fiber = %v (found %v), want 3", fiber, ok)
	}
	// Zero is a value
	meat, meatBody := create("Мясо новое", map[string]any{"fiber": 0})
	if optional(meat.Fiber) != 0.0 || !strings.Contains(meatBody, `"fiber":0`) {
		t.Errorf("entry with fiber 0 = %s", meatBody)
	}
	if fiber, ok := ingredientFiber("Мясо новое"); !ok || fiber != 0.0 {
		t.Errorf("ingredient created with fiber 0: %v (found %v)", fiber, ok)
	}
	// Without fibre everything is as before, no key in the response; the macros are still not range-checked
	plain, plainBody := create("Просто еда", map[string]any{"fats": 150})
	if plain.Fiber != nil || strings.Contains(plainBody, `"fiber"`) || optional(plain.Fats) != 150.0 {
		t.Errorf("entry without fiber = %s, want no fiber key and fats 150", plainBody)
	}
	if fiber, ok := ingredientFiber("Просто еда"); !ok || fiber != nil {
		t.Errorf("ingredient created without fiber: %v (found %v), want unknown", fiber, ok)
	}
	if entries := listed(); optional(entries[oats.ID].Fiber) != 3.0 || optional(entries[meat.ID].Fiber) != 0.0 || entries[plain.ID].Fiber != nil {
		t.Errorf("GET /api/calories: %+v", entries)
	}

	// PUT replaces all values
	entryPath := fmt.Sprintf("/api/calories/%d", oats.ID)
	var updated webEntry
	client.expect(http.StatusOK, http.MethodPut, entryPath, request("Овсянка новая", map[string]any{"fiber": 6}), &updated)
	if optional(updated.Fiber) != 6.0 || optional(listed()[oats.ID].Fiber) != 6.0 {
		t.Errorf("PUT with fiber 6: %v", optional(updated.Fiber))
	}

	// Out of range: 400, nothing stored or changed, no internal error text
	count := len(listed())
	for _, value := range []any{-1, 101} {
		expectError(t, client.do(http.MethodPost, "/api/calories", request("Плохая еда", map[string]any{"fiber": value})),
			http.StatusBadRequest)
		expectError(t, client.do(http.MethodPut, entryPath, request("Другая еда", map[string]any{"fiber": value, "weight": 55.0})),
			http.StatusBadRequest)
	}
	if entries := listed(); len(entries) != count || entries[oats.ID].Food != "Овсянка новая" ||
		entries[oats.ID].Weight != 100 || optional(entries[oats.ID].Fiber) != 6.0 {
		t.Errorf("after rejected requests: %d entries, %+v; want %d and the entry unchanged", len(entries), entries[oats.ID], count)
	}
	if _, ok := ingredientFiber("Плохая еда"); ok {
		t.Error("a rejected entry created an ingredient")
	}

	rec := client.expect(http.StatusOK, http.MethodPut, entryPath, request("Овсянка новая", map[string]any{"fiber": 0}), &updated)
	if optional(updated.Fiber) != 0.0 || !strings.Contains(rec.Body.String(), `"fiber":0`) {
		t.Errorf("PUT with fiber 0 = %s", rec.Body.String())
	}
	rec = client.expect(http.StatusOK, http.MethodPut, entryPath, request("Овсянка новая", nil), &updated)
	if strings.Contains(rec.Body.String(), `"fiber"`) || listed()[oats.ID].Fiber != nil {
		t.Errorf("PUT without fiber = %s, want the fiber cleared", rec.Body.String())
	}
}

// A new user gets the global ingredients with their fibre: unknown stays unknown, a value is copied
func TestFiberIsCopiedFromGlobalIngredients(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)

	for name, fiber := range map[string]any{"Global oats": 10.6, "Global meat": 0, "Global water": nil} {
		result, err := client.db.Exec(`INSERT INTO global_ingredients (kcal_per_100g, fiber) VALUES (100, ?)`, fiber)
		if err != nil {
			t.Fatal(err)
		}
		id, _ := result.LastInsertId()
		if _, err := client.db.Exec(`INSERT INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (?, 'en_US', ?)`,
			id, name); err != nil {
			t.Fatal(err)
		}
	}
	signUp(t, client, smtpServer, "fiber@example.com")

	got := map[string]any{}
	for _, ingredient := range v1Ingredients(t, client) {
		got[ingredient.Name] = optional(ingredient.Fiber)
	}
	if len(got) != 3 || got["Global oats"] != 10.6 || got["Global meat"] != 0.0 || got["Global water"] != nil {
		t.Errorf("copied ingredients = %v, want the fibre of the global ones", got)
	}
}

// exportedFoodSheet downloads the Excel export of a day in the given language and returns the file and its food sheet
func exportedFoodSheet(t *testing.T, client *smokeClient, day string) (*excelize.File, string) {
	t.Helper()
	download := client.expect(http.StatusOK, http.MethodPost, "/api/export",
		map[string]string{"date_from": day, "date_to": day, "data_type": "food", "delivery_type": "download"}, nil)
	file, err := excelize.OpenReader(bytes.NewReader(download.Body.Bytes()))
	if err != nil {
		t.Fatalf("export is not a readable XLSX file: %v", err)
	}
	t.Cleanup(func() { file.Close() })
	sheets := file.GetSheetList()
	if len(sheets) != 1 {
		t.Fatalf("sheets = %q, want the food sheet only", sheets)
	}
	return file, sheets[0]
}

// The scenario of the task end to end: ingredient with fibre, a meal with a known, a zero and an unknown value,
// the totals of the skill client, a later ingredient edit, and the export
func TestFiberEndToEnd(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "fiber@example.com")
	env := newSkillClientEnv(t, client)

	stdout, stderr, code := env.run("add-ingredient", "--name", "Отруби", "--kcal", "200", "--fiber", "10")
	if code != 0 {
		t.Fatalf("add-ingredient: exit code %d; stderr: %s", code, stderr)
	}
	var bran apiIngredient
	if err := json.Unmarshal([]byte(stdout), &bran); err != nil || optional(bran.Fiber) != 10.0 {
		t.Fatalf("add-ingredient output = %s, want the ingredient with fiber 10", stdout)
	}

	const day = "2026-03-10"
	mealJSON := fmt.Sprintf(`{"meal_datetime": "%sT12:00:00Z", "items": [
		{"ingredient_id": %d, "weight": 50},
		{"new_ingredient": {"name": "Стейк", "kcal_per_100g": 250, "fiber": 0}, "weight": 100, "one_off": true},
		{"new_ingredient": {"name": "Суп в кафе", "kcal_per_100g": 60}, "weight": 100, "one_off": true}]}`, day, bran.ID)
	stdout, stderr, code = env.run("add", "--json", mealJSON)
	if code != 0 {
		t.Fatalf("add: exit code %d; stderr: %s", code, stderr)
	}
	var meal foodAPIMeal
	if err := json.Unmarshal([]byte(stdout), &meal); err != nil || len(meal.Entries) != 3 {
		t.Fatalf("add output = %s", stdout)
	}
	if e := meal.Entries; optional(e[0].Fiber) != 10.0 || optional(e[1].Fiber) != 0.0 || e[2].Fiber != nil {
		t.Errorf("entries: fiber = %v, %v, %v; want 10, 0 and none", optional(e[0].Fiber), optional(e[1].Fiber), optional(e[2].Fiber))
	}
	env.takeRequests()

	branEntry, steak, soup := meal.Entries[0].ID, meal.Entries[1].ID, meal.Entries[2].ID
	output, _ := env.totals("--day", day, "--meal", fmt.Sprint(branEntry), fmt.Sprint(steak), fmt.Sprint(soup),
		"--expect-meal-kcal", fmt.Sprint(meal.TotalCalories), "--expect-day-kcal", fmt.Sprint(*meal.DayTotalCalories))
	// 50 g x 10 / 100 + 100 g x 0 / 100 + unknown as 0
	if output.Meal == nil || output.Meal.Fiber != 5 || output.DayTotal.Fiber != 5 {
		t.Errorf("totals: meal %+v, day %+v; want fiber 5 in both", output.Meal, output.DayTotal)
	}
	if w := output.WithoutFiber; len(w) != 1 || w[0].ID != soup || w[0].Food != "Суп в кафе" || !w[0].InMeal {
		t.Errorf("without_fiber = %+v, want only the soup, in the meal", w)
	}
	checkResult(t, "meal", output.Check.Meal, "ok")
	checkResult(t, "day", output.Check.Day, "ok")
	// Calories are not affected by fibre: 100 + 250 + 60
	if meal.TotalCalories != 410 || *meal.DayTotalCalories != 410 {
		t.Errorf("calories = %d / %d, want 410 / 410", meal.TotalCalories, *meal.DayTotalCalories)
	}

	// The website shows the same values (its day view sums them on the client)
	webFiber := map[int]any{}
	for id, entry := range webEntries(t, client, day) {
		webFiber[id] = optional(entry.Fiber)
	}
	if len(webFiber) != 3 || webFiber[branEntry] != 10.0 || webFiber[steak] != 0.0 || webFiber[soup] != nil {
		t.Errorf("website entries: fiber = %v, want 10, 0 and unknown", webFiber)
	}
	if got := optional(webIngredients(t, client)[bran.ID].Fiber); got != 10.0 {
		t.Errorf("website ingredient: fiber = %v, want 10", got)
	}

	// The export has the three values in column J
	file, sheet := exportedFoodSheet(t, client, day)
	if header, _ := file.GetCellValue(sheet, "J1"); header != "Fiber (g)" {
		t.Errorf("J1 = %q, want the translated fibre header", header)
	}
	exported := map[string]string{}
	for row := 2; row <= 4; row++ {
		food, _ := file.GetCellValue(sheet, fmt.Sprintf("C%d", row))
		fiber, _ := file.GetCellValue(sheet, fmt.Sprintf("J%d", row))
		exported[food] = fiber
	}
	if len(exported) != 3 || exported["Отруби"] != "10" || exported["Стейк"] != "0" || exported["Суп в кафе"] != "" {
		t.Errorf("exported fibre = %q, want 10, 0 and an empty cell", exported)
	}

	// A later ingredient edit does not touch the logged entry, switching the entry to the ingredient copies the new value
	if _, stderr, code := env.run("edit-ingredient", fmt.Sprint(bran.ID), "--fiber", "12"); code != 0 {
		t.Fatalf("edit-ingredient: exit code %d; stderr: %s", code, stderr)
	}
	if got := optional(v1Ingredients(t, client)[bran.ID].Fiber); got != 12.0 {
		t.Errorf("ingredient after edit-ingredient: fiber = %v, want 12", got)
	}
	if got := optional(fetchEntry(t, client, day, branEntry).Fiber); got != 10.0 {
		t.Errorf("logged entry after edit-ingredient: fiber = %v, want 10", got)
	}
	if _, stderr, code := env.run("edit", fmt.Sprint(branEntry), "--ingredient-id", fmt.Sprint(bran.ID)); code != 0 {
		t.Fatalf("edit: exit code %d; stderr: %s", code, stderr)
	}
	if got := optional(fetchEntry(t, client, day, branEntry).Fiber); got != 12.0 {
		t.Errorf("entry after edit --ingredient-id: fiber = %v, want 12", got)
	}
}
