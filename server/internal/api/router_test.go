package api

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestRouter(secure bool) http.Handler {
	return NewRouter(Options{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), SecureHeaders: secure})
}

func TestHealthRoutes(t *testing.T) {
	r := newTestRouter(false)
	for _, path := range []string{"/healthz", "/readyz"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
			t.Errorf("%s: got %d %q, want 200 ok", path, rec.Code, rec.Body.String())
		}
	}
}

func TestUnknownRouteIs404(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestRouter(false).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rec.Code)
	}
}

func TestSecureHeaders(t *testing.T) {
	for _, secure := range []bool{false, true} {
		rec := httptest.NewRecorder()
		newTestRouter(secure).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
		got := rec.Header().Get("X-Content-Type-Options") == "nosniff"
		if got != secure {
			t.Errorf("secure=%v: nosniff header present=%v", secure, got)
		}
	}
}
