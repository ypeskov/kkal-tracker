package server

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/web"
)

// newStaticTestServer builds the server with the real embedded frontend (requires web/dist)
func newStaticTestServer(t *testing.T) http.Handler {
	t.Helper()

	cfg := &config.Config{
		Port:      "0",
		JWTSecret: strings.Repeat("s", 32),
		AppURL:    "http://localhost:8080",
	}
	return New(cfg, discardLogger(), nil, web.StaticFiles).Start().Handler
}

func get(handler http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = "203.0.113.7:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestStaticFilesAndSPAFallback(t *testing.T) {
	handler := newStaticTestServer(t)

	index := get(handler, "/")
	if index.Code != http.StatusOK || !strings.Contains(index.Body.String(), `<div id="root">`) {
		t.Fatalf("GET /: status = %d, want the frontend index page", index.Code)
	}

	// Client-side routes have no file and no server route: the index page must be served
	for _, path := range []string{"/reports", "/activate/some-token", "/no/such/page"} {
		rec := get(handler, path)
		if rec.Code != http.StatusOK || rec.Body.String() != index.Body.String() {
			t.Errorf("GET %s: status = %d, want the index page (SPA fallback)", path, rec.Code)
		}
		if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
			t.Errorf("GET %s: Content-Type = %q, want text/html", path, contentType)
		}
	}

	// Real assets referenced by the index page are served as files
	assets := regexp.MustCompile(`/assets/[A-Za-z0-9._-]+`).FindAllString(index.Body.String(), -1)
	if len(assets) == 0 {
		t.Fatal("the index page references no assets")
	}
	for _, asset := range assets {
		rec := get(handler, asset)
		if rec.Code != http.StatusOK || rec.Body.Len() == 0 || rec.Body.String() == index.Body.String() {
			t.Errorf("GET %s: status = %d, want the asset file", asset, rec.Code)
		}
	}
}

func TestAPIErrorsAreNotReplacedByTheIndexPage(t *testing.T) {
	handler := newStaticTestServer(t)

	for _, path := range []string{"/api/unknown", "/api/calories", "/api/v1/data"} {
		rec := get(handler, path)
		if rec.Code < 400 {
			t.Errorf("GET %s: status = %d, want an error status", path, rec.Code)
		}
		if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
			t.Errorf("GET %s: Content-Type = %q, want application/json", path, contentType)
		}
	}
}
