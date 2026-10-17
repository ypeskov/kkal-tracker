package server

import (
	"encoding/json"
	"maps"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

const reportPreferencesPath = "/api/preferences/reports"

// defaultReportPreferences is what a user who has saved nothing gets
const defaultReportPreferences = `{"date_from":null,"date_to":null,"show_weight":true,"show_calories":true,` +
	`"show_goal_projection":true,"period":"daily","step_interval":1}`

// reportPreferences is a valid PUT body of the reports key
func reportPreferences() map[string]any {
	return map[string]any{
		"date_from": "2026-09-01", "date_to": "today",
		"show_weight": true, "show_calories": false, "show_goal_projection": true,
		"period": "weekly", "step_interval": 1,
	}
}

// with returns a copy of the body with the fields changed; a field set to omitField is removed
func with(body map[string]any, changes map[string]any) map[string]any {
	result := maps.Clone(body)
	for name, value := range changes {
		if value == omitField {
			delete(result, name)
			continue
		}
		result[name] = value
	}
	return result
}

const omitField = "<omit>"

// getPreferences returns the raw JSON body of GET /api/preferences/reports
func getPreferences(t *testing.T, client *smokeClient) string {
	t.Helper()
	return strings.TrimSpace(client.expect(http.StatusOK, http.MethodGet, reportPreferencesPath, nil, nil).Body.String())
}

// assertJSON compares two JSON documents by value
func assertJSON(t *testing.T, label, got, want string) {
	t.Helper()
	var gotValue, wantValue any
	if err := json.Unmarshal([]byte(got), &gotValue); err != nil {
		t.Fatalf("%s: not JSON: %s", label, got)
	}
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatalf("%s: invalid expected JSON: %s", label, want)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Errorf("%s = %s, want %s", label, got, want)
	}
}

