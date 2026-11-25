package snapshot

import (
	"archive/tar"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/3nd3r1/kubin/server/internal/blob"
	"github.com/3nd3r1/kubin/server/internal/snapshot/snapshottest"
	"github.com/3nd3r1/kubin/server/internal/store"
)

type F = snapshottest.File

// env is a Service on a temp SQLite file and blob directory.
type env struct {
	svc     *Service
	raw     *sql.DB // a second handle on the same file, for checking rows
	blobDir string
}

func newService(t *testing.T) env {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "kubin.db")
	db, err := store.OpenSQLite(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	raw, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	blobDir := filepath.Join(dir, "blobs")
	blobs, err := blob.NewFS(blobDir)
	if err != nil {
		t.Fatal(err)
	}
	return env{svc: &Service{Store: db, Blobs: blobs, TempDir: t.TempDir()}, raw: raw, blobDir: blobDir}
}

func (e env) count(t *testing.T, table string) int {
	t.Helper()
	var n int
	if err := e.raw.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (e env) assertEmpty(t *testing.T) {
	t.Helper()
	if n := e.count(t, "snapshots"); n != 0 {
		t.Errorf("rejected upload left %d snapshot rows", n)
	}
	if n := e.count(t, "resources"); n != 0 {
		t.Errorf("rejected upload left %d resource rows", n)
	}
	if n := e.count(t, "logs"); n != 0 {
		t.Errorf("rejected upload left %d log rows", n)
	}
	files, _ := filepath.Glob(filepath.Join(e.blobDir, "snapshots", "*"))
	if len(files) != 0 {
		t.Errorf("rejected upload left blobs: %v", files)
	}
}

func TestNewID(t *testing.T) {
	seen := map[string]bool{}
	re := regexp.MustCompile(`^[a-z2-7]{24}$`)
	for range 1000 {
		id := NewID()
		if !re.MatchString(id) {
			t.Fatalf("ID %q isn't 24 lowercase base32 characters", id)
		}
		if seen[id] {
			t.Fatalf("duplicate ID %q", id)
		}
		seen[id] = true
	}
}

func TestUploadRejects(t *testing.T) {
	manifest := F{Name: "manifest.json", Body: snapshottest.Manifest}
	pod := F{Name: "resources/core/v1/Pod/default/web.json", Body: `{"kind":"Pod"}`}
	tests := []struct {
		name string
		body []byte
		want string
	}{
		{"not gzip", []byte("hello"), "not a gzip stream"},
		{"empty archive", snapshottest.Tarball(t), "empty archive"},
		{"manifest not first", snapshottest.Tarball(t, pod, manifest), "first entry"},
		{"bad manifest", snapshottest.Tarball(t, F{Name: "manifest.json", Body: "{"}), "manifest.json"},
		{"unknown version", snapshottest.Tarball(t, F{Name: "manifest.json", Body: `{"format_version":4}`}), "unsupported format_version 4"},
		{"missing version", snapshottest.Tarball(t, F{Name: "manifest.json", Body: `{}`}), "unsupported format_version 0"},
		{"dot dot", snapshottest.Tarball(t, manifest, F{Name: "resources/core/v1/Pod/../web.json", Body: "{}"}), "unsafe path segment"},
		{"absolute", snapshottest.Tarball(t, manifest, F{Name: "/etc/passwd", Body: "x"}), "empty path segment"},
		{"outside layout", snapshottest.Tarball(t, manifest, F{Name: "extra.txt", Body: "x"}), "not part of the snapshot layout"},
		{"short resource path", snapshottest.Tarball(t, manifest, F{Name: "resources/core/v1/Pod/web.json", Body: "{}"}), "not part of the snapshot layout"},
		{"not json suffix", snapshottest.Tarball(t, manifest, F{Name: "resources/core/v1/Pod/default/web.yaml", Body: "{}"}), "<name>.json"},
		{"bad json", snapshottest.Tarball(t, manifest, F{Name: pod.Name, Body: "{"}), "web.json"},
		{"json not object", snapshottest.Tarball(t, manifest, F{Name: pod.Name, Body: "[]"}), "not a JSON object"},
		{"duplicate", snapshottest.Tarball(t, manifest, pod, pod), "appears twice"},
		{"symlink", snapshottest.Tarball(t, manifest, F{Name: "logs/a/b/c.log", Type: tar.TypeSymlink}), "only regular files"},
		{"directory", snapshottest.Tarball(t, manifest, F{Name: "resources/", Type: tar.TypeDir}), "only regular files"},
		{"truncated", snapshottest.Fixture(t)[:200], ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newService(t)
			_, err := e.svc.Upload(context.Background(), bytes.NewReader(tt.body), 0)
			var inv *InvalidError
			if !errors.As(err, &inv) {
				t.Fatalf("got %v, want an InvalidError", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q doesn't mention %q", err, tt.want)
			}
			e.assertEmpty(t)
		})
	}
}

func TestUploadUnpackedLimit(t *testing.T) {
	e := newService(t)
	e.svc.MaxUnpacked = 4096
	big := F{Name: "logs/default/web/app.log", Body: strings.Repeat("x", 8192)}
	body := snapshottest.Tarball(t, F{Name: "manifest.json", Body: snapshottest.Manifest}, big)
	_, err := e.svc.Upload(context.Background(), bytes.NewReader(body), 0)
	var tooLarge *TooLargeError
	if !errors.As(err, &tooLarge) {
		t.Fatalf("got %v, want *TooLargeError", err)
	}
	if tooLarge.Limit != 4096 {
		t.Errorf("Limit = %d, want 4096", tooLarge.Limit)
	}
	var inv *InvalidError
	if errors.As(err, &inv) {
		t.Errorf("got InvalidError, want only TooLargeError")
	}
	if !strings.Contains(err.Error(), "unpacks to more than 4096 bytes") {
		t.Fatalf("got %v, want the unpacked limit error", err)
	}
	e.assertEmpty(t)
}

func TestUploadIndexesFixture(t *testing.T) {
	e := newService(t)
	res, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if n := e.count(t, "snapshots"); n != 1 {
		t.Errorf("snapshots: %d rows, want 1", n)
	}
	var ns, obj string
	err = e.raw.QueryRow(`SELECT namespace, object FROM resources
		WHERE snapshot_id = $1 AND kind = 'Namespace' AND name = 'kube-system'`, res.ID).Scan(&ns, &obj)
	if err != nil {
		t.Fatal(err)
	}
	if ns != "" {
		t.Errorf("cluster-scoped namespace stored as %q, want empty", ns)
	}
	if want := `{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"kube-system"}}`; obj != want {
		t.Errorf("object stored as %s, want compacted %s", obj, want)
	}
}

func TestUploadReadsFormatV1(t *testing.T) {
	// v1 is v3 without previous logs and the logs and redactions sections of
	// the manifest.
	e := newService(t)
	files := snapshottest.FixtureFiles()
	files[0].Body = `{"format_version": 1, "created_at": "2026-11-17T10:00:00Z"}`
	var v1 []snapshottest.File
	for _, f := range files {
		if !strings.HasSuffix(f.Name, ".previous.log") {
			v1 = append(v1, f)
		}
	}
	if _, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Tarball(t, v1...)), 0); err != nil {
		t.Fatal(err)
	}
	var version int
	if err := e.raw.QueryRow(`SELECT format_version FROM snapshots`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Errorf("format_version stored as %d, want 1", version)
	}
}

