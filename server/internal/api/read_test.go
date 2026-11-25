package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/server/internal/blob"
	"github.com/3nd3r1/kubin/server/internal/snapshot"
	"github.com/3nd3r1/kubin/server/internal/snapshot/snapshottest"
	"github.com/3nd3r1/kubin/server/internal/store"
)

// withFixture uploads snapshottest.Fixture and returns the server and the ID.
func withFixture(t *testing.T) (testServer, string) {
	t.Helper()
	s := newUploadServer(t, "", 1<<20)
	resp, body := s.upload(t, snapshottest.Fixture(t))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
	var up uploadResponse
	if err := json.Unmarshal(body, &up); err != nil {
		t.Fatal(err)
	}
	return s, up.ID
}

func (s testServer) get(t *testing.T, path string) (int, []byte) {
	t.Helper()
	resp, err := http.Get(s.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(resp.Body)
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("GET %s: Content-Type %q", path, ct)
	}
	return resp.StatusCode, b
}

func TestGetSnapshot(t *testing.T) {
	s, id := withFixture(t)
	status, body := s.get(t, "/api/v1/snapshots/"+id)
	if status != http.StatusOK {
		t.Fatalf("got %d %s", status, body)
	}
	var got snapshotResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != id || time.Since(got.UploadedAt) > time.Minute {
		t.Errorf("id %q uploaded_at %v", got.ID, got.UploadedAt)
	}
	if got.ExpiresAt == nil || got.ExpiresAt.IsZero() {
		t.Errorf("expected expires_at in GET /api/v1/snapshots/{id}, got %s", body)
	}
	var m struct {
		Cluster struct{ Context string } `json:"cluster"`
	}
	if err := json.Unmarshal(got.Manifest, &m); err != nil || m.Cluster.Context != "kind-kubin" {
		t.Errorf("manifest %s doesn't carry the cluster context", got.Manifest)
	}
	want := []kindCount{{"core", "v1", "Namespace", 2}, {"core", "v1", "Pod", 2}}
	if len(got.Counts) != len(want) || got.Counts[0] != want[0] || got.Counts[1] != want[1] {
		t.Errorf("counts %+v, want %+v", got.Counts, want)
	}
}

func TestListResources(t *testing.T) {
	s, id := withFixture(t)
	for _, tt := range []struct {
		query string
		want  string
	}{
		{"", "Namespace:/default Namespace:/kube-system Pod:default/web-0 Pod:kube-system/coredns-1"},
		{"?kind=Pod", "Pod:default/web-0 Pod:kube-system/coredns-1"},
		{"?kind=Pod&namespace=kube-system", "Pod:kube-system/coredns-1"},
		{"?namespace=_cluster", "Namespace:/default Namespace:/kube-system"},
		{"?kind=Deployment", ""},
	} {
		status, body := s.get(t, "/api/v1/snapshots/"+id+"/resources"+tt.query)
		if status != http.StatusOK {
			t.Fatalf("%s: got %d %s", tt.query, status, body)
		}
		var got resourceList
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
		if got.Items == nil {
			t.Errorf("%s: items is null, want an array", tt.query)
		}
		var names []string
		for _, it := range got.Items {
			names = append(names, it.Kind+":"+it.Namespace+"/"+it.Name)
			if !json.Valid(it.Object) || it.Version != "v1" || it.Group != "core" {
				t.Errorf("%s: bad item %+v", tt.query, it)
			}
		}
		if strings.Join(names, " ") != tt.want {
			t.Errorf("%s: got %v, want %s", tt.query, names, tt.want)
		}
	}
}

func TestGetResource(t *testing.T) {
	s, id := withFixture(t)
	status, body := s.get(t, "/api/v1/snapshots/"+id+"/resources/Pod/default/web-0")
	if status != http.StatusOK {
		t.Fatalf("got %d %s", status, body)
	}
	var pod struct {
		Kind     string
		Metadata struct{ Name, Namespace string }
		Status   struct{ Phase string }
	}
	if err := json.Unmarshal(body, &pod); err != nil {
		t.Fatal(err)
	}
	if pod.Kind != "Pod" || pod.Metadata.Name != "web-0" || pod.Status.Phase != "Running" {
		t.Errorf("got %s", body)
	}

	status, body = s.get(t, "/api/v1/snapshots/"+id+"/resources/Namespace/_cluster/kube-system")
	if status != http.StatusOK || !strings.Contains(string(body), `"kube-system"`) {
		t.Errorf("cluster-scoped: got %d %s", status, body)
	}
}

