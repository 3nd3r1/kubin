package blob

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestFSRoundTrip(t *testing.T) {
	ctx := context.Background()
	s, err := NewFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Put(ctx, "snapshots/abc.tar.gz", strings.NewReader("data")); err != nil {
		t.Fatal(err)
	}
	rc, err := s.Open(ctx, "snapshots/abc.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if string(got) != "data" {
		t.Fatalf("got %q, want data", got)
	}

	if err := s.Delete(ctx, "snapshots/abc.tar.gz"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Open(ctx, "snapshots/abc.tar.gz"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Open after Delete: got %v, want ErrNotFound", err)
	}
	if err := s.Delete(ctx, "snapshots/abc.tar.gz"); err != nil {
		t.Fatalf("second Delete: %v", err)
	}
}

func TestFSFailedPutLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFS(dir)
	if err != nil {
		t.Fatal(err)
	}
	r := io.MultiReader(strings.NewReader("partial"), errReader{})
	if err := s.Put(context.Background(), "x", r); err == nil {
		t.Fatal("Put with a failing reader returned nil")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("left %d files behind", len(entries))
	}
}

func TestFSRejectsUnsafeKeys(t *testing.T) {
	s, err := NewFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"", "/abs", "../up", "a/../b", "a//b", "a/./b"} {
		if err := s.Put(context.Background(), key, strings.NewReader("x")); err == nil {
			t.Errorf("Put(%q) returned nil", key)
		}
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }
