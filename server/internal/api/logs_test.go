package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/3nd3r1/kubin/server/internal/snapshot/snapshottest"
)

func (s testServer) getText(t *testing.T, path string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Get(s.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func TestGetLog(t *testing.T) {
	s, id := withFixture(t)
	base := "/api/v1/snapshots/" + id + "/logs/default/web-0/nginx"
	for _, tt := range []struct {
		query string
		want  string
	}{
		{"", "started\n"},
		{"?previous=false", "started\n"},
		{"?previous=true", "panic: boom\n"},
		{"?previous=1&q=PANIC", "panic: boom\n"},
		{"?q=nothing-matches", ""},
	} {
		resp, body := s.getText(t, base+tt.query)
		if resp.StatusCode != http.StatusOK || body != tt.want {
			t.Errorf("%s: got %d %q, want 200 %q", tt.query, resp.StatusCode, body, tt.want)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
			t.Errorf("%s: Content-Type %q", tt.query, ct)
		}
		if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%s: no nosniff", tt.query)
		}
	}
}

func TestGetLogErrors(t *testing.T) {
	s, id := withFixture(t)
	for _, tt := range []struct {
		path   string
		status int
		msg    string
	}{
		{"/api/v1/snapshots/nope/logs/default/web-0/nginx", http.StatusNotFound, "snapshot not found"},
		{"/api/v1/snapshots/" + id + "/logs/default/web-0/other", http.StatusNotFound, "log not found"},
		{"/api/v1/snapshots/" + id + "/logs/kube-system/coredns-1/coredns", http.StatusNotFound, "log not found"},
		{"/api/v1/snapshots/" + id + "/logs/default/web-0/nginx?previous=maybe", http.StatusBadRequest, "previous"},
	} {
		resp, body := s.getText(t, tt.path)
		if resp.StatusCode != tt.status || resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("%s: got %d %s, want JSON %d", tt.path, resp.StatusCode, resp.Header.Get("Content-Type"), tt.status)
		}
		var e struct{ Error string }
		if err := json.Unmarshal([]byte(body), &e); err != nil || !strings.Contains(e.Error, tt.msg) {
			t.Errorf("%s: body %s doesn't say %q", tt.path, body, tt.msg)
		}
	}
}

func TestGrepLines(t *testing.T) {
	long := strings.Repeat("x", maxLogLine+500) + " needle\n"
	for _, tt := range []struct {
		name, in, q, want string
	}{
		{"case-insensitive", "Error one\nok\nerror two\n", "ERROR", "Error one\nerror two\n"},
		{"last line without newline", "a\nneedle", "needle", "needle"},
		{"empty", "", "x", ""},
		// The match is past the cut, so it is not seen.
		{"overlong line is cut", long + "needle\n", "needle", "needle\n"},
		{"overlong match is cut", "needle" + long, "needle", "needle" + strings.Repeat("x", maxLogLine-6) + "\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := grepLines(&out, strings.NewReader(tt.in), []byte(tt.q)); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.want {
				t.Errorf("got %d bytes %.40q..., want %d bytes %.40q...", out.Len(), out.String(), len(tt.want), tt.want)
			}
		})
	}
}

// discard is a ResponseWriter that keeps nothing, so allocations measured
// around it are the handler's own.
type discard struct {
	h      http.Header
	status int
	n      int64
}

func (d *discard) Header() http.Header         { return d.h }
func (d *discard) WriteHeader(status int)      { d.status = status }
func (d *discard) Write(p []byte) (int, error) { d.n += int64(len(p)); return len(p), nil }

func TestGetLogStreamsLargeLogs(t *testing.T) {
	const size = 32 << 20
	line := strings.Repeat("a", 99) + "\n"
	// Built in a closure so only the stored copy outlives the upload.
	tarball := func() []byte {
		files := snapshottest.FixtureFiles()
		for i := range files {
			if files[i].Name == "logs/default/web-0/nginx.log" {
				files[i].Body = strings.Repeat(line, size/len(line))
			}
		}
		return snapshottest.Tarball(t, files...)
	}()
	s := newUploadServer(t, "", 64<<20)
	resp, body := s.upload(t, tarball)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
	var up uploadResponse
	if err := json.Unmarshal(body, &up); err != nil {
		t.Fatal(err)
	}

	h := s.Config.Handler
	for _, query := range []string{"", "?q=AAA"} {
		path := "/api/v1/snapshots/" + up.ID + "/logs/default/web-0/nginx" + query
		w := &discard{h: http.Header{}}
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		runtime.ReadMemStats(&after)

		if w.status != 0 && w.status != http.StatusOK {
			t.Fatalf("%s: status %d", query, w.status)
		}
		if want := int64(size / len(line) * len(line)); w.n != want {
			t.Errorf("%s: wrote %d bytes, want %d", query, w.n, want)
		}
		// Both copy through small reused buffers: far less than the log.
		const limit = 1 << 20
		if alloc := after.TotalAlloc - before.TotalAlloc; alloc > limit {
			t.Errorf("%s: allocated %d bytes for a %d byte log, want under %d", query, alloc, size, limit)
		}
	}
}
