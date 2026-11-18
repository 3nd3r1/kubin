package store

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func openTest(t *testing.T) (*SQL, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kubin.db")
	s, err := OpenSQLite(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, path
}

func testSnapshot(id string) Snapshot {
	now := time.Now()
	return Snapshot{
		ID: id, UploadedAt: now, CapturedAt: now, FormatVersion: 1,
		Manifest: []byte(`{"format_version":1}`), BlobKey: "snapshots/" + id, SizeBytes: 10,
	}
}

func count(t *testing.T, s *SQL, query string, args ...any) int {
	t.Helper()
	var n int
	if err := s.db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestMigrationsRunOnceOnReopen(t *testing.T) {
	s, path := openTest(t)
	_ = s.Close()
	s2, err := OpenSQLite(context.Background(), path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = s2.Close() }()
	files, err := migrations.ReadDir("migrations/sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if n := count(t, s2, `SELECT COUNT(*) FROM schema_migrations`); n != len(files) {
		t.Fatalf("schema_migrations has %d rows, want %d", n, len(files))
	}
}

func TestSnapshotCommitAndRollback(t *testing.T) {
	ctx := context.Background()
	s, _ := openTest(t)

	tx, err := s.BeginSnapshot(ctx, testSnapshot("kept"))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []Resource{
		{Group: "core", Version: "v1", Kind: "Namespace", Name: "default", Object: []byte(`{}`)},
		{Group: "core", Version: "v1", Kind: "Pod", Namespace: "default", Name: "web", Object: []byte(`{}`)},
	} {
		if err := tx.AddResource(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback after Commit: %v", err)
	}

	tx, err = s.BeginSnapshot(ctx, testSnapshot("dropped"))
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.AddResource(ctx, Resource{Group: "core", Version: "v1", Kind: "Pod", Namespace: "a", Name: "b", Object: []byte(`{}`)}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	if n := count(t, s, `SELECT COUNT(*) FROM snapshots`); n != 1 {
		t.Errorf("snapshots: %d rows, want 1", n)
	}
	if n := count(t, s, `SELECT COUNT(*) FROM resources WHERE snapshot_id = $1`, "kept"); n != 2 {
		t.Errorf("resources for kept: %d, want 2", n)
	}
	if n := count(t, s, `SELECT COUNT(*) FROM resources WHERE snapshot_id = $1`, "dropped"); n != 0 {
		t.Errorf("resources for dropped: %d, want 0", n)
	}
}

func TestDuplicateResourceFails(t *testing.T) {
	ctx := context.Background()
	s, _ := openTest(t)
	tx, err := s.BeginSnapshot(ctx, testSnapshot("dup"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	r := Resource{Group: "core", Version: "v1", Kind: "Pod", Namespace: "a", Name: "b", Object: []byte(`{}`)}
	if err := tx.AddResource(ctx, r); err != nil {
		t.Fatal(err)
	}
	if err := tx.AddResource(ctx, r); err == nil {
		t.Fatal("second insert of the same resource returned nil")
	}
}

func TestReads(t *testing.T) {
	ctx := context.Background()
	s, _ := openTest(t)
	snap := testSnapshot("snap")
	tx, err := s.BeginSnapshot(ctx, snap)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range []Resource{
		{Group: "core", Version: "v1", Kind: "Namespace", Name: "default", Object: []byte(`{"n":1}`)},
		{Group: "core", Version: "v1", Kind: "Event", Namespace: "default", Name: "e1", Object: []byte(`{"core":true}`)},
		{Group: "events.k8s.io", Version: "v1", Kind: "Event", Namespace: "default", Name: "e1", Object: []byte(`{"core":false}`)},
		{Group: "core", Version: "v1", Kind: "Pod", Namespace: "default", Name: "b", Object: []byte(`{}`)},
		{Group: "core", Version: "v1", Kind: "Pod", Namespace: "default", Name: "a", Object: []byte(`{}`)},
		{Group: "core", Version: "v1", Kind: "Pod", Namespace: "other", Name: "c", Object: []byte(`{}`)},
	} {
		if err := tx.AddResource(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetSnapshot(ctx, "snap")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Manifest) != string(snap.Manifest) || !got.UploadedAt.Equal(snap.UploadedAt) || got.BlobKey != snap.BlobKey {
		t.Errorf("GetSnapshot = %+v, want %+v", got, snap)
	}
	if _, err := s.GetSnapshot(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetSnapshot(missing) = %v, want ErrNotFound", err)
	}

	counts, err := s.CountResources(ctx, "snap")
	if err != nil {
		t.Fatal(err)
	}
	want := []KindCount{
		{"core", "v1", "Event", 1}, {"core", "v1", "Namespace", 1}, {"core", "v1", "Pod", 3},
		{"events.k8s.io", "v1", "Event", 1},
	}
	if fmt.Sprint(counts) != fmt.Sprint(want) {
		t.Errorf("CountResources = %v, want %v", counts, want)
	}

	names := func(f Filter) string {
		rs, err := s.ListResources(ctx, "snap", f)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, r := range rs {
			out = append(out, r.Kind+":"+r.Namespace+"/"+r.Name)
		}
		return strings.Join(out, " ")
	}
	for _, tt := range []struct {
		f    Filter
		want string
	}{
		{Filter{Kind: "Pod"}, "Pod:default/a Pod:default/b Pod:other/c"},
		{Filter{Kind: "Pod", Namespace: "other"}, "Pod:other/c"},
		{Filter{ClusterScoped: true}, "Namespace:/default"},
		{Filter{Group: "events.k8s.io"}, "Event:default/e1"},
		{Filter{Kind: "Nope"}, ""},
	} {
		if got := names(tt.f); got != tt.want {
			t.Errorf("ListResources(%+v) = %q, want %q", tt.f, got, tt.want)
		}
	}

	r, err := s.GetResource(ctx, "snap", Key{Kind: "Event", Namespace: "default", Name: "e1"})
	if err != nil || string(r.Object) != `{"core":true}` {
		t.Errorf("GetResource without group = %s, %v; want the core Event", r.Object, err)
	}
	r, err = s.GetResource(ctx, "snap", Key{Group: "events.k8s.io", Kind: "Event", Namespace: "default", Name: "e1"})
	if err != nil || string(r.Object) != `{"core":false}` {
		t.Errorf("GetResource with group = %s, %v; want the events.k8s.io Event", r.Object, err)
	}
	if _, err := s.GetResource(ctx, "snap", Key{Kind: "Pod", Namespace: "default", Name: "zz"}); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetResource(missing) = %v, want ErrNotFound", err)
	}
}

func TestLogs(t *testing.T) {
	ctx := context.Background()
	s, _ := openTest(t)

	snap := testSnapshot("withlogs")
	snap.LogsBlobKey = "snapshots/withlogs.logs.txt"
	tx, err := s.BeginSnapshot(ctx, snap)
	if err != nil {
		t.Fatal(err)
	}
	current := Log{LogKey: LogKey{Namespace: "prod", Pod: "web", Container: "app"}, Offset: 0, Size: 10}
	previous := Log{LogKey: LogKey{Namespace: "prod", Pod: "web", Container: "app", Previous: true}, Offset: 10, Size: 5}
	for _, l := range []Log{current, previous} {
		if err := tx.AddLog(ctx, l); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.AddLog(ctx, current); err == nil {
		t.Error("adding the same log twice succeeded")
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetSnapshot(ctx, "withlogs")
	if err != nil {
		t.Fatal(err)
	}
	if got.LogsBlobKey != snap.LogsBlobKey {
		t.Errorf("LogsBlobKey %q, want %q", got.LogsBlobKey, snap.LogsBlobKey)
	}
	for _, want := range []Log{current, previous} {
		l, err := s.GetLog(ctx, "withlogs", want.LogKey)
		if err != nil {
			t.Fatal(err)
		}
		if l != want {
			t.Errorf("GetLog: got %+v, want %+v", l, want)
		}
	}
	for _, k := range []LogKey{
		{Namespace: "prod", Pod: "web", Container: "other"},
		{Namespace: "prod", Pod: "api", Container: "app"},
	} {
		if _, err := s.GetLog(ctx, "withlogs", k); !errors.Is(err, ErrNotFound) {
			t.Errorf("GetLog %+v: got %v, want ErrNotFound", k, err)
		}
	}
	if _, err := s.GetLog(ctx, "missing", current.LogKey); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetLog on a missing snapshot: got %v, want ErrNotFound", err)
	}

	tx, err = s.BeginSnapshot(ctx, testSnapshot("dropped"))
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.AddLog(ctx, current); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if n := count(t, s, `SELECT COUNT(*) FROM logs WHERE snapshot_id = $1`, "dropped"); n != 0 {
		t.Errorf("rolled back snapshot left %d log rows", n)
	}
}
