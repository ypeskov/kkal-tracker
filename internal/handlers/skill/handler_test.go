package skill

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	skillservice "ypeskov/kkal-tracker/internal/services/skill"

	"github.com/labstack/echo/v5"
)

type fakeSkillService struct {
	download *skillservice.Download
	err      error
}

func (f *fakeSkillService) DownloadFoodLog(int) (*skillservice.Download, error) {
	return f.download, f.err
}

func serveDownload(service skillservice.Servicer, logs *bytes.Buffer) *httptest.ResponseRecorder {
	h := New(service, slog.New(slog.NewTextHandler(logs, nil)))

	e := echo.New()
	g := e.Group("", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("user_id", 7)
			return next(c)
		}
	})
	h.RegisterRoutes(g)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/food-log/download", nil))
	return rec
}

func TestDownloadFoodLogResponse(t *testing.T) {
	content := []byte("PK\x03\x04 zip")
	var logs bytes.Buffer
	rec := serveDownload(&fakeSkillService{
		download: &skillservice.Download{FileName: "kkal-tracker-food-log-dev.zip", Content: content},
	}, &logs)

	if rec.Code != http.StatusOK || !bytes.Equal(rec.Body.Bytes(), content) {
		t.Fatalf("response = %d %q, want 200 with the zip", rec.Code, rec.Body.String())
	}
	for header, want := range map[string]string{
		echo.HeaderContentType:        "application/zip",
		echo.HeaderContentDisposition: `attachment; filename="kkal-tracker-food-log-dev.zip"`,
		echo.HeaderCacheControl:       "no-store",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestDownloadFoodLogHidesInternalErrors(t *testing.T) {
	var logs bytes.Buffer
	rec := serveDownload(&fakeSkillService{err: errors.New("create API key: SQL logic error: database is locked (5)")}, &logs)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != `{"message":"Failed to prepare the skill"}` {
		t.Errorf("body = %s, want the generic message only", body)
	}
	if rec.Header().Get(echo.HeaderContentDisposition) != "" {
		t.Error("a failed download must not be offered as an attachment")
	}
	for _, part := range []string{"level=ERROR", "database is locked", "user_id=7"} {
		if !strings.Contains(logs.String(), part) {
			t.Errorf("log %q does not contain %q", logs.String(), part)
		}
	}
}
