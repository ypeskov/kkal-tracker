package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
)

// recordedRequest is a request that reached the server from the skill client
type recordedRequest struct {
	method, uri, body string
}

// skillClientEnv is the downloaded skill, unpacked like an agent installs it, and the server it talks to over HTTP
type skillClientEnv struct {
	t       *testing.T
	python  string
	dir     string
	baseURL string

	mu       sync.Mutex
	requests []recordedRequest
	// respond, when set, answers the requests it returns true for instead of the server, e.g. like an older server
	respond func(w http.ResponseWriter, r *http.Request) bool
}

// newSkillClientEnv serves the smoke environment over a real listener, downloads the skill of the client's user
// and unpacks it into a temporary folder. It skips the test when python3 is not installed.
func newSkillClientEnv(t *testing.T, client *smokeClient) *skillClientEnv {
	t.Helper()

	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is not installed: the skill client cannot be tested")
	}

	env := &skillClientEnv{t: t, python: python, dir: t.TempDir()}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		env.mu.Lock()
		env.requests = append(env.requests, recordedRequest{r.Method, r.URL.RequestURI(), string(body)})
		respond := env.respond
		env.mu.Unlock()
		if respond != nil && respond(w, r) {
			return
		}
		client.handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	env.baseURL = server.URL

	_, files := downloadSkill(t, client)
	skillKey(t, files, "kkal-tracker-food-log")
	for name, content := range files {
		path := filepath.Join(env.dir, filepath.FromSlash(strings.TrimPrefix(name, "kkal-tracker-food-log/")))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// A key file of an old installation next to SKILL.md must be ignored: the key is in the client now
	if err := os.WriteFile(filepath.Join(env.dir, "api_key"), []byte("old-file-key\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return env
}

// runIn starts the client of the skill in dir with the server address and the given environment variables only
// (KKAL_API_KEY is unset unless given) and returns its output and exit code
func (e *skillClientEnv) runIn(dir string, extraEnv map[string]string, args ...string) (stdout, stderr string, code int) {
	e.t.Helper()

	cmd := exec.Command(e.python, append([]string{filepath.Join(dir, "scripts", "kkal_client.py")}, args...)...)
	cmd.Dir = dir
	for _, variable := range os.Environ() {
		if !strings.HasPrefix(variable, "KKAL_") {
			cmd.Env = append(cmd.Env, variable)
		}
	}
	cmd.Env = append(cmd.Env, "KKAL_BASE_URL="+e.baseURL, "PYTHONDONTWRITEBYTECODE=1", "PYTHONIOENCODING=utf-8")
	for name, value := range extraEnv {
		cmd.Env = append(cmd.Env, name+"="+value)
	}

	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case errors.As(err, &exitErr):
		code = exitErr.ExitCode()
	case err != nil:
		e.t.Fatalf("failed to run the skill client: %v", err)
	}
	return out.String(), errOut.String(), code
}

func (e *skillClientEnv) run(args ...string) (stdout, stderr string, code int) {
	e.t.Helper()
	return e.runIn(e.dir, nil, args...)
}

// answerAs makes the server answer the given method and path with status and a JSON message, like a server without
// the route; an empty status forwards everything again
func (e *skillClientEnv) answerAs(method, path string, status int, message string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if status == 0 {
		e.respond = nil
		return
	}
	e.respond = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method != method || r.URL.Path != path {
			return false
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"message":%q}`+"\n", message)
		return true
	}
}

// takeRequests returns the requests recorded since the last call
func (e *skillClientEnv) takeRequests() []recordedRequest {
	e.mu.Lock()
	defer e.mu.Unlock()
	requests := e.requests
	e.requests = nil
	return requests
}

type totalsRow struct {
	EntryIDs []int `json:"entry_ids"`
	Entries  int   `json:"entries"`
	Kcal     int   `json:"kcal"`
	Fats     int   `json:"fats"`
	Carbs    int   `json:"carbs"`
	Proteins int   `json:"proteins"`
}

type totalsOutput struct {
	Day            string     `json:"day"`
	Meal           *totalsRow `json:"meal"`
	DayTotal       totalsRow  `json:"day_total"`
	MissingMealIDs []int      `json:"missing_meal_ids"`
	WithoutMacros  []struct {
		ID     int    `json:"id"`
		Food   string `json:"food"`
		InMeal bool   `json:"in_meal"`
	} `json:"without_macros"`
	Check struct {
		Meal *string `json:"meal"`
		Day  *string `json:"day"`
	} `json:"check"`
}

// objectKeys returns the keys of a JSON object in the order they are written
func objectKeys(t *testing.T, raw []byte) []string {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		t.Fatalf("not a JSON object: %s", raw)
	}
	var keys []string
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			t.Fatalf("invalid JSON object: %v", err)
		}
		keys = append(keys, token.(string))
		var skip json.RawMessage
		if err := decoder.Decode(&skip); err != nil {
			t.Fatalf("invalid JSON value: %v", err)
		}
	}
	return keys
}

// totals runs the totals command, expects success with exactly one request for the day and decodes the output
func (e *skillClientEnv) totals(args ...string) (totalsOutput, string) {
	e.t.Helper()

	stdout, stderr, code := e.run(append([]string{"totals"}, args...)...)
	if code != 0 {
		e.t.Fatalf("totals %v: exit code %d; stderr: %s", args, code, stderr)
	}
	day := args[slices.Index(args, "--day")+1]
	want := []recordedRequest{{http.MethodGet, "/api/v1/data?type=food&from=" + day + "&to=" + day, ""}}
	if requests := e.takeRequests(); !slices.Equal(requests, want) {
		e.t.Errorf("totals %v: requests = %+v, want exactly %+v", args, requests, want)
	}

	var output totalsOutput
	if err := json.Unmarshal([]byte(stdout), &output); err != nil {
		e.t.Fatalf("totals %v: output is not JSON: %v; stdout: %s", args, err, stdout)
	}
	return output, stdout
}

func checkTotalsRow(t *testing.T, name string, got *totalsRow, want totalsRow) {
	t.Helper()
	if got == nil {
		t.Errorf("%s row = null, want %+v", name, want)
		return
	}
	if !slices.Equal(got.EntryIDs, want.EntryIDs) || got.Entries != want.Entries || got.Kcal != want.Kcal ||
		got.Fats != want.Fats || got.Carbs != want.Carbs || got.Proteins != want.Proteins {
		t.Errorf("%s row = %+v, want %+v", name, *got, want)
	}
}

func checkResult(t *testing.T, name string, got *string, want string) {
	t.Helper()
	switch {
	case want == "" && got != nil:
		t.Errorf("check.%s = %q, want null", name, *got)
	case want != "" && (got == nil || *got != want):
		t.Errorf("check.%s = %v, want %q", name, got, want)
	}
}

// The downloaded client runs with the key embedded by the download and adds up the macros of a meal and of a day
func TestSkillClientTotals(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-client@example.com")
	env := newSkillClientEnv(t, client)

	oneOff := func(name string, kcal float64, weight float64, nutrients map[string]float64) map[string]any {
		food := map[string]any{"name": name, "kcal_per_100g": kcal}
		for nutrient, value := range nutrients {
			food[nutrient] = value
		}
		return map[string]any{"new_ingredient": food, "weight": weight, "one_off": true}
	}
	storeMeal := func(datetime string, items ...map[string]any) foodAPIMeal {
		var meal foodAPIMeal
		client.expect(http.StatusCreated, http.MethodPost, "/api/v1/food",
			map[string]any{"meal_datetime": datetime, "items": items}, &meal)
		return meal
	}

	// The meal: carbs of A are unknown and count as 0
	meal := storeMeal("2026-03-10T12:00:00Z",
		oneOff("Свиная шея", 267, 200, map[string]float64{"fats": 22.5, "proteins": 16}),
		oneOff("Помидор", 20, 120, map[string]float64{"fats": 0.2, "carbs": 4.2, "proteins": 0.6}),
		oneOff("Сырники", 220, 150, map[string]float64{"fats": 9, "carbs": 20, "proteins": 14}))
	a, b, c := meal.Entries[0].ID, meal.Entries[1].ID, meal.Entries[2].ID
	// Another entry of the same day without any macros, and one of the next UTC day that must not count
	coffee := storeMeal("2026-03-10T18:00:00Z", oneOff("Кофе", 50, 100, nil)).Entries[0].ID
	storeMeal("2026-03-11T00:30:00Z", oneOff("Завтра", 100, 100, map[string]float64{"fats": 50}))
	// Halves: 12.5 g of fats, 0.5 g of carbs and 2.5 g of proteins
	storeMeal("2026-03-12T12:00:00Z", oneOff("Половинки", 100, 50, map[string]float64{"fats": 25, "carbs": 1, "proteins": 5}))

	ids := func(values ...int) []string {
		var args []string
		for _, value := range values {
			args = append(args, fmt.Sprint(value))
		}
		return args
	}

	t.Run("meal and day", func(t *testing.T) {
		args := append(append([]string{"--day", "2026-03-10", "--meal"}, ids(a, b, c)...),
			"--expect-meal-kcal", "888", "--expect-day-kcal", "938")
		output, stdout := env.totals(args...)

		if keys := objectKeys(t, []byte(stdout)); !slices.Equal(keys,
			[]string{"day", "meal", "day_total", "missing_meal_ids", "without_macros", "check"}) {
			t.Errorf("output keys = %q", keys)
		}
		var rows struct {
			Meal     json.RawMessage `json:"meal"`
			DayTotal json.RawMessage `json:"day_total"`
		}
		if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
			t.Fatal(err)
		}
		if keys := objectKeys(t, rows.Meal); !slices.Equal(keys, []string{"entry_ids", "kcal", "fats", "carbs", "proteins"}) {
			t.Errorf("meal keys = %q", keys)
		}
		if keys := objectKeys(t, rows.DayTotal); !slices.Equal(keys, []string{"entries", "kcal", "fats", "carbs", "proteins"}) {
			t.Errorf("day_total keys = %q", keys)
		}

		if output.Day != "2026-03-10" {
			t.Errorf("day = %q", output.Day)
		}
		// fats 45 + 0.24 + 13.5, carbs 0 + 5.04 + 30, proteins 32 + 0.72 + 21
		checkTotalsRow(t, "meal", output.Meal, totalsRow{EntryIDs: []int{a, b, c}, Kcal: 888, Fats: 59, Carbs: 35, Proteins: 54})
		checkTotalsRow(t, "day", &output.DayTotal, totalsRow{Entries: 4, Kcal: 938, Fats: 59, Carbs: 35, Proteins: 54})
		if len(output.MissingMealIDs) != 0 {
			t.Errorf("missing_meal_ids = %v, want []", output.MissingMealIDs)
		}
		if w := output.WithoutMacros; len(w) != 1 || w[0].ID != coffee || w[0].Food != "Кофе" || w[0].InMeal {
			t.Errorf("without_macros = %+v, want only the coffee, not in the meal", w)
		}
		checkResult(t, "meal", output.Check.Meal, "ok")
		checkResult(t, "day", output.Check.Day, "ok")
	})

	t.Run("mismatches and missing entries", func(t *testing.T) {
		args := append(append([]string{"--day", "2026-03-10", "--meal"}, ids(a, 999999, coffee)...),
			"--expect-meal-kcal", "584", "--expect-day-kcal", "1000")
		output, _ := env.totals(args...)

		// The meal row sums what was found, and still fails because an entry is missing
		checkTotalsRow(t, "meal", output.Meal, totalsRow{EntryIDs: []int{a, coffee}, Kcal: 584, Fats: 45, Carbs: 0, Proteins: 32})
		if !slices.Equal(output.MissingMealIDs, []int{999999}) {
			t.Errorf("missing_meal_ids = %v, want [999999]", output.MissingMealIDs)
		}
		if w := output.WithoutMacros; len(w) != 1 || w[0].ID != coffee || !w[0].InMeal {
			t.Errorf("without_macros = %+v, want the coffee in the meal", w)
		}
		checkResult(t, "meal", output.Check.Meal, "mismatch")
		checkResult(t, "day", output.Check.Day, "mismatch")
	})

	t.Run("meal kcal mismatch", func(t *testing.T) {
		output, _ := env.totals(append(append([]string{"--day", "2026-03-10", "--meal"}, ids(a, b)...),
			"--expect-meal-kcal", "888")...)
		checkResult(t, "meal", output.Check.Meal, "mismatch")
		checkResult(t, "day", output.Check.Day, "")
	})

	t.Run("day only", func(t *testing.T) {
		output, _ := env.totals("--day", "2026-03-10")
		if output.Meal != nil {
			t.Errorf("meal = %+v, want null without --meal", *output.Meal)
		}
		checkTotalsRow(t, "day", &output.DayTotal, totalsRow{Entries: 4, Kcal: 938, Fats: 59, Carbs: 35, Proteins: 54})
		checkResult(t, "meal", output.Check.Meal, "")
		checkResult(t, "day", output.Check.Day, "")
	})

	t.Run("halves round up", func(t *testing.T) {
		output, _ := env.totals("--day", "2026-03-12")
		checkTotalsRow(t, "day", &output.DayTotal, totalsRow{Entries: 1, Kcal: 50, Fats: 13, Carbs: 1, Proteins: 3})
	})

	t.Run("empty day", func(t *testing.T) {
		output, stdout := env.totals("--day", "2026-03-20")
		checkTotalsRow(t, "day", &output.DayTotal, totalsRow{})
		if output.Meal != nil || output.MissingMealIDs == nil || len(output.MissingMealIDs) != 0 ||
			output.WithoutMacros == nil || len(output.WithoutMacros) != 0 {
			t.Errorf("empty day output = %s", stdout)
		}
	})

	t.Run("invalid arguments make no request", func(t *testing.T) {
		for _, args := range [][]string{
			{"totals"},
			{"totals", "--day", "2026-13-01"},
			{"totals", "--day", "13.10.2026"},
			{"totals", "--day", "2026-3-1"},
			{"totals", "--day", "2026-03-10", "--expect-meal-kcal", "888"},
		} {
			if _, stderr, code := env.run(args...); code == 0 || stderr == "" {
				t.Errorf("%v: exit code %d, stderr %q; want a failure with a message", args, code, stderr)
			}
		}
		if _, stderr, code := env.run("totals", "--day", "2026-13-01"); code != 1 || !strings.Contains(stderr, "YYYY-MM-DD") {
			t.Errorf("invalid --day: exit code %d, stderr %q; want 1 and the expected format", code, stderr)
		}
		if requests := env.takeRequests(); len(requests) != 0 {
			t.Errorf("requests = %+v, want none", requests)
		}
	})

	t.Run("errors", func(t *testing.T) {
		// KKAL_API_KEY overrides the embedded key
		_, stderr, code := env.runIn(env.dir, map[string]string{"KKAL_API_KEY": "rejected-key"}, "totals", "--day", "2026-03-10")
		if code != 1 || !strings.Contains(stderr, "HTTP 401") {
			t.Errorf("rejected key: exit code %d, stderr %q; want 1 and HTTP 401", code, stderr)
		}
		output, _, code := env.runIn(env.dir, map[string]string{"KKAL_API_KEY": client.apiKey}, "totals", "--day", "2026-03-12")
		if code != 0 || !strings.Contains(output, `"entries": 1`) {
			t.Errorf("another valid key: exit code %d, output %q", code, output)
		}
		if requests := env.takeRequests(); len(requests) != 2 {
			t.Errorf("requests = %+v, want the two totals requests", requests)
		}

		// Nothing listens on port 1
		_, stderr, code = env.runIn(env.dir, map[string]string{"KKAL_BASE_URL": "http://127.0.0.1:1"}, "totals", "--day", "2026-03-10")
		if code != 1 || !strings.Contains(stderr, "NETWORK ERROR") {
			t.Errorf("network error: exit code %d, stderr %q; want 1 and NETWORK ERROR", code, stderr)
		}
	})

	t.Run("edit macros", func(t *testing.T) {
		stdout, stderr, code := env.run("edit", fmt.Sprint(b), "--fats", "3", "--carbs", "48", "--proteins", "9")
		if code != 0 {
			t.Fatalf("edit: exit code %d; stderr: %s", code, stderr)
		}
		requests := env.takeRequests()
		if len(requests) != 1 || requests[0].method != http.MethodPut || requests[0].uri != fmt.Sprintf("/api/v1/food/%d", b) {
			t.Fatalf("edit requests = %+v, want one PUT of the entry", requests)
		}
		var sent map[string]any
		if err := json.Unmarshal([]byte(requests[0].body), &sent); err != nil ||
			!maps.Equal(sent, map[string]any{"fats": 3.0, "carbs": 48.0, "proteins": 9.0}) {
			t.Errorf("edit sent %s, want exactly fats, carbs and proteins", requests[0].body)
		}
		var response struct {
			Entry foodAPIEntry `json:"entry"`
		}
		if err := json.Unmarshal([]byte(stdout), &response); err != nil || response.Entry.Weight != 120 ||
			optional(response.Entry.Fats) != 3.0 || optional(response.Entry.Carbs) != 48.0 || optional(response.Entry.Proteins) != 9.0 {
			t.Errorf("edit output = %s, want the entry with the new macros and the old weight", stdout)
		}

		_, stderr, code = env.run("edit", fmt.Sprint(b))
		if code != 1 || !strings.Contains(stderr, "--fats") || !strings.Contains(stderr, "--carbs") || !strings.Contains(stderr, "--proteins") {
			t.Errorf("edit without options: exit code %d, stderr %q; want 1 and the macro options listed", code, stderr)
		}
		if requests := env.takeRequests(); len(requests) != 0 {
			t.Errorf("edit without options sent %+v", requests)
		}
	})
}

// The repository template has no key: without KKAL_API_KEY the client stops before any request and tells the user
// to download the skill again
func TestSkillClientTemplateWithoutKey(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-template@example.com")
	env := newSkillClientEnv(t, client)

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(filepath.Join("..", "..", "skills", "kkal-tracker-food-log", "scripts", "kkal_client.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "kkal_client.py"), template, 0o600); err != nil {
		t.Fatal(err)
	}
	// A key file next to SKILL.md does not configure the client any more
	if err := os.WriteFile(filepath.Join(dir, "api_key"), []byte(client.apiKey+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := env.runIn(dir, nil, "ingredients")
	if code != 1 || !strings.Contains(stderr, "not configured") || !strings.Contains(stderr, "Settings -> AI") {
		t.Errorf("exit code %d, stderr %q; want 1 and the download hint", code, stderr)
	}
	if strings.Contains(strings.ToLower(stderr), "api_key") || strings.Contains(stderr, "API Keys") {
		t.Errorf("the message points to an api_key file or to Settings -> API Keys: %q", stderr)
	}
	if requests := env.takeRequests(); len(requests) != 0 {
		t.Errorf("requests = %+v, want none", requests)
	}

	// The same template works once KKAL_API_KEY is set
	if _, stderr, code := env.runIn(dir, map[string]string{"KKAL_API_KEY": client.apiKey}, "ingredients"); code != 0 {
		t.Errorf("with KKAL_API_KEY: exit code %d, stderr %q", code, stderr)
	}
}

// expectNoRequests fails when the client sent anything since the last check
func (e *skillClientEnv) expectNoRequests(what string) {
	e.t.Helper()
	if requests := e.takeRequests(); len(requests) != 0 {
		e.t.Errorf("%s: requests = %+v, want none", what, requests)
	}
}

// The client adds an ingredient to the list with one request and reports conflicts, invalid values and older servers
func TestSkillClientAddIngredient(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-ingredient@example.com")
	env := newSkillClientEnv(t, client)

	t.Run("created", func(t *testing.T) {
		stdout, stderr, code := env.run("add-ingredient", "--name", "Хлебцы ржаные", "--kcal", "310", "--fats", "0", "--proteins", "10")
		if code != 0 {
			t.Fatalf("exit code %d; stderr: %s", code, stderr)
		}
		requests := env.takeRequests()
		if len(requests) != 1 || requests[0].method != http.MethodPost || requests[0].uri != "/api/v1/ingredients" {
			t.Fatalf("requests = %+v, want one POST /api/v1/ingredients", requests)
		}
		// 0 is sent, carbs that were not passed are not
		var sent map[string]any
		if err := json.Unmarshal([]byte(requests[0].body), &sent); err != nil || !maps.Equal(sent,
			map[string]any{"name": "Хлебцы ржаные", "kcal_per_100g": 310.0, "fats": 0.0, "proteins": 10.0}) {
			t.Errorf("sent %s, want exactly name, kcal_per_100g, fats 0 and proteins", requests[0].body)
		}

		var created apiIngredient
		if err := json.Unmarshal([]byte(stdout), &created); err != nil || !created.is("Хлебцы ржаные", 310, 0.0, nil, 10.0) {
			t.Fatalf("output = %s, want the created ingredient", stdout)
		}
		if !strings.Contains(stdout, `"times_used": 0`) {
			t.Errorf("output = %s, want the list row with times_used", stdout)
		}
		if listed := v1Ingredients(t, client)[created.ID]; !listed.is("Хлебцы ржаные", 310, 0.0, nil, 10.0) {
			t.Errorf("list: %v", listed)
		}
	})

	t.Run("missing options make no request", func(t *testing.T) {
		for _, args := range [][]string{
			{"add-ingredient", "--name", "Хлебцы"},
			{"add-ingredient", "--kcal", "310"},
			{"add-ingredient", "--name", "Хлебцы", "--kcal", "много"},
		} {
			if _, stderr, code := env.run(args...); code == 0 || !strings.Contains(stderr, "usage") {
				t.Errorf("%v: exit code %d, stderr %q; want a failure with the usage", args, code, stderr)
			}
		}
		env.expectNoRequests("missing options")
	})

	t.Run("duplicate", func(t *testing.T) {
		// The existing ingredient comes from the website, so the subtest does not depend on the others
		existing := createWebIngredient(t, client,
			map[string]any{"name": "Хлебцы гречневые", "kcalPer100g": 310, "fats": 0, "proteins": 10})
		before := len(webIngredients(t, client))

		stdout, stderr, code := env.run("add-ingredient", "--name", "хлебцы  гречневые", "--kcal", "300")
		if code != 1 || !strings.Contains(stderr, "HTTP 409") {
			t.Errorf("exit code %d, stderr %q; want 1 and HTTP 409", code, stderr)
		}
		var response struct {
			Message            string        `json:"message"`
			ExistingIngredient apiIngredient `json:"existing_ingredient"`
		}
		if err := json.Unmarshal([]byte(stdout), &response); err != nil || response.ExistingIngredient.ID != existing.ID ||
			!response.ExistingIngredient.is("Хлебцы гречневые", 310, 0.0, nil, 10.0) {
			t.Errorf("output = %s, want existing_ingredient #%d", stdout, existing.ID)
		}
		if requests := env.takeRequests(); len(requests) != 1 {
			t.Errorf("requests = %+v, want one", requests)
		}
		if after := len(webIngredients(t, client)); after != before {
			t.Errorf("ingredients = %d, want %d: a duplicate was created", after, before)
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		before := len(webIngredients(t, client))

		stdout, stderr, code := env.run("add-ingredient", "--name", "Чипсы", "--kcal", "1000")
		if code != 1 || !strings.Contains(stderr, "HTTP 400") || !strings.Contains(stdout, "kcal_per_100g must be between 0 and 900") {
			t.Errorf("exit code %d, stdout %q, stderr %q; want 1, HTTP 400 and the kcal_per_100g message", code, stdout, stderr)
		}
		requests := env.takeRequests()
		if len(requests) != 1 || requests[0].method != http.MethodPost || requests[0].uri != "/api/v1/ingredients" {
			t.Errorf("requests = %+v, want one POST /api/v1/ingredients", requests)
		}
		if after := len(webIngredients(t, client)); after != before {
			t.Errorf("ingredients = %d, want %d: a rejected ingredient was created", after, before)
		}
	})

	t.Run("older server", func(t *testing.T) {
		defer env.answerAs("", "", 0, "")
		for status, message := range map[int]string{
			http.StatusMethodNotAllowed: "Method Not Allowed",
			// What a server without the route answers (checked on v5.14.0)
			http.StatusNotFound: "Not Found",
		} {
			env.answerAs(http.MethodPost, "/api/v1/ingredients", status, message)
			stdout, stderr, code := env.run("add-ingredient", "--name", "Хлебцы", "--kcal", "310")
			if code != 1 || !strings.Contains(stderr, fmt.Sprintf("HTTP %d", status)) ||
				!strings.Contains(stderr, "does not support adding ingredients yet") || !strings.Contains(stderr, "Nothing was created") {
				t.Errorf("%d: exit code %d, stderr %q; want 1 and the older-server message", status, code, stderr)
			}
			if stdout != "" || strings.Contains(stderr, "Ingredient not found") {
				t.Errorf("%d: stdout %q, stderr %q", status, stdout, stderr)
			}
			if requests := env.takeRequests(); len(requests) != 1 {
				t.Errorf("%d: requests = %+v, want one", status, requests)
			}
		}
	})

	t.Run("edit and delete keep their messages", func(t *testing.T) {
		defer env.answerAs("", "", 0, "")
		env.answerAs(http.MethodPut, "/api/v1/ingredients/999999", http.StatusNotFound, "Not Found")
		_, stderr, code := env.run("edit-ingredient", "999999", "--kcal", "1")
		if code != 1 || !strings.Contains(stderr, "does not support editing or deleting ingredients yet") ||
			!strings.Contains(stderr, "Nothing was changed") {
			t.Errorf("edit on an older server: exit code %d, stderr %q", code, stderr)
		}
		env.answerAs("", "", 0, "")

		// A missing ingredient on a current server is not an older server
		stdout, stderr, code := env.run("delete-ingredient", "999999")
		if code != 1 || !strings.Contains(stderr, "HTTP 404") || strings.Contains(stderr, "does not support") ||
			!strings.Contains(stdout, "Ingredient not found") {
			t.Errorf("delete of a missing ingredient: exit code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
		env.takeRequests()
	})
}

// templateSkillVersion is the SKILL_VERSION constant of the client in the repository
func templateSkillVersion(t *testing.T) string {
	t.Helper()
	template, err := os.ReadFile(filepath.Join("..", "..", "skills", "kkal-tracker-food-log", "scripts", "kkal_client.py"))
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile(`(?m)^SKILL_VERSION = "(\d{4}-\d{2}-\d{2})"$`).FindAllStringSubmatch(string(template), -1)
	if len(matches) != 1 {
		t.Fatalf("the client must define SKILL_VERSION as a YYYY-MM-DD date exactly once, found %d", len(matches))
	}
	return matches[0][1]
}

// version prints the version of the skill without a request, also when the key is not configured
func TestSkillClientVersion(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-version@example.com")
	env := newSkillClientEnv(t, client)
	want := templateSkillVersion(t)

	// The downloaded client
	if stdout, stderr, code := env.run("version"); code != 0 || stdout != want+"\n" {
		t.Errorf("downloaded client: exit code %d, stdout %q, stderr %q; want %s", code, stdout, stderr, want)
	}

	// The template without a key
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(filepath.Join("..", "..", "skills", "kkal-tracker-food-log", "scripts", "kkal_client.py"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "kkal_client.py"), template, 0o600); err != nil {
		t.Fatal(err)
	}
	if stdout, stderr, code := env.runIn(dir, nil, "version"); code != 0 || stdout != want+"\n" {
		t.Errorf("template: exit code %d, stdout %q, stderr %q; want %s", code, stdout, stderr, want)
	}
	env.expectNoRequests("version")
}
