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

func TestIngredientChangesHideInternalErrors(t *testing.T) {
	internalErr := errors.New("SQL logic error: database is locked (5)")

	tests := []struct {
		method, body, wantMessage string
	}{
		{http.MethodPut, `{"kcal_per_100g": 260}`, "Failed to update ingredient"},
		{http.MethodDelete, "", "Failed to delete ingredient"},
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

		req := httptest.NewRequest(tt.method, "/ingredients/42", strings.NewReader(tt.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), tt.wantMessage) {
			t.Errorf("%s: response = %d %s, want 500 %q", tt.method, rec.Code, rec.Body.String(), tt.wantMessage)
		}
		if strings.Contains(rec.Body.String(), "SQL") || strings.Contains(rec.Body.String(), "locked") {
			t.Errorf("%s: response reveals the internal error: %s", tt.method, rec.Body.String())
		}
		for _, part := range []string{"level=ERROR", "database is locked", "user_id=7", "ingredient_id=42"} {
			if !strings.Contains(logs.String(), part) {
				t.Errorf("%s: log %q does not contain %q", tt.method, logs.String(), part)
			}
		}
	}
}
