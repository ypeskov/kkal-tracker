package server

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// sentBody returns the JSON body of the only request the client sent since the last call
func (e *skillClientEnv) sentBody(method, uri string) map[string]any {
	e.t.Helper()
	requests := e.takeRequests()
	if len(requests) != 1 || requests[0].method != method || requests[0].uri != uri {
		e.t.Fatalf("requests = %+v, want one %s %s", requests, method, uri)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(requests[0].body), &sent); err != nil {
		e.t.Fatalf("request body is not JSON: %s", requests[0].body)
	}
	return sent
}

// The downloaded client sums the fibre of a meal and of a day and tells a known zero from an unknown value
func TestSkillClientFiberTotals(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-fiber@example.com")
	env := newSkillClientEnv(t, client)

	oneOff := func(name string, weight float64, values map[string]any) map[string]any {
		food := map[string]any{"name": name, "kcal_per_100g": 100}
		maps.Copy(food, values)
		return map[string]any{"new_ingredient": food, "weight": weight, "one_off": true}
	}
	storeMeal := func(datetime string, items ...map[string]any) foodAPIMeal {
		var meal foodAPIMeal
		client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food",
			map[string]any{"meal_datetime": datetime, "items": items}, &meal)
		return meal
	}

	// 1.25 g + 1.25 g of fibre = 2.5 g, which rounds up to 3; the steak has a known zero; the soup has no value,
	// but has macros; the coffee has neither
	meal := storeMeal("2026-03-10T12:00:00Z",
		oneOff("Хлеб", 50, map[string]any{"fiber": 2.5, "fats": 1}),
		oneOff("Яблоко", 50, map[string]any{"fiber": 2.5}),
		oneOff("Стейк", 200, map[string]any{"fiber": 0, "proteins": 20}),
		oneOff("Суп", 300, map[string]any{"fats": 2, "carbs": 5, "proteins": 3}))
	bread, apple, steak, soup := meal.Entries[0].ID, meal.Entries[1].ID, meal.Entries[2].ID, meal.Entries[3].ID
	later := storeMeal("2026-03-10T18:00:00Z",
		oneOff("Отруби", 30, map[string]any{"fiber": 40}),
		oneOff("Кофе", 100, nil))
	bran, coffee := later.Entries[0].ID, later.Entries[1].ID
	// A day on which every entry has a known zero
	storeMeal("2026-03-11T12:00:00Z", oneOff("Курица", 150, map[string]any{"fiber": 0, "proteins": 23}))

	t.Run("meal and day", func(t *testing.T) {
		output, stdout := env.totals("--day", "2026-03-10", "--meal",
			fmt.Sprint(bread), fmt.Sprint(apple), fmt.Sprint(steak), fmt.Sprint(soup))

		if keys := objectKeys(t, []byte(stdout)); !slices.Equal(keys,
			[]string{"day", "meal", "day_total", "missing_meal_ids", "without_macros", "without_fiber", "check"}) {
			t.Errorf("output keys = %q", keys)
		}
		// Meal: 2.5 g rounds up to 3. Day: 2.5 + 12 = 14.5 rounds up to 15.
		checkTotalsRow(t, "meal", output.Meal, totalsRow{EntryIDs: []int{bread, apple, steak, soup},
			Kcal: 600, Fats: 7, Carbs: 15, Proteins: 49, Fiber: 3})
		checkTotalsRow(t, "day", &output.DayTotal, totalsRow{Entries: 6, Kcal: 730, Fats: 7, Carbs: 15, Proteins: 49, Fiber: 15})

		// Exactly the entries without a fibre value, in the shape of without_macros; the zero of the steak is known
		want := []totalsFood{{soup, "Суп", true}, {coffee, "Кофе", false}}
		got := slices.Clone(output.WithoutFiber)
		slices.SortFunc(got, func(a, b totalsFood) int { return a.ID - b.ID })
		if !slices.Equal(got, want) {
			t.Errorf("without_fiber = %+v, want %+v", got, want)
		}
		var raw struct {
			WithoutFiber []json.RawMessage `json:"without_fiber"`
		}
		if err := json.Unmarshal([]byte(stdout), &raw); err != nil || len(raw.WithoutFiber) == 0 {
			t.Fatalf("without_fiber is not a list of objects: %s", stdout)
		}
		if keys := objectKeys(t, raw.WithoutFiber[0]); !slices.Equal(keys, []string{"id", "food", "in_meal"}) {
			t.Errorf("without_fiber item keys = %q", keys)
		}

		// without_macros keeps its meaning: fibre does not take part in it. The apple and the bran have fibre only
		// and no macros, the coffee has nothing, the soup has macros and no fibre.
		gotMacros := slices.Clone(output.WithoutMacros)
		slices.SortFunc(gotMacros, func(a, b totalsFood) int { return a.ID - b.ID })
		if wantMacros := []totalsFood{{apple, "Яблоко", true}, {bran, "Отруби", false}, {coffee, "Кофе", false}}; !slices.Equal(gotMacros, wantMacros) {
			t.Errorf("without_macros = %+v, want %+v", gotMacros, wantMacros)
		}
	})

	t.Run("known zeros only", func(t *testing.T) {
		output, _ := env.totals("--day", "2026-03-11")
		checkTotalsRow(t, "day", &output.DayTotal, totalsRow{Entries: 1, Kcal: 150, Proteins: 35, Fiber: 0})
		if output.WithoutFiber == nil || len(output.WithoutFiber) != 0 {
			t.Errorf("without_fiber = %+v, want an empty list", output.WithoutFiber)
		}
	})

	t.Run("edit fiber", func(t *testing.T) {
		stdout, stderr, code := env.run("edit", fmt.Sprint(soup), "--fiber", "4")
		if code != 0 {
			t.Fatalf("edit: exit code %d; stderr: %s", code, stderr)
		}
		if sent := env.sentBody(http.MethodPut, fmt.Sprintf("/api/v1/food/%d", soup)); !maps.Equal(sent, map[string]any{"fiber": 4.0}) {
			t.Errorf("edit sent %v, want exactly the fiber", sent)
		}
		var response struct {
			Entry foodAPIEntry `json:"entry"`
		}
		if err := json.Unmarshal([]byte(stdout), &response); err != nil || optional(response.Entry.Fiber) != 4.0 ||
			response.Entry.Weight != 300 || !response.Entry.hasValues("Суп", 100, 2.0, 5.0, 3.0, 300) {
			t.Errorf("edit output = %s, want only the fiber changed", stdout)
		}

		// 0 is a value and is sent
		if stdout, stderr, code = env.run("edit", fmt.Sprint(soup), "--fiber", "0"); code != 0 || !strings.Contains(stdout, `"fiber": 0`) {
			t.Fatalf("edit --fiber 0: exit code %d; stdout: %s; stderr: %s", code, stdout, stderr)
		}
		if sent := env.sentBody(http.MethodPut, fmt.Sprintf("/api/v1/food/%d", soup)); !maps.Equal(sent, map[string]any{"fiber": 0.0}) {
			t.Errorf("edit --fiber 0 sent %v, want the fiber 0", sent)
		}

		// The server's answer to an invalid value is shown
		stdout, stderr, code = env.run("edit", fmt.Sprint(soup), "--fiber", "150")
		if code != 1 || !strings.Contains(stderr, "HTTP 400") || !strings.Contains(stdout, "fiber must be between 0 and 100") {
			t.Errorf("edit --fiber 150: exit code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
		env.takeRequests()
	})
}

func TestSkillClientFiberOfIngredients(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-fiber@example.com")
	env := newSkillClientEnv(t, client)

	add := func(args ...string) apiIngredient {
		t.Helper()
		stdout, stderr, code := env.run(append([]string{"add-ingredient"}, args...)...)
		if code != 0 {
			t.Fatalf("add-ingredient %v: exit code %d; stderr: %s", args, code, stderr)
		}
		var created apiIngredient
		if err := json.Unmarshal([]byte(stdout), &created); err != nil {
			t.Fatalf("add-ingredient output is not JSON: %s", stdout)
		}
		return created
	}

	// With a value, with zero, and without the option
	withFiber := add("--name", "N", "--kcal", "100", "--fiber", "3")
	if sent := env.sentBody(http.MethodPost, "/api/v1/ingredients"); !maps.Equal(sent,
		map[string]any{"name": "N", "kcal_per_100g": 100.0, "fiber": 3.0}) {
		t.Errorf("add-ingredient sent %v, want name, kcal_per_100g and fiber", sent)
	}
	zero := add("--name", "Мясо", "--kcal", "200", "--fiber", "0")
	if sent := env.sentBody(http.MethodPost, "/api/v1/ingredients"); sent["fiber"] != 0.0 {
		t.Errorf("add-ingredient --fiber 0 sent %v, want the fiber 0", sent)
	}
	unknown := add("--name", "Соус", "--kcal", "120")
	if sent := env.sentBody(http.MethodPost, "/api/v1/ingredients"); len(sent) != 2 {
		t.Errorf("add-ingredient without --fiber sent %v, want name and kcal_per_100g only", sent)
	}
	if optional(withFiber.Fiber) != 3.0 || optional(zero.Fiber) != 0.0 || unknown.Fiber != nil {
		t.Errorf("created: fiber = %v, %v, %v; want 3, 0 and unknown",
			optional(withFiber.Fiber), optional(zero.Fiber), optional(unknown.Fiber))
	}

	// The table: fiber after proteins, "-" for unknown and "0" for zero
	stdout, stderr, code := env.run("ingredients")
	if code != 0 {
		t.Fatalf("ingredients: exit code %d; stderr: %s", code, stderr)
	}
	env.takeRequests()
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if lines[0] != "id\tname\tkcal_per_100g\tfats\tcarbs\tproteins\tfiber\ttimes_used\tlast_used" {
		t.Errorf("header = %q", lines[0])
	}
	rows := map[string]string{}
	for _, line := range lines[1:] {
		rows[strings.Split(line, "\t")[1]] = line
	}
	for name, want := range map[string]string{
		"N":    fmt.Sprintf("%d\tN\t100\t-\t-\t-\t3\t0\t-", withFiber.ID),
		"Мясо": fmt.Sprintf("%d\tМясо\t200\t-\t-\t-\t0\t0\t-", zero.ID),
		"Соус": fmt.Sprintf("%d\tСоус\t120\t-\t-\t-\t-\t0\t-", unknown.ID),
	} {
		if rows[name] != want {
			t.Errorf("row of %s = %q, want %q", name, rows[name], want)
		}
	}

	// edit-ingredient changes only the fibre
	path := fmt.Sprintf("/api/v1/ingredients/%d", unknown.ID)
	stdout, stderr, code = env.run("edit-ingredient", fmt.Sprint(unknown.ID), "--fiber", "2.5")
	if code != 0 {
		t.Fatalf("edit-ingredient: exit code %d; stderr: %s", code, stderr)
	}
	if sent := env.sentBody(http.MethodPut, path); !maps.Equal(sent, map[string]any{"fiber": 2.5}) {
		t.Errorf("edit-ingredient sent %v, want exactly the fiber", sent)
	}
	var response struct {
		Ingredient apiIngredient `json:"ingredient"`
	}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil || !response.Ingredient.is("Соус", 120, nil, nil, nil) ||
		optional(response.Ingredient.Fiber) != 2.5 {
		t.Errorf("edit-ingredient output = %s, want only the fiber changed to 2.5", stdout)
	}
	if _, stderr, code = env.run("edit-ingredient", fmt.Sprint(unknown.ID), "--fiber", "0"); code != 0 {
		t.Fatalf("edit-ingredient --fiber 0: exit code %d; stderr: %s", code, stderr)
	}
	if sent := env.sentBody(http.MethodPut, path); !maps.Equal(sent, map[string]any{"fiber": 0.0}) {
		t.Errorf("edit-ingredient --fiber 0 sent %v, want the fiber 0", sent)
	}
	if got := optional(webIngredients(t, client)[unknown.ID].Fiber); got != 0.0 {
		t.Errorf("stored fibre after --fiber 0 = %v, want 0", got)
	}

	// Without options nothing is sent and the message lists --fiber
	for _, command := range []string{"edit-ingredient", "edit"} {
		_, stderr, code := env.run(command, fmt.Sprint(unknown.ID))
		if code != 1 || !strings.Contains(stderr, "Nothing to change") || !strings.Contains(stderr, "--fiber") {
			t.Errorf("%s without options: exit code %d, stderr %q; want 1 and --fiber listed", command, code, stderr)
		}
	}
	env.expectNoRequests("commands without options")

	// The meal JSON is passed through as is: fiber inside new_ingredient and as an item override
	mealJSON := fmt.Sprintf(`{"meal_datetime": "2026-03-10T12:00:00Z", "items": [
		{"ingredient_id": %d, "weight": 100, "fiber": 8},
		{"new_ingredient": {"name": "Груша", "kcal_per_100g": 57, "fiber": 3.1}, "weight": 100}]}`, withFiber.ID)
	stdout, stderr, code = env.run("add", "--json", mealJSON)
	if code != 0 {
		t.Fatalf("add: exit code %d; stderr: %s", code, stderr)
	}
	var meal foodAPIMeal
	if err := json.Unmarshal([]byte(stdout), &meal); err != nil || len(meal.Entries) != 2 ||
		optional(meal.Entries[0].Fiber) != 8.0 || optional(meal.Entries[1].Fiber) != 3.1 {
		t.Errorf("add output = %s, want the entries with fiber 8 and 3.1", stdout)
	}
	items := env.sentBody(http.MethodPost, "/api/v1/food")["items"].([]any)
	if items[0].(map[string]any)["fiber"] != 8.0 || items[1].(map[string]any)["new_ingredient"].(map[string]any)["fiber"] != 3.1 {
		t.Errorf("add sent %v, want the fiber of both items", items)
	}
}
