package preferences

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

const user = 7

// fakePreferenceRepo keeps stored values in memory
type fakePreferenceRepo struct {
	values   map[string]string
	getErr   error
	writeErr error
	writes   int
}

func newFakeRepo() *fakePreferenceRepo {
	return &fakePreferenceRepo{values: make(map[string]string)}
}

func repoKey(userID int, key string) string {
	return fmt.Sprintf("%d/%s", userID, key)
}

func (r *fakePreferenceRepo) Get(userID int, key string) (*models.UserPreference, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	value, ok := r.values[repoKey(userID, key)]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &models.UserPreference{UserID: userID, Key: key, Value: value}, nil
}

func (r *fakePreferenceRepo) Upsert(userID int, key, value string) error {
	if r.writeErr != nil {
		return r.writeErr
	}
	r.writes++
	r.values[repoKey(userID, key)] = value
	return nil
}

func newTestService() (*Service, *fakePreferenceRepo, *bytes.Buffer) {
	repo := newFakeRepo()
	var logs bytes.Buffer
	return New(repo, slog.New(slog.NewTextHandler(&logs, nil))), repo, &logs
}

func ptr(s string) *string { return &s }

// validReports is a valid request body; fields can be overridden, and set to deleteField to remove them
func validReports(overrides map[string]any) []byte {
	body := map[string]any{
		"date_from": "2026-09-01", "date_to": "today",
		"show_weight": true, "show_calories": false, "show_goal_projection": true,
		"period": "weekly", "step_interval": 1,
	}
	for name, value := range overrides {
		if value == deleteField {
			delete(body, name)
			continue
		}
		body[name] = value
	}
	encoded, _ := json.Marshal(body)
	return encoded
}

// deleteField removes a field from validReports
const deleteField = "<delete>"

func TestSaveValidatesReports(t *testing.T) {
	tests := []struct {
		name      string
		overrides map[string]any
		// wantErr is the field the message must name; empty means the body is valid
		wantErr string
	}{
		{"valid", nil, ""},
		{"end of February", map[string]any{"date_from": "2026-02-28", "date_to": "2026-03-01"}, ""},
		{"lower bound", map[string]any{"date_from": "1900-01-01"}, ""},
		{"upper bound", map[string]any{"date_from": "2100-12-31", "date_to": "2100-12-31"}, ""},
		{"from after to is accepted", map[string]any{"date_from": "2026-09-30", "date_to": "2026-09-01"}, ""},
		{"no range", map[string]any{"date_from": nil, "date_to": nil}, ""},
		{"range fields missing count as null", map[string]any{"date_from": deleteField, "date_to": deleteField}, ""},

		{"nonexistent day", map[string]any{"date_from": "2026-02-30"}, "date_from"},
		{"nonexistent month", map[string]any{"date_from": "2026-13-01"}, "date_from"},
		{"two-digit year", map[string]any{"date_from": "26-09-01"}, "date_from"},
		{"empty date", map[string]any{"date_from": ""}, "date_from"},
		{"today as from", map[string]any{"date_from": "today"}, "date_from"},
		{"before lower bound", map[string]any{"date_from": "1899-12-31"}, "date_from"},
		{"after upper bound", map[string]any{"date_from": "2101-01-01"}, "date_from"},
		{"date with time", map[string]any{"date_from": "2026-09-01T00:00:00Z"}, "date_from"},
		{"date as number", map[string]any{"date_from": 20260901}, "date_from"},
		{"tomorrow as to", map[string]any{"date_to": "tomorrow"}, "date_to"},
		{"nonexistent to", map[string]any{"date_to": "2026-02-30"}, "date_to"},
		{"to after upper bound", map[string]any{"date_to": "2101-01-01"}, "date_to"},
		{"only from", map[string]any{"date_to": nil}, "date_from and date_to"},
		{"only to", map[string]any{"date_from": nil}, "date_from and date_to"},

		{"weight missing", map[string]any{"show_weight": deleteField}, "show_weight"},
		{"calories null", map[string]any{"show_calories": nil}, "show_calories"},
		{"projection as string", map[string]any{"show_goal_projection": "yes"}, "show_goal_projection"},
		{"weight as number", map[string]any{"show_weight": 1}, "show_weight"},

		{"period hourly", map[string]any{"period": "hourly"}, "period"},
		{"period missing", map[string]any{"period": deleteField}, "period"},
		{"period in another case", map[string]any{"period": "Daily"}, "period"},
		{"period as number", map[string]any{"period": 1}, "period"},

		{"step 7", map[string]any{"step_interval": 7}, "step_interval"},
		{"step as string", map[string]any{"step_interval": "5"}, "step_interval"},
		{"step missing", map[string]any{"step_interval": deleteField}, "step_interval"},
		{"step fraction", map[string]any{"step_interval": 5.5}, "step_interval"},
		{"step zero", map[string]any{"step_interval": 0}, "step_interval"},
	}
	for _, period := range reportPeriods {
		tests = append(tests, struct {
			name      string
			overrides map[string]any
			wantErr   string
		}{"period " + period, map[string]any{"period": period}, ""})
	}
	for _, step := range reportStepIntervals {
		tests = append(tests, struct {
			name      string
			overrides map[string]any
			wantErr   string
		}{fmt.Sprintf("step %d", step), map[string]any{"step_interval": step}, ""})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _ := newTestService()
			_, err := svc.Save(user, KeyReports, validReports(tt.overrides))

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Save() error = %v, want none", err)
				}
				return
			}
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Save() error = %v, want a validation error", err)
			}
			if !strings.HasPrefix(validationErr.Message, tt.wantErr+" ") {
				t.Errorf("message = %q, want it to name %s", validationErr.Message, tt.wantErr)
			}
			if repo.writes != 0 {
				t.Error("an invalid value was stored")
			}
		})
	}
}

