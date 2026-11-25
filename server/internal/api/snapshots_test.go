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
	"time"

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
	if got.ExpiresAt.IsZero() {
		t.Error("missing expires_at in upload response")
	}
	if got.DeleteToken == "" {
		t.Error("missing delete_token in upload response")
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

func TestUploadUnpackedTooLarge(t *testing.T) {
	dir := t.TempDir()
	st, err := store.OpenSQLite(context.Background(), filepath.Join(dir, "kubin.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	blobs, err := blob.NewFS(filepath.Join(dir, "blobs"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      &snapshot.Service{Store: st, Blobs: blobs, TempDir: t.TempDir(), MaxUnpacked: 4096},
		Store:          st,
		Logs:           blobs,
		MaxUploadBytes: 1 << 20,
	}))
	t.Cleanup(srv.Close)

	// Highly compressible log: small on the wire, larger than 4096 unpacked.
	body := snapshottest.Tarball(t,
		snapshottest.File{Name: "manifest.json", Body: snapshottest.Manifest},
		snapshottest.File{Name: "logs/default/web/app.log", Body: strings.Repeat("x", 16384)},
	)
	resp, err := http.Post(srv.URL+"/api/v1/snapshots", "application/gzip", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("got %d %s, want 413", resp.StatusCode, raw)
	}
	var e errorResponse
	if err := json.Unmarshal(raw, &e); err != nil || !strings.Contains(e.Error, "4096") {
		t.Errorf("body %s doesn't name unpacked limit 4096", raw)
	}
}

type slowReader struct {
	data  []byte
	pos   int
	delay time.Duration
	sent  bool
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.pos >= len(s.data) {
		return 0, io.EOF
	}
	if !s.sent {
		// Send the first half immediately, then pause longer than the server's
		// ReadTimeout/WriteTimeout before sending the rest.
		s.sent = true
		mid := max(1, len(s.data)/2)
		n := copy(p, s.data[:mid])
		s.pos += n
		return n, nil
	}
	time.Sleep(s.delay)
	n := copy(p, s.data[s.pos:])
	s.pos += n
	return n, nil
}

func TestUploadTimeoutExtendsDeadlinesOnlyForUploads(t *testing.T) {
	dir := t.TempDir()
	st, err := store.OpenSQLite(context.Background(), filepath.Join(dir, "kubin.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	blobs, err := blob.NewFS(filepath.Join(dir, "blobs"))
	if err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      &snapshot.Service{Store: st, Blobs: blobs, TempDir: t.TempDir()},
		Store:          st,
		Logs:           blobs,
		MaxUploadBytes: 1 << 20,
		UploadTimeout:  2 * time.Second,
		Ready: func(ctx context.Context) error {
			time.Sleep(250 * time.Millisecond)
			return nil
		},
	})
	srv := httptest.NewUnstartedServer(router)
	srv.Config.ReadTimeout = 100 * time.Millisecond
	srv.Config.WriteTimeout = 100 * time.Millisecond
	srv.Start()
	t.Cleanup(srv.Close)

	// Upload takes 250ms (> 100ms ReadTimeout and WriteTimeout, < 2s UploadTimeout) and succeeds.
	fixture := snapshottest.Fixture(t)
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/snapshots", &slowReader{
		data:  fixture,
		delay: 250 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/gzip")
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("slow upload failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("slow upload got %d %s, want 201", resp.StatusCode, body)
	}

	// Another route (/readyz) keeps the server's 100ms WriteTimeout and fails when slow.
	readyResp, err := srv.Client().Get(srv.URL + "/readyz")
	if err == nil {
		b, readErr := io.ReadAll(readyResp.Body)
		_ = readyResp.Body.Close()
		if readErr == nil && readyResp.StatusCode == http.StatusOK && string(b) == "ok" {
			t.Fatal("slow /readyz unexpectedly succeeded past WriteTimeout")
		}
	}
}