func TestUploadStoresLogs(t *testing.T) {
	e := newService(t)
	res, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)), 0)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := e.svc.Store.GetSnapshot(context.Background(), res.ID)
	if err != nil {
		t.Fatal(err)
	}
	if snap.LogsBlobKey != LogsBlobKey(res.ID) {
		t.Fatalf("LogsBlobKey %q, want %q", snap.LogsBlobKey, LogsBlobKey(res.ID))
	}
	for _, tt := range []struct {
		previous bool
		want     string
	}{{false, "started\n"}, {true, "panic: boom\n"}} {
		l, err := e.svc.Store.GetLog(context.Background(), res.ID, store.LogKey{
			Namespace: "default", Pod: "web-0", Container: "nginx", Previous: tt.previous,
		})
		if err != nil {
			t.Fatal(err)
		}
		rc, err := e.svc.Blobs.OpenRange(context.Background(), snap.LogsBlobKey, l.Offset, l.Size)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := io.ReadAll(rc)
		_ = rc.Close()
		if string(got) != tt.want {
			t.Errorf("previous=%v: got %q, want %q", tt.previous, got, tt.want)
		}
	}
}

func TestUploadRejectsBadLogNames(t *testing.T) {
	manifest := F{Name: "manifest.json", Body: snapshottest.Manifest}
	for _, name := range []string{"logs/a/b/.previous.log", "logs/a/b/c.txt", "logs/a/b/c/d.log"} {
		t.Run(name, func(t *testing.T) {
			e := newService(t)
			_, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Tarball(t, manifest, F{Name: name, Body: "x"})), 0)
			var inv *InvalidError
			if !errors.As(err, &inv) {
				t.Fatalf("got %v, want an InvalidError", err)
			}
			e.assertEmpty(t)
		})
	}
}