func TestReadNotFound(t *testing.T) {
	s, id := withFixture(t)
	for _, tt := range []struct{ path, msg string }{
		{"/api/v1/snapshots/aaaaaaaaaaaaaaaaaaaaaaaa", "snapshot not found"},
		{"/api/v1/snapshots/aaaaaaaaaaaaaaaaaaaaaaaa/resources", "snapshot not found"},
		{"/api/v1/snapshots/aaaaaaaaaaaaaaaaaaaaaaaa/resources/Pod/default/web-0", "snapshot not found"},
		{"/api/v1/snapshots/" + id + "/resources/Pod/default/missing", "resource not found"},
		{"/api/v1/snapshots/" + id + "/resources/Pod/_cluster/web-0", "resource not found"},
	} {
		status, body := s.get(t, tt.path)
		if status != http.StatusNotFound || !strings.Contains(string(body), tt.msg) {
			t.Errorf("GET %s: got %d %s, want 404 %q", tt.path, status, body, tt.msg)
		}
	}
}

func TestExpiredSnapshotReadRoutesReturn410Then404(t *testing.T) {
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
		Store:   st,
		Blobs:   blobs,
		TempDir: t.TempDir(),
		Now:     func() time.Time { return now },
	}
	r := NewRouter(Options{
		Log:            slog.New(slog.NewTextHandler(io.Discard, nil)),
		Snapshots:      svc,
		Store:          st,
		Logs:           blobs,
		MaxUploadBytes: 1 << 20,
		Now:            func() time.Time { return now },
	})

	// Upload a snapshot expiring in 1h.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/snapshots?expires_in=1h", bytes.NewReader(snapshottest.Fixture(t))))
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body.String())
	}
	var up uploadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &up); err != nil {
		t.Fatal(err)
	}

	// Insert a legacy snapshot with NULL expires_at.
	tx, err := st.BeginSnapshot(context.Background(), store.Snapshot{
		ID: "legacy000000000000000000", UploadedAt: now, CapturedAt: now, FormatVersion: 1,
		Manifest: []byte(snapshottest.Manifest), BlobKey: snapshot.BlobKey("legacy000000000000000000"), SizeBytes: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	// GET on legacy snapshot omits expires_at.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/legacy000000000000000000", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("legacy GET: %d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"expires_at"`) {
		t.Errorf("legacy snapshot unexpectedly included expires_at: %s", rec.Body.String())
	}

	// GET on new snapshot carries the exact expires_at from upload.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/"+up.ID, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET before expiry: %d %s", rec.Code, rec.Body.String())
	}
	var meta snapshotResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.ExpiresAt == nil || !meta.ExpiresAt.Equal(up.ExpiresAt) {
		t.Errorf("GET expires_at = %v, want %v", meta.ExpiresAt, up.ExpiresAt)
	}

	paths := []string{
		"/api/v1/snapshots/" + up.ID,
		"/api/v1/snapshots/" + up.ID + "/resources",
		"/api/v1/snapshots/" + up.ID + "/resources/Pod/default/web-0",
		"/api/v1/snapshots/" + up.ID + "/logs/default/web-0/nginx",
	}

	// Advance clock to exact expiry boundary: every read route returns 410 JSON error.
	now = now.Add(time.Hour)
	for _, p := range paths {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusGone {
			t.Errorf("GET %s after expiry: got %d %s, want 410", p, rec.Code, rec.Body.String())
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("GET %s after expiry: Content-Type = %q, want application/json", p, ct)
		}
		var e errorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil || e.Error != "snapshot expired" {
			t.Errorf("GET %s after expiry: body = %s, want snapshot expired", p, rec.Body.String())
		}
	}

	// Legacy snapshot still answers 200 after advancing time.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/snapshots/legacy000000000000000000", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("legacy GET after time advance: got %d, want 200", rec.Code)
	}

	// After cleanup removes the expired snapshot, every read route returns 404.
	deleted, err := svc.DeleteExpired(context.Background())
	if err != nil || deleted != 1 {
		t.Fatalf("DeleteExpired = (%d, %v), want (1, nil)", deleted, err)
	}
	for _, p := range paths {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s after cleanup: got %d %s, want 404", p, rec.Code, rec.Body.String())
		}
	}
}