func TestUploadExpiresIn(t *testing.T) {
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	dir := t.TempDir()
	st, err := store.OpenSQLite(context.Background(), filepath.Join(dir, "kubin.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	blobs, err := blob.NewFS(filepath.Join(dir, "blobs"))
	if err != nil {
		t.Fatal(err)
	}
	svc := &snapshot.Service{
		Store:      st,
		Blobs:      blobs,
		TempDir:    t.TempDir(),
		DefaultTTL: 24 * time.Hour,
		MaxTTL:     72 * time.Hour,
		Now:        func() time.Time { return now },
	}
	r := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      svc,
		Store:          st,
		Logs:           blobs,
		MaxUploadBytes: 1 << 20,
		SnapshotTTL:    24 * time.Hour,
		SnapshotMaxTTL: 72 * time.Hour,
		Now:            func() time.Time { return now },
	})

	fixture := snapshottest.Fixture(t)

	// Default TTL when expires_in is omitted.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", bytes.NewReader(fixture)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("default upload: got %d %s, want 201", rec.Code, rec.Body.String())
	}
	var defResp uploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &defResp); err != nil {
		t.Fatal(err)
	}
	if want := now.Add(24 * time.Hour); !defResp.ExpiresAt.Equal(want) {
		t.Errorf("default expires_at = %v, want %v", defResp.ExpiresAt, want)
	}
	if defResp.DeleteToken == "" {
		t.Error("default upload missing delete_token")
	}

	// Custom TTL within max.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/snapshots?expires_in=48h", bytes.NewReader(fixture)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("custom upload: got %d %s, want 201", rec.Code, rec.Body.String())
	}
	var customResp uploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &customResp); err != nil {
		t.Fatal(err)
	}
	if want := now.Add(48 * time.Hour); !customResp.ExpiresAt.Equal(want) {
		t.Errorf("custom expires_at = %v, want %v", customResp.ExpiresAt, want)
	}

	// Invalid, non-positive, and over-max expires_in rejected with 400 before reading body.
	for _, tt := range []struct {
		query string
		want  string
	}{
		{"?expires_in=nope", "invalid expires_in"},
		{"?expires_in=0s", "72h0m0s"},
		{"?expires_in=-5m", "72h0m0s"},
		{"?expires_in=73h", "72h0m0s"},
	} {
		t.Run(tt.query, func(t *testing.T) {
			unread := &failOnReadBody{}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/snapshots"+tt.query, unread))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("got %d %s, want 400", rec.Code, rec.Body.String())
			}
			if unread.readCalled {
				t.Error("body was read despite invalid expires_in")
			}
			var e errorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || !strings.Contains(e.Error, tt.want) {
				t.Errorf("body %s does not mention %q", rec.Body.String(), tt.want)
			}
		})
	}
}

func TestDeleteSnapshot(t *testing.T) {
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
	svc := &snapshot.Service{Store: st, Blobs: blobs, TempDir: t.TempDir()}
	r := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      svc,
		Store:          st,
		Logs:           blobs,
		MaxUploadBytes: 1 << 20,
	})

	// Upload a snapshot.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/snapshots", bytes.NewReader(snapshottest.Fixture(t))))
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body.String())
	}
	var up uploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &up); err != nil {
		t.Fatal(err)
	}

	del := func(id, auth string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/snapshots/"+id, nil)
		if auth != "" {
			req.Header.Set("Authorization", auth)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}

	// Missing, malformed, or wrong bearer tokens give 403 and remove nothing.
	for _, auth := range []string{"", "Basic abc", "Bearer", "Bearer ", "Bearer a b", "Bearer wrong-token"} {
		rec := del(up.ID, auth)
		if rec.Code != http.StatusForbidden {
			t.Errorf("DELETE auth=%q: got %d %s, want 403", auth, rec.Code, rec.Body.String())
		}
	}
	// Missing auth on unknown ID also gives 403 without touching storage.
	stub := &stubUploader{}
	stubRouter := NewRouter(Options{
		Log:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots: stub,
	})
	rec = httptest.NewRecorder()
	stubRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/snapshots/anything", nil))
	if rec.Code != http.StatusForbidden || stub.deleteCalls != 0 {
		t.Errorf("missing auth touched storage: code=%d deleteCalls=%d", rec.Code, stub.deleteCalls)
	}

	// Unknown ID with a bearer token gives 404.
	if rec := del("aaaaaaaaaaaaaaaaaaaaaaaa", "Bearer some-token"); rec.Code != http.StatusNotFound {
		t.Errorf("DELETE unknown: got %d %s, want 404", rec.Code, rec.Body.String())
	}

	// Snapshot is still intact after all rejected deletes.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/"+up.ID, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("snapshot damaged by rejected deletes: %d %s", rec.Code, rec.Body.String())
	}

	// Valid bearer token (case-insensitive scheme) removes rows and blobs with 204.
	rec = del(up.ID, "bearer "+up.DeleteToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE valid: got %d %s, want 204", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 response had body %q", rec.Body.String())
	}

	// Both blobs and the snapshot are gone.
	for _, key := range []string{snapshot.BlobKey(up.ID), snapshot.LogsBlobKey(up.ID)} {
		if _, err := os.Stat(filepath.Join(blobDir, filepath.FromSlash(key))); !os.IsNotExist(err) {
			t.Errorf("blob %s still exists after DELETE (err=%v)", key, err)
		}
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/"+up.ID, nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET after DELETE: got %d, want 404", rec.Code)
	}
	if rec := del(up.ID, "Bearer "+up.DeleteToken); rec.Code != http.StatusNotFound {
		t.Errorf("second DELETE: got %d, want 404", rec.Code)
	}
}
