package apidata

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ypeskov/kkal-tracker/internal/middleware"
	"ypeskov/kkal-tracker/internal/models"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	ingredientservice "ypeskov/kkal-tracker/internal/services/ingredient"

	"github.com/labstack/echo/v5"
)

// failingIngredientService fails every change with an internal error
type failingIngredientService struct {
	ingredientservice.Servicer
	err error
}

func (f *failingIngredientService) PatchIngredient(*ingredientservice.PatchIngredientRequest) (*models.UserIngredient, error) {
	return nil, f.err
}

func (f *failingIngredientService) DeleteIngredient(int, int) error {
	return f.err
}

func (f *failingIngredientService) AddIngredient(*ingredientservice.AddIngredientRequest) (*models.UserIngredientUsage, error) {
	return nil, f.err
}

func TestIngredientChangesHideInternalErrors(t *testing.T) {
	internalErr := errors.New("SQL logic error: database is locked (5)")

	tests := []struct {
		method, path, body, wantMessage string
		wantLog                         []string
	}{
		{http.MethodPut, "/ingredients/42", `{"kcal_per_100g": 260}`, "Failed to update ingredient", []string{"ingredient_id=42"}},
		{http.MethodDelete, "/ingredients/42", "", "Failed to delete ingredient", []string{"ingredient_id=42"}},
		{http.MethodPost, "/ingredients", `{"name": "Хлебцы", "kcal_per_100g": 310}`, "Failed to create ingredient", nil},
	}
	for _, tt := range tests {
		var logs bytes.Buffer
		h := New(nil, nil, &failingIngredientService{err: internalErr}, slog.New(slog.NewTextHandler(&logs, nil)))

		e := echo.New()
		e.Validator = middleware.NewValidator()
		g := e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set("user_id", 7)
				return next(c)
			}
		})
		h.RegisterRoutes(g)

		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), tt.wantMessage) {
			t.Errorf("%s: response = %d %s, want 500 %q", tt.method, rec.Code, rec.Body.String(), tt.wantMessage)
		}
		if strings.Contains(rec.Body.String(), "SQL") || strings.Contains(rec.Body.String(), "locked") {
			t.Errorf("%s: response reveals the internal error: %s", tt.method, rec.Body.String())
		}
		for _, part := range append([]string{"level=ERROR", "database is locked", "user_id=7"}, tt.wantLog...) {
			if !strings.Contains(logs.String(), part) {
				t.Errorf("%s: log %q does not contain %q", tt.method, logs.String(), part)
			}
		}
	}
}

// recordingCalorieService keeps the last meal and entry change it was asked for
type recordingCalorieService struct {
	calorieservice.Servicer

	meal   *calorieservice.CreateMealRequest
	update *calorieservice.UpdateMealEntryRequest
}

func (r *recordingCalorieService) CreateMeal(req *calorieservice.CreateMealRequest) (*calorieservice.CreateMealResult, error) {
	r.meal = req
	return &calorieservice.CreateMealResult{}, nil
}

func (r *recordingCalorieService) UpdateMealEntry(req *calorieservice.UpdateMealEntryRequest) (*calorieservice.UpdateMealEntryResult, error) {
	r.update = req
	return &calorieservice.UpdateMealEntryResult{Entry: &models.CalorieEntry{ID: req.EntryID}}, nil
}

func serveFood(h *Handler, method, path, body string) *httptest.ResponseRecorder {
	e := echo.New()
	e.Validator = middleware.NewValidator()
	g := e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("user_id", 7)
			return next(c)
		}
	})
	h.RegisterRoutes(g)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func nutrients(fats, carbs, proteins *float64) [3]any {
	values := [3]any{}
	for i, v := range []*float64{fats, carbs, proteins} {
		if v != nil {
			values[i] = *v
		}
	}
	return values
}

func TestFoodHandlersPassNutrientOverrides(t *testing.T) {
	service := &recordingCalorieService{}
	h := New(service, nil, nil, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	rec := serveFood(h, http.MethodPost, "/food", `{"items": [
		{"ingredient_id": 94, "weight": 100, "kcal_per_100g": 260, "fats": 3, "carbs": 48, "proteins": 9},
		{"ingredient_id": 94, "weight": 50, "carbs": 0},
		{"new_ingredient": {"name": "Батон", "kcal_per_100g": 260, "fats": 1}, "weight": 100}]}`)
	if rec.Code != http.StatusCreated || service.meal == nil || len(service.meal.Items) != 3 {
		t.Fatalf("POST /food = %d %s, want 201 with 3 items passed on", rec.Code, rec.Body.String())
	}
	items := service.meal.Items
	if got := nutrients(items[0].Fats, items[0].Carbs, items[0].Proteins); got != [3]any{3.0, 48.0, 9.0} {
		t.Errorf("item 0 overrides = %v", got)
	}
	if got := nutrients(items[1].Fats, items[1].Carbs, items[1].Proteins); got != [3]any{nil, 0.0, nil} {
		t.Errorf("item 1 overrides = %v, want only carbs 0", got)
	}
	if got := nutrients(items[2].Fats, items[2].Carbs, items[2].Proteins); got != [3]any{} {
		t.Errorf("item 2 overrides = %v, want none: its nutrients belong to new_ingredient", got)
	}
	if fats := items[2].NewIngredient.Fats; fats == nil || *fats != 1 {
		t.Errorf("new_ingredient fats = %v, want 1", fats)
	}

	rec = serveFood(h, http.MethodPut, "/food/12", `{"fats": 5, "carbs": null, "proteins": 0, "ingredient_id": 76}`)
	if rec.Code != http.StatusOK || service.update == nil {
		t.Fatalf("PUT /food/12 = %d %s, want 200", rec.Code, rec.Body.String())
	}
	update := service.update
	if got := nutrients(update.Fats, update.Carbs, update.Proteins); got != [3]any{5.0, nil, 0.0} ||
		update.EntryID != 12 || update.UserID != 7 || update.IngredientID == nil || *update.IngredientID != 76 {
		t.Errorf("update = %+v with nutrients %v, want entry 12 of user 7, ingredient 76, fats 5, proteins 0", update, got)
	}
}