func TestUploadStoresOnlyDeleteTokenHash(t *testing.T) {
	e := newService(t)
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	e.svc.Now = func() time.Time { return now }

	res, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)), 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if res.DeleteToken == "" {
		t.Fatal("empty DeleteToken")
	}
	if want := now.Add(24 * time.Hour); !res.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", res.ExpiresAt, want)
	}

	var storedHash string
	if err := e.raw.QueryRow(`SELECT delete_token_hash FROM snapshots WHERE id = $1`, res.ID).Scan(&storedHash); err != nil {
		t.Fatal(err)
	}
	if storedHash == res.DeleteToken {
		t.Fatal("raw delete token was stored instead of its hash")
	}
	if strings.Contains(storedHash, res.DeleteToken) {
		t.Fatal("stored delete_token_hash contains raw token")
	}
	if want := HashDeleteToken(res.DeleteToken); storedHash != want {
		t.Errorf("stored hash = %q, want SHA-256 %q", storedHash, want)
	}
	if !VerifyDeleteToken(storedHash, res.DeleteToken) {
		t.Error("VerifyDeleteToken rejected valid token")
	}
	if VerifyDeleteToken(storedHash, "wrong-token") {
		t.Error("VerifyDeleteToken accepted wrong token")
	}
	if VerifyDeleteToken("", "") {
		t.Error("VerifyDeleteToken accepted empty hash and token")
	}
}

type failingDeleteBlobs struct {
	blob.Store
	mu      sync.Mutex
	failKey string
	failErr error
}

func (f *failingDeleteBlobs) Delete(ctx context.Context, key string) error {
	f.mu.Lock()
	shouldFail := f.failKey == key
	err := f.failErr
	f.mu.Unlock()
	if shouldFail {
		return err
	}
	return f.Store.Delete(ctx, key)
}

func TestDeleteByToken(t *testing.T) {
	e := newService(t)
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	e.svc.Now = func() time.Time { return now }

	res, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)), time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Wrong and empty tokens are refused and leave everything intact.
	for _, tok := range []string{"", "wrong-token"} {
		if err := e.svc.Delete(context.Background(), res.ID, tok); !errors.Is(err, ErrInvalidDeleteToken) {
			t.Errorf("Delete(%q) = %v, want ErrInvalidDeleteToken", tok, err)
		}
	}
	if n := e.count(t, "snapshots"); n != 1 {
		t.Fatalf("snapshots = %d after rejected delete, want 1", n)
	}

	// Missing snapshot returns store.ErrNotFound.
	if err := e.svc.Delete(context.Background(), "missing", res.DeleteToken); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Delete(missing) = %v, want store.ErrNotFound", err)
	}

	// Even after expiry (before cleanup), the owner can delete with their token.
	now = now.Add(2 * time.Hour)
	if err := e.svc.Delete(context.Background(), res.ID, res.DeleteToken); err != nil {
		t.Fatalf("Delete(valid): %v", err)
	}
	e.assertEmpty(t)
}

func TestDeleteBlobFailureKeepsRowForRetry(t *testing.T) {
	e := newService(t)
	fb := &failingDeleteBlobs{Store: e.svc.Blobs}
	e.svc.Blobs = fb

	res, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	fb.failKey = LogsBlobKey(res.ID)
	fb.failErr = errors.New("disk hiccup")

	if err := e.svc.Delete(context.Background(), res.ID, res.DeleteToken); err == nil {
		t.Fatal("expected blob deletion error")
	}
	// Row must remain so a retry can finish cleaning up.
	if n := e.count(t, "snapshots"); n != 1 {
		t.Fatalf("snapshots = %d after blob failure, want 1", n)
	}

	// Retry succeeds even though the tarball blob was already deleted on the first attempt.
	fb.failKey = ""
	if err := e.svc.Delete(context.Background(), res.ID, res.DeleteToken); err != nil {
		t.Fatalf("retry Delete: %v", err)
	}
	e.assertEmpty(t)
}

