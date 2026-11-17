// Package store keeps snapshot metadata and the resource index. SQLite is the
// default; Postgres comes with TASK-26. Queries use $N placeholders and
// standard SQL so both can share them, and each dialect has its own migrations.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a snapshot or resource doesn't exist.
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

	Ping(ctx context.Context) error
	Close() error
}

// SnapshotTx writes one snapshot and its resources atomically.
type SnapshotTx interface {
	AddResource(ctx context.Context, r Resource) error
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
