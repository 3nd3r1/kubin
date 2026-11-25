// Package store keeps snapshot metadata and the resource index. SQLite is the
// default; Postgres comes with TASK-26. Queries use $N placeholders and
// standard SQL so both can share them, and each dialect has its own migrations.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a snapshot, resource or log doesn't exist.
var ErrNotFound = errors.New("not found")

// Store is the database behind the server.
type Store interface {
	// BeginSnapshot starts writing a new snapshot. Nothing is visible to
	// readers until Commit.
	BeginSnapshot(ctx context.Context, s Snapshot) (SnapshotTx, error)

	GetSnapshot(ctx context.Context, id string) (Snapshot, error)
	// CountResources counts what was indexed per kind, sorted by group,
	// version, kind.
	CountResources(ctx context.Context, id string) ([]KindCount, error)
	// ListResources returns a snapshot's resources sorted by group, version,
	// kind, namespace, name. It doesn't check that the snapshot exists.
	ListResources(ctx context.Context, id string, f Filter) ([]Resource, error)
	// GetResource returns one resource. When the key's Group is empty and
	// several groups have the kind, the first by group and version wins.
	GetResource(ctx context.Context, id string, k Key) (Resource, error)
	// GetLog says where one container log sits in the snapshot's logs blob.
	GetLog(ctx context.Context, id string, k LogKey) (Log, error)
	// ListExpired returns up to limit snapshots whose expires_at is at or
	// before before, oldest first. Snapshots with a nil ExpiresAt are skipped.
	ListExpired(ctx context.Context, before time.Time, limit int) ([]ExpiredSnapshot, error)
	// DeleteSnapshot removes a snapshot and its cascaded resources and logs.
	// Deleting a snapshot that doesn't exist is not an error.
	DeleteSnapshot(ctx context.Context, id string) error

	Ping(ctx context.Context) error
	Close() error
}

// SnapshotTx writes one snapshot and its resources atomically.
type SnapshotTx interface {
	AddResource(ctx context.Context, r Resource) error
	AddLog(ctx context.Context, l Log) error
	Commit() error
	// Rollback discards the snapshot. It is safe to call after Commit.
	Rollback() error
}

type Snapshot struct {
	ID            string
	UploadedAt    time.Time
	CapturedAt    time.Time // manifest created_at
	FormatVersion int
	Manifest      []byte // manifest.json as uploaded
	BlobKey       string
	SizeBytes     int64 // compressed tarball size
	// LogsBlobKey holds every log of the snapshot back to back. Empty for
	// snapshots stored before logs were served.
	LogsBlobKey string
	// ExpiresAt is when the snapshot stops being served and becomes eligible
	// for cleanup. Nil means it never expires (snapshots stored before expiry).
	ExpiresAt *time.Time
	// DeleteTokenHash is the hex SHA-256 of the snapshot's delete token. Empty
	// for snapshots stored before delete tokens existed.
	DeleteTokenHash string
}

// ExpiredSnapshot identifies a snapshot ready to be removed along with its blobs.
type ExpiredSnapshot struct {
	ID          string
	BlobKey     string
	LogsBlobKey string
}

// Resource is one object from a snapshot. Namespace is empty for
// cluster-scoped kinds. Group is "core" for the core API group, as in the
// tarball layout.
type Resource struct {
	Group     string
	Version   string
	Kind      string
	Namespace string
	Name      string
	Object    []byte // the object as JSON
}

type KindCount struct {
	Group   string
	Version string
	Kind    string
	Count   int
}

// Filter narrows ListResources. Empty fields match everything. Set
// ClusterScoped to select only resources without a namespace.
type Filter struct {
	Group         string
	Kind          string
	Namespace     string
	ClusterScoped bool
}

// Key names one resource. Namespace is empty for cluster-scoped kinds and
// Group may be empty to match any group.
type Key struct {
	Group     string
	Kind      string
	Namespace string
	Name      string
}

// LogKey names one container log. Previous selects the run before the last
// restart.
type LogKey struct {
	Namespace string
	Pod       string
	Container string
	Previous  bool
}

// Log is where one container log sits in the snapshot's logs blob.
type Log struct {
	LogKey
	Offset int64
	Size   int64
}
