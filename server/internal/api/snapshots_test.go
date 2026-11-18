package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/3nd3r1/kubin/server/internal/blob"
	"github.com/3nd3r1/kubin/server/internal/snapshot"
	"github.com/3nd3r1/kubin/server/internal/snapshot/snapshottest"
	"github.com/3nd3r1/kubin/server/internal/store"
)

type testServer struct {
	*httptest.Server
	db      *sql.DB
	blobDir string
}

// newUploadServer runs the router on a temporary SQLite file and blob
// directory, wired like cmd/main.go.
func newUploadServer(t *testing.T, publicURL string, maxUpload int64) testServer {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "kubin.db")
	st, err := store.OpenSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	blobDir := filepath.Join(dir, "blobs")
	blobs, err := blob.NewFS(blobDir)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      &snapshot.Service{Store: st, Blobs: blobs, TempDir: t.TempDir()},
		Store:          st,
		Logs:           blobs,
		Ready:          st.Ping,
		PublicURL:      publicURL,
		MaxUploadBytes: maxUpload,
	}))
	t.Cleanup(srv.Close)
	raw, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	return testServer{Server: srv, db: raw, blobDir: blobDir}
}

func (s testServer) upload(t *testing.T, body []byte) (*http.Response, []byte) {
	t.Helper()
	resp, err := http.Post(s.URL+"/api/v1/snapshots", "application/gzip", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	return resp, b
}

func TestUploadFixture(t *testing.T) {
	s := newUploadServer(t, "https://kubin.example.com/", 1<<20)
	fixture := snapshottest.Fixture(t)
	resp, body := s.upload(t, fixture)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("got %d %s, want 201", resp.StatusCode, body)
	}

	var got uploadResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[a-z2-7]{24}$`).MatchString(got.ID) {
		t.Errorf("id %q isn't a 120-bit base32 ID", got.ID)
	}
	if want := "https://kubin.example.com/s/" + got.ID; got.URL != want {
		t.Errorf("url %q, want %q", got.URL, want)
	}

	kinds := map[string]int{}
	rows, err := s.db.Query(`SELECT kind, COUNT(*) FROM resources WHERE snapshot_id = $1 GROUP BY kind`, got.ID)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var k string
		var n int
		if err := rows.Scan(&k, &n); err != nil {
			t.Fatal(err)
		}
		kinds[k] = n
	}
	_ = rows.Close()
	if kinds["Namespace"] != 2 || kinds["Pod"] != 2 || len(kinds) != 2 {
		t.Errorf("indexed kinds %v, want 2 Namespace and 2 Pod", kinds)
	}

	var manifest, blobKey string
	var size int64
	err = s.db.QueryRow(`SELECT manifest, blob_key, size_bytes FROM snapshots WHERE id = $1`, got.ID).
		Scan(&manifest, &blobKey, &size)
	if err != nil {
		t.Fatal(err)
	}
	if manifest != snapshottest.Manifest {
		t.Errorf("manifest not stored as uploaded: %s", manifest)
	}
	if size != int64(len(fixture)) {
		t.Errorf("size_bytes %d, want %d", size, len(fixture))
	}
	stored, err := os.ReadFile(filepath.Join(s.blobDir, filepath.FromSlash(blobKey)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, fixture) {
		t.Error("blob differs from the uploaded tarball")
	}
}

func TestUploadURLFromRequest(t *testing.T) {
	s := newUploadServer(t, "", 1<<20)
	resp, body := s.upload(t, snapshottest.Fixture(t))
	var got uploadResponse
	_ = json.Unmarshal(body, &got)
	if resp.StatusCode != http.StatusCreated || got.URL != s.URL+"/s/"+got.ID {
		t.Fatalf("got %d url %q, want 201 and %s/s/<id>", resp.StatusCode, got.URL, s.URL)
	}
}

func TestUploadErrors(t *testing.T) {
	s := newUploadServer(t, "", 2048)
	tests := []struct {
		name   string
		body   []byte
		status int
		msg    string
	}{
		{"unknown version", snapshottest.Tarball(t, snapshottest.File{Name: "manifest.json", Body: `{"format_version":9}`}),
			http.StatusBadRequest, "unsupported format_version 9"},
		{"not an archive", []byte("plain text"), http.StatusBadRequest, "not a gzip stream"},
		{"oversized", bytes.Repeat([]byte("x"), 4096), http.StatusRequestEntityTooLarge, "larger than 2048 bytes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := s.upload(t, tt.body)
			if resp.StatusCode != tt.status {
				t.Fatalf("got %d %s, want %d", resp.StatusCode, body, tt.status)
			}
			var e errorResponse
			if err := json.Unmarshal(body, &e); err != nil || !strings.Contains(e.Error, tt.msg) {
				t.Errorf("body %s doesn't carry %q", body, tt.msg)
			}
		})
	}
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM snapshots`).Scan(&n); err != nil || n != 0 {
		t.Errorf("rejected uploads left %d snapshots (err %v)", n, err)
	}
}

func TestReadyzChecksDatabase(t *testing.T) {
	r := NewRouter(Options{
		Log:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Ready: func(context.Context) error { return io.ErrClosedPipe },
	})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("got %d, want 503", rec.Code)
	}
}
