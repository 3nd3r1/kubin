// Package blob keeps raw snapshot tarballs. A local directory is the default;
// S3 comes with TASK-20.
package blob

import (
	"context"
	"errors"
	"io"
)

// ErrNotFound is returned by Open for a key that was never stored.
var ErrNotFound = errors.New("blob not found")

// Store keeps opaque blobs under string keys. Keys are slash-separated paths
// like "snapshots/<id>.tar.gz".
type Store interface {
	// Put stores r under key, replacing any existing blob. A reader that
	// fails part way leaves nothing behind.
	Put(ctx context.Context, key string, r io.Reader) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes key. Deleting a missing key is not an error.
	Delete(ctx context.Context, key string) error
}
