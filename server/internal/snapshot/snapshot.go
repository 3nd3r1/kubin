// Package snapshot takes an uploaded snapshot tarball, checks it against the
// v1 format, keeps the raw tarball in the blob store and indexes every
// resource in the database.
package snapshot

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/3nd3r1/kubin/server/internal/blob"
	"github.com/3nd3r1/kubin/server/internal/store"
)

// DefaultMaxUnpacked caps what an archive may expand to, so a small upload
// can't unpack into gigabytes.
const DefaultMaxUnpacked = 1 << 30

// Service ingests uploads.
type Service struct {
	Store store.Store
	Blobs blob.Store
	// TempDir holds uploads while they are checked. Empty means os.TempDir.
	TempDir string
	// MaxUnpacked is the most an archive may unpack to. Zero means
	// DefaultMaxUnpacked.
	MaxUnpacked int64
	Now         func() time.Time
}

// idBytes is 15 random bytes: 120 bits, 24 base32 characters with no padding.
const idBytes = 15

var idEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// NewID returns a random, unguessable snapshot ID. The ID is the only access
// control until there is auth.
func NewID() string {
	b := make([]byte, idBytes)
	_, _ = rand.Read(b) // never fails, see crypto/rand
	return strings.ToLower(idEncoding.EncodeToString(b))
}

// BlobKey is where a snapshot's raw tarball is kept.
func BlobKey(id string) string { return "snapshots/" + id + ".tar.gz" }

// Upload reads a snapshot tarball from r and stores it. It returns the new
// snapshot's ID. An *InvalidError means the upload was rejected; errors from
// reading r are returned as they are.
func (s *Service) Upload(ctx context.Context, r io.Reader) (string, error) {
	// Spool the upload so it can be read twice: once to check and index it,
	// once into the blob store.
	f, err := os.CreateTemp(s.TempDir, "upload-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	size, err := io.Copy(f, r)
	if err != nil {
		return "", err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	maxUnpacked := s.MaxUnpacked
	if maxUnpacked == 0 {
		maxUnpacked = DefaultMaxUnpacked
	}
	a, m, rawManifest, err := openArchive(f, maxUnpacked)
	if err != nil {
		return "", err
	}

	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	id := NewID()
	snap := store.Snapshot{
		ID:            id,
		UploadedAt:    now(),
		CapturedAt:    m.CreatedAt,
		FormatVersion: m.FormatVersion,
		Manifest:      rawManifest,
		BlobKey:       BlobKey(id),
		SizeBytes:     size,
	}
	tx, err := s.Store.BeginSnapshot(ctx, snap)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	for {
		e, err := a.nextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if e.Resource == nil {
			continue
		}
		if err := tx.AddResource(ctx, *e.Resource); err != nil {
			return "", err
		}
	}

	// The blob goes in before the rows are committed, so a committed
	// snapshot always has its tarball.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	if err := s.Blobs.Put(ctx, snap.BlobKey, f); err != nil {
		return "", fmt.Errorf("store tarball: %w", err)
	}
	if err := tx.Commit(); err != nil {
		// Don't leave an orphan blob. Use a fresh context: ctx may be the
		// reason the commit failed.
		delCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		return "", errors.Join(fmt.Errorf("commit snapshot: %w", err), s.Blobs.Delete(delCtx, snap.BlobKey))
	}
	return id, nil
}
