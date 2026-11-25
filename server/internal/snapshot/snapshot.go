// Package snapshot takes an uploaded snapshot tarball, checks it against the
// format, keeps the raw tarball in the blob store, indexes every resource in
// the database and copies the logs into a blob of their own so they can be
// read without unpacking the tarball.
package snapshot

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/3nd3r1/kubin/server/internal/blob"
	"github.com/3nd3r1/kubin/server/internal/store"
)

const (
	// DefaultMaxUnpacked caps what an archive may expand to, so a small upload
	// can't unpack into gigabytes.
	DefaultMaxUnpacked = 1 << 30
	// DefaultTTL is how long an uploaded snapshot lives when no expires_in is given.
	DefaultTTL = 168 * time.Hour
	// DefaultMaxTTL caps the lifetime requested via expires_in.
	DefaultMaxTTL = 720 * time.Hour
	// DefaultCleanupBatch limits how many expired snapshots are listed per pass.
	DefaultCleanupBatch = 100
)

// ErrInvalidDeleteToken is returned when a delete request carries the wrong token.
var ErrInvalidDeleteToken = errors.New("invalid delete token")

// UploadResult holds the details returned to the caller after storing a snapshot.
type UploadResult struct {
	ID          string
	ExpiresAt   time.Time
	DeleteToken string
}

// Service ingests, deletes, and cleans up snapshots.
type Service struct {
	Store store.Store
	Blobs blob.Store
	// TempDir holds uploads while they are checked. Empty means os.TempDir.
	TempDir string
	// MaxUnpacked is the most an archive may unpack to. Zero means
	// DefaultMaxUnpacked.
	MaxUnpacked int64
	// DefaultTTL is used when Upload receives a zero ttl. Zero means DefaultTTL.
	DefaultTTL time.Duration
	// MaxTTL caps the ttl accepted by Upload. Zero means DefaultMaxTTL.
	MaxTTL time.Duration
	// CleanupBatch is the maximum number of expired snapshots listed per pass.
	// Zero means DefaultCleanupBatch.
	CleanupBatch int
	Now          func() time.Time
}

// idBytes is 15 random bytes: 120 bits, 24 base32 characters with no padding.
const idBytes = 15

// deleteTokenBytes is 32 random bytes (256 bits) of bearer token entropy.
const deleteTokenBytes = 32

var idEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// NewID returns a random, unguessable snapshot ID. The ID is the only access
// control until there is auth.
func NewID() string {
	b := make([]byte, idBytes)
	_, _ = rand.Read(b) // never fails, see crypto/rand
	return strings.ToLower(idEncoding.EncodeToString(b))
}

// NewDeleteToken returns a random bearer token that authorizes deleting a snapshot.
func NewDeleteToken() string {
	b := make([]byte, deleteTokenBytes)
	_, _ = rand.Read(b) // never fails, see crypto/rand
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashDeleteToken returns the lowercase hex SHA-256 digest stored in the database.
func HashDeleteToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// VerifyDeleteToken compares token against a stored SHA-256 hex digest in constant time.
func VerifyDeleteToken(storedHash, token string) bool {
	if storedHash == "" || token == "" {
		return false
	}
	got := HashDeleteToken(token)
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(got)) == 1
}

// BlobKey is where a snapshot's raw tarball is kept.
func BlobKey(id string) string { return "snapshots/" + id + ".tar.gz" }