func TestDeleteExpiredWithInjectedClock(t *testing.T) {
	e := newService(t)
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	e.svc.Now = func() time.Time { return now }
	e.svc.CleanupBatch = 2 // force multiple batches

	fixture := snapshottest.Fixture(t)
	upload := func(ttl time.Duration) UploadResult {
		t.Helper()
		res, err := e.svc.Upload(context.Background(), bytes.NewReader(fixture), ttl)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	// 3 snapshots expiring within 2 hours, 1 expiring in 10 hours.
	exp1 := upload(30 * time.Minute)
	exp2 := upload(60 * time.Minute)
	exp3 := upload(90 * time.Minute)
	kept := upload(10 * time.Hour)

	// One legacy snapshot with NULL expires_at and empty logs_blob_key.
	tx, err := e.svc.Store.BeginSnapshot(context.Background(), store.Snapshot{
		ID: "legacy", UploadedAt: now, CapturedAt: now, FormatVersion: 1,
		Manifest: []byte(snapshottest.Manifest), BlobKey: BlobKey("legacy"), SizeBytes: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.Blobs.Put(context.Background(), BlobKey("legacy"), strings.NewReader("tar")); err != nil {
		t.Fatal(err)
	}

	// One expired legacy snapshot with empty logs_blob_key to verify empty key is skipped.
	legacyExpTime := now.Add(45 * time.Minute)
	tx, err = e.svc.Store.BeginSnapshot(context.Background(), store.Snapshot{
		ID: "legacy-expired", UploadedAt: now, CapturedAt: now, FormatVersion: 1,
		Manifest: []byte(snapshottest.Manifest), BlobKey: BlobKey("legacy-expired"),
		SizeBytes: 10, ExpiresAt: &legacyExpTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := e.svc.Blobs.Put(context.Background(), BlobKey("legacy-expired"), strings.NewReader("tar")); err != nil {
		t.Fatal(err)
	}

	// Advance clock by 2 hours: exp1, exp2, exp3, legacy-expired are expired; kept and legacy stay.
	now = now.Add(2 * time.Hour)
	deleted, err := e.svc.DeleteExpired(context.Background())
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if deleted != 4 {
		t.Errorf("deleted = %d, want 4", deleted)
	}

	// Verify expired rows and blobs are gone.
	for _, id := range []string{exp1.ID, exp2.ID, exp3.ID, "legacy-expired"} {
		if _, err := e.svc.Store.GetSnapshot(context.Background(), id); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("expired snapshot %s still in store: %v", id, err)
		}
		for _, key := range []string{BlobKey(id), LogsBlobKey(id)} {
			if _, err := os.Stat(filepath.Join(e.blobDir, filepath.FromSlash(key))); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("expired blob %s still exists (err=%v)", key, err)
			}
		}
	}

	// Verify unexpired and no-expiry snapshots and their blobs remain.
	for _, id := range []string{kept.ID, "legacy"} {
		if _, err := e.svc.Store.GetSnapshot(context.Background(), id); err != nil {
			t.Errorf("snapshot %s should remain: %v", id, err)
		}
		if _, err := os.Stat(filepath.Join(e.blobDir, filepath.FromSlash(BlobKey(id)))); err != nil {
			t.Errorf("blob for %s missing: %v", id, err)
		}
	}
	if _, err := os.Stat(filepath.Join(e.blobDir, filepath.FromSlash(LogsBlobKey(kept.ID)))); err != nil {
		t.Errorf("logs blob for %s missing: %v", kept.ID, err)
	}
	if n := e.count(t, "snapshots"); n != 2 {
		t.Errorf("remaining snapshots = %d, want 2", n)
	}
	// Resources and logs belong only to kept.
	if n := e.count(t, "resources"); n != 4 {
		t.Errorf("remaining resources = %d, want 4", n)
	}
	if n := e.count(t, "logs"); n != 2 {
		t.Errorf("remaining logs = %d, want 2", n)
	}
}

func TestDeleteExpiredStopsOnBlobFailureAndRetries(t *testing.T) {
	e := newService(t)
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	e.svc.Now = func() time.Time { return now }
	fb := &failingDeleteBlobs{Store: e.svc.Blobs}
	e.svc.Blobs = fb

	res, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Hour)
	fb.failKey = BlobKey(res.ID)
	fb.failErr = errors.New("temporary error")

	n, err := e.svc.DeleteExpired(context.Background())
	if err == nil || n != 0 {
		t.Fatalf("DeleteExpired = (%d, %v), want (0, err)", n, err)
	}
	if count := e.count(t, "snapshots"); count != 1 {
		t.Fatalf("snapshots = %d, want 1", count)
	}

	fb.failKey = ""
	n, err = e.svc.DeleteExpired(context.Background())
	if err != nil || n != 1 {
		t.Fatalf("retry DeleteExpired = (%d, %v), want (1, nil)", n, err)
	}
	e.assertEmpty(t)
}

func TestRunCleanupStopsOnCancel(t *testing.T) {
	e := newService(t)
	var mu sync.Mutex
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	e.svc.Now = func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return now
	}

	res, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	now = now.Add(2 * time.Hour)
	mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		e.svc.RunCleanup(ctx, 10*time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil)))
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := e.svc.Store.GetSnapshot(context.Background(), res.ID); errors.Is(err, store.ErrNotFound) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := e.svc.Store.GetSnapshot(context.Background(), res.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("RunCleanup did not remove expired snapshot: %v", err)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunCleanup did not stop after context cancellation")
	}
}
