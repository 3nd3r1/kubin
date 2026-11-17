package snapshot

import (
	"archive/tar"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

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
		{"unknown version", snapshottest.Tarball(t, F{Name: "manifest.json", Body: `{"format_version":2}`}), "unsupported format_version 2"},
		{"missing version", snapshottest.Tarball(t, F{Name: "manifest.json", Body: `{}`}), "unsupported format_version 0"},
		{"dot dot", snapshottest.Tarball(t, manifest, F{Name: "resources/core/v1/Pod/../web.json", Body: "{}"}), "unsafe path segment"},
		{"absolute", snapshottest.Tarball(t, manifest, F{Name: "/etc/passwd", Body: "x"}), "empty path segment"},
		{"outside layout", snapshottest.Tarball(t, manifest, F{Name: "extra.txt", Body: "x"}), "not part of the v1 layout"},
		{"short resource path", snapshottest.Tarball(t, manifest, F{Name: "resources/core/v1/Pod/web.json", Body: "{}"}), "not part of the v1 layout"},
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
			_, err := e.svc.Upload(context.Background(), bytes.NewReader(tt.body))
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
	_, err := e.svc.Upload(context.Background(), bytes.NewReader(body))
	if err == nil || !strings.Contains(err.Error(), "unpacks to more than 4096 bytes") {
		t.Fatalf("got %v, want the unpacked limit error", err)
	}
	e.assertEmpty(t)
}

func TestUploadIndexesFixture(t *testing.T) {
	e := newService(t)
	id, err := e.svc.Upload(context.Background(), bytes.NewReader(snapshottest.Fixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	if n := e.count(t, "snapshots"); n != 1 {
		t.Errorf("snapshots: %d rows, want 1", n)
	}
	var ns, obj string
	err = e.raw.QueryRow(`SELECT namespace, object FROM resources
		WHERE snapshot_id = $1 AND kind = 'Namespace' AND name = 'kube-system'`, id).Scan(&ns, &obj)
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