// LogsBlobKey is where a snapshot's logs are kept, back to back. The logs
// table says where each one starts.
func LogsBlobKey(id string) string { return "snapshots/" + id + ".logs.txt" }

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// Upload reads a snapshot tarball from r and stores it with the given ttl (or
// DefaultTTL when ttl is zero). An *InvalidError means the upload was malformed,
// a *TooLargeError means it unpacked past MaxUnpacked; errors from reading r are
// returned as they are.
func (s *Service) Upload(ctx context.Context, r io.Reader, ttl time.Duration) (UploadResult, error) {
	if ttl == 0 {
		ttl = s.DefaultTTL
		if ttl <= 0 {
			ttl = DefaultTTL
		}
	}
	maxTTL := s.MaxTTL
	if maxTTL <= 0 {
		maxTTL = DefaultMaxTTL
	}
	if ttl < 0 {
		return UploadResult{}, invalid("expires_in must be positive (maximum %s)", maxTTL)
	}
	if ttl > maxTTL {
		return UploadResult{}, invalid("expires_in %s exceeds maximum %s", ttl, maxTTL)
	}

	// Spool the upload so it can be read twice: once to check and index it,
	// once into the blob store.
	f, err := os.CreateTemp(s.TempDir, "upload-*.tar.gz")
	if err != nil {
		return UploadResult{}, err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	size, err := io.Copy(f, r)
	if err != nil {
		return UploadResult{}, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return UploadResult{}, err
	}

	maxUnpacked := s.MaxUnpacked
	if maxUnpacked == 0 {
		maxUnpacked = DefaultMaxUnpacked
	}
	a, m, rawManifest, err := openArchive(f, maxUnpacked)
	if err != nil {
		return UploadResult{}, err
	}

	uploadedAt := s.now().UTC().Round(0)
	expiresAt := uploadedAt.Add(ttl)
	id := NewID()
	deleteToken := NewDeleteToken()
	snap := store.Snapshot{
		ID:              id,
		UploadedAt:      uploadedAt,
		CapturedAt:      m.CreatedAt,
		FormatVersion:   m.FormatVersion,
		Manifest:        rawManifest,
		BlobKey:         BlobKey(id),
		SizeBytes:       size,
		LogsBlobKey:     LogsBlobKey(id),
		ExpiresAt:       &expiresAt,
		DeleteTokenHash: HashDeleteToken(deleteToken),
	}
	tx, err := s.Store.BeginSnapshot(ctx, snap)
	if err != nil {
		return UploadResult{}, err
	}
	defer func() { _ = tx.Rollback() }()

	logs, err := os.CreateTemp(s.TempDir, "logs-*.txt")
	if err != nil {
		return UploadResult{}, err
	}
	defer func() {
		_ = logs.Close()
		_ = os.Remove(logs.Name())
	}()

	var logsSize int64
	for {
		e, err := a.nextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return UploadResult{}, err
		}
		switch {
		case e.Resource != nil:
			if err := tx.AddResource(ctx, *e.Resource); err != nil {
				return UploadResult{}, err
			}
		case e.Log != nil:
			n, err := a.copyBody(e, logs)
			if err != nil {
				return UploadResult{}, err
			}
			if err := tx.AddLog(ctx, store.Log{LogKey: *e.Log, Offset: logsSize, Size: n}); err != nil {
				return UploadResult{}, err
			}
			logsSize += n
		}
	}

	// The blobs go in before the rows are committed, so a committed
	// snapshot always has them. Anything stored is removed if a later step
	// fails, using a fresh context: ctx may be the reason it failed.
	var stored []string
	fail := func(err error) (UploadResult, error) {
		delCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		for _, key := range stored {
			err = errors.Join(err, s.Blobs.Delete(delCtx, key))
		}
		return UploadResult{}, err
	}
	for _, b := range []struct {
		key, what string
		f         *os.File
	}{{snap.BlobKey, "tarball", f}, {snap.LogsBlobKey, "logs", logs}} {
		if _, err := b.f.Seek(0, io.SeekStart); err != nil {
			return fail(err)
		}
		if err := s.Blobs.Put(ctx, b.key, b.f); err != nil {
			return fail(fmt.Errorf("store %s: %w", b.what, err))
		}
		stored = append(stored, b.key)
	}
	if err := tx.Commit(); err != nil {
		return fail(fmt.Errorf("commit snapshot: %w", err))
	}
	return UploadResult{ID: id, ExpiresAt: expiresAt, DeleteToken: deleteToken}, nil
}

// Delete verifies the bearer token and removes a snapshot's blobs and rows.
// It returns store.ErrNotFound if the snapshot does not exist and
// ErrInvalidDeleteToken if the token does not match.
func (s *Service) Delete(ctx context.Context, id, token string) error {
	snap, err := s.Store.GetSnapshot(ctx, id)
	if err != nil {
		return err
	}
	if !VerifyDeleteToken(snap.DeleteTokenHash, token) {
		return ErrInvalidDeleteToken
	}
	return s.deleteSnapshot(ctx, snap.ID, snap.BlobKey, snap.LogsBlobKey)
}

// deleteSnapshot removes both blobs before deleting the database row, so a blob
// deletion failure leaves a row the next run can retry and never orphans blobs.
func (s *Service) deleteSnapshot(ctx context.Context, id, blobKey, logsBlobKey string) error {
	if blobKey != "" {
		if err := s.Blobs.Delete(ctx, blobKey); err != nil {
			return fmt.Errorf("delete tarball %s: %w", blobKey, err)
		}
	}
	if logsBlobKey != "" {
		if err := s.Blobs.Delete(ctx, logsBlobKey); err != nil {
			return fmt.Errorf("delete logs %s: %w", logsBlobKey, err)
		}
	}
	if err := s.Store.DeleteSnapshot(ctx, id); err != nil {
		return fmt.Errorf("delete snapshot %s: %w", id, err)
	}
	return nil
}
