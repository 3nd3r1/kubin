package store

import (
	"context"
	"path/filepath"
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
	if n := count(t, s2, `SELECT COUNT(*) FROM schema_migrations`); n != 1 {
		t.Fatalf("schema_migrations has %d rows, want 1", n)
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