func TestSaveRejectsBodiesThatAreNotObjects(t *testing.T) {
	for _, body := range []string{"", "{", "null", "[]", `"reports"`, "1", `{"period": "daily"} {}`} {
		svc, repo, _ := newTestService()
		if _, err := svc.Save(user, KeyReports, []byte(body)); !errors.Is(err, ErrInvalidBody) {
			t.Errorf("Save(%q) error = %v, want ErrInvalidBody", body, err)
		}
		if repo.writes != 0 {
			t.Errorf("Save(%q) stored a value", body)
		}
	}
}

func TestSaveStoresNormalizedValueWithVersion(t *testing.T) {
	svc, repo, _ := newTestService()

	// Unknown fields, including a forged version, and other spacing do not reach the database
	body := []byte(`{ "step_interval": 10, "period": "daily", "show_goal_projection": false, "show_calories": true,
		"show_weight": false, "date_to": "2026-08-31", "date_from": "2026-08-01", "foo": {"bar": 1}, "v": 99 }`)
	value, err := svc.Save(user, KeyReports, body)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	want := ReportsPreferences{
		DateFrom: ptr("2026-08-01"), DateTo: ptr("2026-08-31"),
		ShowWeight: false, ShowCalories: true, ShowGoalProjection: false,
		Period: "daily", StepInterval: 10,
	}
	if !reflect.DeepEqual(value, want) {
		t.Errorf("Save() = %+v, want %+v", value, want)
	}

	const wantStored = `{"date_from":"2026-08-01","date_to":"2026-08-31","period":"daily","show_calories":true,` +
		`"show_goal_projection":false,"show_weight":false,"step_interval":10,"v":1}`
	if stored := repo.values[repoKey(user, KeyReports)]; stored != wantStored {
		t.Errorf("stored value = %s, want %s", stored, wantStored)
	}

	got, err := svc.Get(user, KeyReports)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Errorf("Get() = %+v, %v, want %+v", got, err, want)
	}
}

func TestSaveReplacesPreviousValue(t *testing.T) {
	svc, _, _ := newTestService()

	if _, err := svc.Save(user, KeyReports, validReports(nil)); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}
	second := validReports(map[string]any{"date_from": nil, "date_to": nil, "period": "yearly", "show_weight": false})
	if _, err := svc.Save(user, KeyReports, second); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}

	got, _ := svc.Get(user, KeyReports)
	want := ReportsPreferences{
		ShowWeight: false, ShowCalories: false, ShowGoalProjection: true, Period: "yearly", StepInterval: 1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Get() = %+v, want %+v", got, want)
	}
}

func TestGetReturnsDefaultsWhenNothingIsStored(t *testing.T) {
	svc, _, logs := newTestService()

	got, err := svc.Get(user, KeyReports)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(got, defaultReports()) {
		t.Errorf("Get() = %+v, want the defaults", got)
	}
	encoded, _ := json.Marshal(got)
	const wantJSON = `{"date_from":null,"date_to":null,"show_weight":true,"show_calories":true,` +
		`"show_goal_projection":true,"period":"daily","step_interval":1}`
	if string(encoded) != wantJSON {
		t.Errorf("defaults JSON = %s, want %s", encoded, wantJSON)
	}
	if strings.Contains(logs.String(), "level=WARN") {
		t.Errorf("a missing value is logged as a problem: %s", logs.String())
	}
}

