package ingredients

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

func (f *failingIngredientService) CreateIngredient(*ingredientservice.CreateIngredientRequest) (*models.UserIngredient, error) {
	return nil, f.err
}

func (f *failingIngredientService) UpdateIngredient(*ingredientservice.UpdateIngredientRequest) (*models.UserIngredient, error) {
	return nil, f.err
}

func TestIngredientChangesHideInternalErrors(t *testing.T) {
	internalErr := errors.New("constraint failed: UNIQUE constraint failed: user_ingredients.user_id, user_ingredients.name (2067)")

	tests := []struct {
		method, path string
	}{
		{http.MethodPost, "/ingredients"},
		{http.MethodPut, "/ingredients/42"},
	}
	for _, tt := range tests {
		var logs bytes.Buffer
		h := NewHandler(&failingIngredientService{err: internalErr}, slog.New(slog.NewTextHandler(&logs, nil)))

		e := echo.New()
		e.Validator = middleware.NewValidator()
		g := e.Group("/ingredients", func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set("user_id", 7)
				return next(c)
			}
		})
		h.RegisterRoutes(g)

		req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"name": "Гречка", "kcalPer100g": 330}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), "Internal server error") {
			t.Errorf("%s: response = %d %s, want a generic 500", tt.method, rec.Code, rec.Body.String())
		}
		if strings.Contains(strings.ToLower(rec.Body.String()), "constraint") {
			t.Errorf("%s: response reveals the internal error: %s", tt.method, rec.Body.String())
		}
		if !strings.Contains(logs.String(), "level=ERROR") || !strings.Contains(logs.String(), "user_id=7") {
			t.Errorf("%s: the failure is not logged with the user: %s", tt.method, logs.String())
		}
	}
}
