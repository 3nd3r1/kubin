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

func TestUnknownAPIRouteIsJSON404(t *testing.T) {
	ui := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ui")) })
	r := NewRouter(Options{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), UI: ui})
	for _, path := range []string{"/api", "/api/v2/x", "/api/v1/nope", "/api/v1/snapshots/a/b/c/d/e/f/g"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound || rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("%s: got %d %s, want a JSON 404", path, rec.Code, rec.Header().Get("Content-Type"))
		}
	}
	for _, path := range []string{"/", "/s/abc", "/apiary"} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Body.String() != "ui" {
			t.Errorf("%s: got %d %q, want the UI", path, rec.Code, rec.Body.String())
		}
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
