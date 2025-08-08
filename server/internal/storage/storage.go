package storage

import (
	"context"

	"github.com/3nd3r1/kubin/server/internal/domain"
)

// Storage defines the interface for snapshot storage operations
type Storage interface {
	// Snapshot operations
	StoreSnapshot(ctx context.Context, snapshot *domain.Snapshot, content []byte) error
	GetSnapshot(ctx context.Context, id string) (*domain.Snapshot, error)
	ListSnapshots(ctx context.Context, limit, offset int) ([]*domain.Snapshot, error)
	DeleteSnapshot(ctx context.Context, id string) error
	GetSnapshotContent(ctx context.Context, id string) ([]byte, error)

	// Health check
	Health(ctx context.Context) error
}