// currentUserID returns the ID of the user the client is logged in as
func currentUserID(t *testing.T, client *smokeClient) int {
	t.Helper()
	var me struct {
		ID int `json:"id"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/auth/me", nil, &me)
	return me.ID
}

func TestPreferencesSaveAndLoad(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "prefs@example.com")

	// Nothing saved yet: the defaults
	assertJSON(t, "GET of a new user", getPreferences(t, client), defaultReportPreferences)

	// PUT answers with the stored object, GET returns exactly that object
	saved := client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, reportPreferences(), nil)
	assertJSON(t, "PUT response", saved.Body.String(), `{"date_from":"2026-09-01","date_to":"today","show_weight":true,`+
		`"show_calories":false,"show_goal_projection":true,"period":"weekly","step_interval":1}`)
	if got := getPreferences(t, client); got != strings.TrimSpace(saved.Body.String()) {
		t.Errorf("GET after PUT = %s, want the PUT response %s", got, saved.Body.String())
	}

	// The database holds the server's own encoding with the schema version, not the request body
	var stored string
	if err := client.db.QueryRow(`SELECT value FROM user_preferences WHERE user_id = ? AND pref_key = 'reports'`,
		currentUserID(t, client)).Scan(&stored); err != nil {
		t.Fatalf("failed to read the stored value: %v", err)
	}
	const wantStored = `{"date_from":"2026-09-01","date_to":"today","period":"weekly","show_calories":false,` +
		`"show_goal_projection":true,"show_weight":true,"step_interval":1,"v":1}`
	if stored != wantStored {
		t.Errorf("stored value = %s, want %s", stored, wantStored)
	}

	// A second PUT replaces the whole value; unknown fields are ignored and not returned
	second := map[string]any{
		"date_from": "2026-08-01", "date_to": "2026-08-31",
		"show_weight": false, "show_calories": true, "show_goal_projection": false,
		"period": "daily", "step_interval": 10, "foo": 1, "v": 99,
	}
	const wantSecond = `{"date_from":"2026-08-01","date_to":"2026-08-31","show_weight":false,"show_calories":true,` +
		`"show_goal_projection":false,"period":"daily","step_interval":10}`
	replaced := client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, second, nil)
	assertJSON(t, "second PUT response", replaced.Body.String(), wantSecond)
	assertJSON(t, "GET after the second PUT", getPreferences(t, client), wantSecond)

	// No range: the dates stay null, so the client keeps its rolling default
	noRange := with(reportPreferences(), map[string]any{"date_from": nil, "date_to": nil})
	client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, noRange, nil)
	assertJSON(t, "GET after a PUT without a range", getPreferences(t, client),
		`{"date_from":null,"date_to":null,"show_weight":true,"show_calories":false,"show_goal_projection":true,`+
			`"period":"weekly","step_interval":1}`)
}

func TestPreferencesAcceptBoundaries(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "prefs-bounds@example.com")

	for _, dates := range [][2]string{
		{"1900-01-01", "2100-12-31"},
		{"2026-02-28", "today"},
		{"2026-09-30", "2026-09-01"}, // From after To is accepted, as the date inputs allow it
	} {
		body := with(reportPreferences(), map[string]any{"date_from": dates[0], "date_to": dates[1]})
		var saved struct {
			DateFrom string `json:"date_from"`
			DateTo   string `json:"date_to"`
		}
		client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, body, &saved)
		if saved.DateFrom != dates[0] || saved.DateTo != dates[1] {
			t.Errorf("saved range = %+v, want %v", saved, dates)
		}
	}

	for _, period := range []string{"daily", "weekly", "monthly", "yearly"} {
		client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath,
			with(reportPreferences(), map[string]any{"period": period}), nil)
	}
	for _, step := range []int{1, 5, 10, 15} {
		client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath,
			with(reportPreferences(), map[string]any{"step_interval": step}), nil)
	}
}

func TestPreferencesRejectInvalidValues(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "prefs-invalid@example.com")

	client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, reportPreferences(), nil)
	before := getPreferences(t, client)

	tests := []struct {
		changes map[string]any
		// wantField is the start of the message: the field it names
		wantField string
	}{
		{map[string]any{"date_from": "2026-02-30"}, "date_from"},
		{map[string]any{"date_from": "2026-13-01"}, "date_from"},
		{map[string]any{"date_from": "26-09-01"}, "date_from"},
		{map[string]any{"date_from": ""}, "date_from"},
		{map[string]any{"date_from": "today"}, "date_from"},
		{map[string]any{"date_from": "1899-12-31"}, "date_from"},
		{map[string]any{"date_from": "2101-01-01"}, "date_from"},
		{map[string]any{"date_to": "tomorrow"}, "date_to"},
		{map[string]any{"date_to": "2026-02-30"}, "date_to"},
		{map[string]any{"date_to": nil}, "date_from and date_to"},
		{map[string]any{"date_from": nil}, "date_from and date_to"},
		{map[string]any{"period": "hourly"}, "period"},
		{map[string]any{"period": omitField}, "period"},
		{map[string]any{"step_interval": 7}, "step_interval"},
		{map[string]any{"step_interval": "5"}, "step_interval"},
		{map[string]any{"step_interval": omitField}, "step_interval"},
		{map[string]any{"show_weight": omitField}, "show_weight"},
		{map[string]any{"show_calories": "yes"}, "show_calories"},
		{map[string]any{"show_goal_projection": nil}, "show_goal_projection"},
	}
	for _, tt := range tests {
		var response struct {
			Message string `json:"message"`
		}
		client.expect(http.StatusBadRequest, http.MethodPut, reportPreferencesPath,
			with(reportPreferences(), tt.changes), &response)
		if !strings.HasPrefix(response.Message, tt.wantField+" ") {
			t.Errorf("%v: message = %q, want it to name %s", tt.changes, response.Message, tt.wantField)
		}
	}

	// Malformed JSON and bodies that are not an object
	for _, body := range []rawBody{`{"period": "daily"`, "", "null", "[]", `"reports"`} {
		var response struct {
			Message string `json:"message"`
		}
		client.expect(http.StatusBadRequest, http.MethodPut, reportPreferencesPath, body, &response)
		if response.Message != "Invalid request body" {
			t.Errorf("%q: message = %q, want \"Invalid request body\"", body, response.Message)
		}
	}

	if after := getPreferences(t, client); after != before {
		t.Errorf("rejected requests changed the stored value: %s, want %s", after, before)
	}
}

func TestPreferencesBodyLimit(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "prefs-limit@example.com")

	client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, reportPreferences(), nil)
	before := getPreferences(t, client)

	// bodyOfSize pads a valid body with an unknown field to exactly size bytes
	bodyOfSize := func(size int, changes map[string]any) rawBody {
		body := with(reportPreferences(), changes)
		body["padding"] = ""
		encoded, _ := json.Marshal(body)
		body["padding"] = strings.Repeat("a", size-len(encoded))
		encoded, _ = json.Marshal(body)
		if len(encoded) != size {
			t.Fatalf("padded body has %d bytes, want %d", len(encoded), size)
		}
		return rawBody(encoded)
	}

	// Exactly 4 KB passes, one byte more is rejected and changes nothing
	client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, bodyOfSize(4096, nil), nil)
	client.expect(http.StatusRequestEntityTooLarge, http.MethodPut, reportPreferencesPath,
		bodyOfSize(4097, map[string]any{"period": "yearly"}), nil)

	if after := getPreferences(t, client); after != before {
		t.Errorf("an oversized request changed the stored value: %s, want %s", after, before)
	}
}

func TestPreferencesAccess(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "prefs-access@example.com")
	token, apiKey := client.token, client.apiKey

	// Only a JWT authorizes: neither no credentials nor an API key
	client.token, client.apiKey = "", ""
	client.expect(http.StatusUnauthorized, http.MethodGet, reportPreferencesPath, nil, nil)
	client.expect(http.StatusUnauthorized, http.MethodPut, reportPreferencesPath, reportPreferences(), nil)
	client.apiKey = apiKey
	client.expect(http.StatusUnauthorized, http.MethodGet, reportPreferencesPath, nil, nil)
	client.expect(http.StatusUnauthorized, http.MethodPut, reportPreferencesPath, reportPreferences(), nil)

	// Only allowed keys
	client.token, client.apiKey = token, ""
	for _, key := range []string{"dashboard", "Reports", "users"} {
		for _, method := range []string{http.MethodGet, http.MethodPut} {
			var response struct {
				Message string `json:"message"`
			}
			client.expect(http.StatusNotFound, method, "/api/preferences/"+key, reportPreferences(), &response)
			if response.Message != "Unknown preference" {
				t.Errorf("%s %s: message = %q, want \"Unknown preference\"", method, key, response.Message)
			}
		}
	}
	assertJSON(t, "GET after the rejected requests", getPreferences(t, client), defaultReportPreferences)
}

func TestPreferencesIsolateUsers(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "prefs-first@example.com")
	firstToken := client.token

	client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, reportPreferences(), nil)
	firstValue := getPreferences(t, client)

	signUp(t, client, smtpServer, "prefs-second@example.com")
	assertJSON(t, "GET of the second user", getPreferences(t, client), defaultReportPreferences)
	client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath,
		with(reportPreferences(), map[string]any{"period": "yearly", "date_from": nil, "date_to": nil}), nil)

	client.token = firstToken
	if got := getPreferences(t, client); got != firstValue {
		t.Errorf("the second user's PUT changed the first user's value: %s, want %s", got, firstValue)
	}
}

func TestPreferencesFallBackOnCorruptStoredValues(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "prefs-corrupt@example.com")
	userID := currentUserID(t, client)

	store := func(value string) {
		t.Helper()
		if _, err := client.db.Exec(`INSERT INTO user_preferences (user_id, pref_key, value) VALUES (?, 'reports', ?)
			ON CONFLICT (user_id, pref_key) DO UPDATE SET value = excluded.value`, userID, value); err != nil {
			t.Fatalf("failed to store a value: %v", err)
		}
	}

	store("{not json")
	assertJSON(t, "GET of an undecodable value", getPreferences(t, client), defaultReportPreferences)

	// One invalid field gets its default, the others are kept
	store(`{"v":1,"date_from":"2026-08-01","date_to":"today","show_weight":false,"show_calories":true,` +
		`"show_goal_projection":false,"period":"hourly","step_interval":5}`)
	assertJSON(t, "GET of a value with an invalid period", getPreferences(t, client),
		`{"date_from":"2026-08-01","date_to":"today","show_weight":false,"show_calories":true,`+
			`"show_goal_projection":false,"period":"daily","step_interval":5}`)

	// An invalid range is reset as a whole
	store(`{"v":1,"date_from":"2026-02-30","date_to":"2026-03-31","show_weight":false,"show_calories":true,` +
		`"show_goal_projection":false,"period":"monthly","step_interval":5}`)
	assertJSON(t, "GET of a value with an invalid range", getPreferences(t, client),
		`{"date_from":null,"date_to":null,"show_weight":false,"show_calories":true,`+
			`"show_goal_projection":false,"period":"monthly","step_interval":5}`)

	// A PUT overwrites a corrupt value
	client.expect(http.StatusOK, http.MethodPut, reportPreferencesPath, reportPreferences(), nil)
	assertJSON(t, "GET after a PUT over a corrupt value", getPreferences(t, client),
		`{"date_from":"2026-09-01","date_to":"today","show_weight":true,"show_calories":false,`+
			`"show_goal_projection":true,"period":"weekly","step_interval":1}`)
}
