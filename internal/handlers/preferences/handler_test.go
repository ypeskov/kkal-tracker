package preferences

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	preferencesservice "ypeskov/kkal-tracker/internal/services/preferences"

	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

// failingPreferencesService fails every call with an internal error
type failingPreferencesService struct {
	err   error
	calls int
}

func (f *failingPreferencesService) Get(int, string) (any, error) {
	f.calls++
	return nil, f.err
}

func (f *failingPreferencesService) Save(int, string, []byte) (any, error) {
	f.calls++
	return nil, f.err
}

// newTestEcho serves the handler as the server does, with user 7 authenticated and a body limit of limit bytes
func newTestEcho(h *Handler, limit int64) *echo.Echo {
	e := echo.New()
	g := e.Group("/preferences", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("user_id", 7)
			return next(c)
		}
	}, echomiddleware.BodyLimit(limit))
	h.RegisterRoutes(g)
	return e
}

func TestPreferencesHideInternalErrors(t *testing.T) {
	internalErr := errors.New("no such table: user_preferences (1)")

	tests := []struct {
		method, body, wantMessage string
	}{
		{http.MethodGet, "", "Failed to load preferences"},
		{http.MethodPut, `{"period": "daily"}`, "Failed to save preferences"},
	}
	for _, tt := range tests {
		var logs bytes.Buffer
		h := New(&failingPreferencesService{err: internalErr}, slog.New(slog.NewTextHandler(&logs, nil)))

		req := httptest.NewRequest(tt.method, "/preferences/reports", strings.NewReader(tt.body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		newTestEcho(h, 4<<10).ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError || !strings.Contains(rec.Body.String(), tt.wantMessage) {
			t.Errorf("%s: response = %d %s, want 500 %q", tt.method, rec.Code, rec.Body.String(), tt.wantMessage)
		}
		if strings.Contains(rec.Body.String(), "no such table") {
			t.Errorf("%s: response reveals the internal error: %s", tt.method, rec.Body.String())
		}
		log := logs.String()
		if !strings.Contains(log, "level=ERROR") || !strings.Contains(log, "user_id=7") ||
			!strings.Contains(log, "key=reports") || !strings.Contains(log, "no such table") {
			t.Errorf("%s: the failure is not logged with the user, the key and the error: %s", tt.method, log)
		}
	}
}

func TestPreferencesMapDomainErrors(t *testing.T) {
	tests := []struct {
		err         error
		wantStatus  int
		wantMessage string
	}{
		{preferencesservice.ErrUnknownKey, http.StatusNotFound, "Unknown preference"},
		{preferencesservice.ErrInvalidBody, http.StatusBadRequest, "Invalid request body"},
		{&preferencesservice.ValidationError{Message: "period must be one of daily"}, http.StatusBadRequest,
			"period must be one of daily"},
	}
	for _, tt := range tests {
		for _, method := range []string{http.MethodGet, http.MethodPut} {
			h := New(&failingPreferencesService{err: tt.err}, slog.New(slog.DiscardHandler))

			req := httptest.NewRequest(method, "/preferences/reports", strings.NewReader("{}"))
			rec := httptest.NewRecorder()
			newTestEcho(h, 4<<10).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus || !strings.Contains(rec.Body.String(), `"message":"`+tt.wantMessage) {
				t.Errorf("%s %v: response = %d %s, want %d %q",
					method, tt.err, rec.Code, rec.Body.String(), tt.wantStatus, tt.wantMessage)
			}
		}
	}
}

func TestPreferencesRejectStreamedOversizedBody(t *testing.T) {
	service := &failingPreferencesService{err: errors.New("must not be called")}
	h := New(service, slog.New(slog.DiscardHandler))

	// Not a strings/bytes reader, so the request has no Content-Length and the limit applies while reading
	body := io.MultiReader(strings.NewReader(`{"padding": "`), strings.NewReader(strings.Repeat("a", 5000)),
		strings.NewReader(`"}`))
	req := httptest.NewRequest(http.MethodPut, "/preferences/reports", body)
	if req.ContentLength > 0 {
		t.Fatalf("test request has Content-Length %d, want none", req.ContentLength)
	}
	rec := httptest.NewRecorder()
	newTestEcho(h, 4<<10).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d; body: %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if service.calls != 0 {
		t.Error("the service was called with a truncated body")
	}
}
