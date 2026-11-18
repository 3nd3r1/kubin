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

func TestFSOpenRange(t *testing.T) {
	ctx := context.Background()
	s, err := NewFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Put(ctx, "snapshots/abc.logs.txt", strings.NewReader("0123456789")); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		off, n int64
		want   string
	}{
		{0, 10, "0123456789"},
		{3, 4, "3456"},
		{8, 10, "89"},
		{10, 5, ""},
		{4, 0, ""},
	} {
		rc, err := s.OpenRange(ctx, "snapshots/abc.logs.txt", tt.off, tt.n)
		if err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil || string(got) != tt.want {
			t.Errorf("OpenRange(%d, %d): got %q %v, want %q", tt.off, tt.n, got, err, tt.want)
		}
	}
	if _, err := s.OpenRange(ctx, "snapshots/missing", 0, 1); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing blob: got %v, want ErrNotFound", err)
	}
	if _, err := s.OpenRange(ctx, "snapshots/abc.logs.txt", -1, 1); err == nil {
		t.Error("negative offset accepted")
	}
}
