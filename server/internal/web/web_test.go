package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func serve(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestHandler(t *testing.T) {
	h := Handler(fstest.MapFS{
		"index.html":      {Data: []byte("<html>app</html>")},
		"logo.png":        {Data: []byte("png")},
		"assets/app-1.js": {Data: []byte("js")},
		"assets/nested/x": {Data: []byte("x")},
	})
	for _, tt := range []struct {
		path, body, cache string
		status            int
	}{
		{"/", "<html>app</html>", "no-cache", 200},
		{"/index.html", "<html>app</html>", "no-cache", 200},
		{"/s/abc", "<html>app</html>", "no-cache", 200},
		{"/s/abc/pods/default/web", "<html>app</html>", "no-cache", 200},
		{"/logo.png", "png", "", 200},
		{"/assets/app-1.js", "js", "public, max-age=31536000, immutable", 200},
		{"/assets/old-0.js", "", "", 404},
		{"/assets/nested", "", "", 404},
		{"/../../etc/passwd", "<html>app</html>", "no-cache", 200},
	} {
		rec := serve(h, http.MethodGet, tt.path)
		if rec.Code != tt.status {
			t.Errorf("%s: status %d, want %d", tt.path, rec.Code, tt.status)
			continue
		}
		if tt.body != "" && rec.Body.String() != tt.body {
			t.Errorf("%s: body %q, want %q", tt.path, rec.Body.String(), tt.body)
		}
		if got := rec.Header().Get("Cache-Control"); got != tt.cache {
			t.Errorf("%s: Cache-Control %q, want %q", tt.path, got, tt.cache)
		}
	}

	if rec := serve(h, http.MethodPost, "/"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /: got %d, want 405", rec.Code)
	}
}

func TestHandlerWithoutBuiltUI(t *testing.T) {
	h := Handler(fstest.MapFS{".gitkeep": {}})
	rec := serve(h, http.MethodGet, "/s/abc")
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "make build") {
		t.Errorf("got %d %q, want 503 with build instructions", rec.Code, rec.Body.String())
	}
}