func TestGetFallsBackOnInvalidStoredValues(t *testing.T) {
	const secret = "s3cr3t-marker"
	valid := `"date_from":"2026-08-01","date_to":"today","show_weight":false,"show_calories":true,` +
		`"show_goal_projection":false,"period":"monthly","step_interval":5`
	kept := ReportsPreferences{
		DateFrom: ptr("2026-08-01"), DateTo: ptr("today"),
		ShowWeight: false, ShowCalories: true, ShowGoalProjection: false, Period: "monthly", StepInterval: 5,
	}

	tests := []struct {
		name   string
		stored string
		want   ReportsPreferences
		// wantFields are the field names the warning must list; nil for a value that cannot be decoded at all
		wantFields []string
	}{
		{"not JSON", "not json " + secret, defaultReports(), nil},
		{"JSON array", `["` + secret + `"]`, defaultReports(), nil},
		{"no version", "{" + valid + "}", defaultReports(), nil},
		{"other version", `{"v":2,` + valid + "}", defaultReports(), nil},
		{"invalid period", `{"v":1,` + strings.Replace(valid, `"monthly"`, `"hourly`+secret+`"`, 1) + "}",
			withField(kept, func(p *ReportsPreferences) { p.Period = "daily" }), []string{"period"}},
		{"invalid step", `{"v":1,` + strings.Replace(valid, `"step_interval":5`, `"step_interval":7`, 1) + "}",
			withField(kept, func(p *ReportsPreferences) { p.StepInterval = 1 }), []string{"step_interval"}},
		{"missing flag", `{"v":1,` + strings.Replace(valid, `"show_weight":false,`, "", 1) + "}",
			withField(kept, func(p *ReportsPreferences) { p.ShowWeight = true }), []string{"show_weight"}},
		{"invalid date resets the range",
			`{"v":1,` + strings.Replace(valid, `"2026-08-01"`, `"2026-02-30`+secret+`"`, 1) + "}",
			withField(kept, func(p *ReportsPreferences) { p.DateFrom, p.DateTo = nil, nil }), []string{"date_from"}},
		{"half a range resets the range", `{"v":1,` + strings.Replace(valid, `"today"`, "null", 1) + "}",
			withField(kept, func(p *ReportsPreferences) { p.DateFrom, p.DateTo = nil, nil }), []string{"date_range"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, logs := newTestService()
			repo.values[repoKey(user, KeyReports)] = tt.stored

			got, err := svc.Get(user, KeyReports)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %+v, want %+v", got, tt.want)
			}

			log := logs.String()
			if !strings.Contains(log, "level=WARN") || !strings.Contains(log, "user_id=7") ||
				!strings.Contains(log, "key=reports") {
				t.Errorf("no warning with the user and the key: %s", log)
			}
			for _, field := range tt.wantFields {
				if !strings.Contains(log, field) {
					t.Errorf("warning does not name %s: %s", field, log)
				}
			}
			if strings.Contains(log, secret) || strings.Contains(log, "2026-08-01") {
				t.Errorf("warning reveals the stored value: %s", log)
			}
		})
	}
}

func withField(p ReportsPreferences, change func(*ReportsPreferences)) ReportsPreferences {
	change(&p)
	return p
}

func TestUnknownKeyIsRejected(t *testing.T) {
	svc, repo, _ := newTestService()

	for _, key := range []string{"dashboard", "Reports", "", "reports "} {
		if _, err := svc.Get(user, key); !errors.Is(err, ErrUnknownKey) {
			t.Errorf("Get(%q) error = %v, want ErrUnknownKey", key, err)
		}
		if _, err := svc.Save(user, key, validReports(nil)); !errors.Is(err, ErrUnknownKey) {
			t.Errorf("Save(%q) error = %v, want ErrUnknownKey", key, err)
		}
	}
	if repo.writes != 0 {
		t.Error("a value of an unknown key was stored")
	}
}

func TestRepositoryErrorsArePassedOn(t *testing.T) {
	dbErr := errors.New("database is locked")

	svc, repo, _ := newTestService()
	repo.getErr = dbErr
	if _, err := svc.Get(user, KeyReports); !errors.Is(err, dbErr) {
		t.Errorf("Get() error = %v, want the repository error", err)
	}

	svc, repo, _ = newTestService()
	repo.writeErr = dbErr
	if _, err := svc.Save(user, KeyReports, validReports(nil)); !errors.Is(err, dbErr) {
		t.Errorf("Save() error = %v, want the repository error", err)
	}
}

func TestIsCalendarDate(t *testing.T) {
	for value, want := range map[string]bool{
		"2026-02-28": true, "2024-02-29": true, "1900-01-01": true, "2100-12-31": true,
		"2026-02-29": false, "2026-02-30": false, "2026-13-01": false, "26-09-01": false, "": false,
		"1899-12-31": false, "2101-01-01": false, "2026-9-01": false, " 2026-09-01": false, "2026-09-01 ": false,
	} {
		if got := isCalendarDate(value); got != want {
			t.Errorf("isCalendarDate(%q) = %v, want %v", value, got, want)
		}
	}
}
