package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/3nd3r1/kubin/server/internal/domain"
	"github.com/3nd3r1/kubin/server/internal/log"
	"github.com/3nd3r1/kubin/server/internal/storage"
)

type SnapshotService struct {
	storage storage.Storage
}

// NewSnapshotService constructs the service and internally initializes storage.
func NewSnapshotService() (*SnapshotService, error) {
	// Storage is responsible for sourcing its own configuration
	st, err := storage.NewPostgresStorage()
	if err != nil {
		return nil, err
	}
	return &SnapshotService{storage: st}, nil
}

func (s *SnapshotService) CreateSnapshot(ctx context.Context, metadata *domain.SnapshotMetadata, content []byte) (*domain.Snapshot, error) {
	// Generate unique ID
	id, err := generateSnapshotID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate snapshot ID: %w", err)
	}

	// Create snapshot object
	snapshot := &domain.Snapshot{
		ID:        id,
		Metadata:  metadata,
		Size:      int64(len(content)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store snapshot data
	if err := s.storage.StoreSnapshot(ctx, snapshot, content); err != nil {
		return nil, fmt.Errorf("failed to store snapshot: %w", err)
	}

	log.With(
		"id", snapshot.ID,
		"name", metadata.Name,
		"size", snapshot.Size,
		"created_by", metadata.CreatedBy,
	).Info("Snapshot created successfully")

	return snapshot, nil
}

func (s *SnapshotService) GetSnapshot(ctx context.Context, id string) (*domain.Snapshot, error) {
	snapshot, err := s.storage.GetSnapshot(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}
	return snapshot, nil
}

func (s *SnapshotService) ListSnapshots(ctx context.Context, limit, offset int) ([]*domain.Snapshot, error) {
	snapshots, err := s.storage.ListSnapshots(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	return snapshots, nil
}

func (s *SnapshotService) DeleteSnapshot(ctx context.Context, id string) error {
	if err := s.storage.DeleteSnapshot(ctx, id); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	log.With("id", id).Info("Snapshot deleted successfully")
	return nil
}

func (s *SnapshotService) GetSnapshotContent(ctx context.Context, id string) ([]byte, error) {
	content, err := s.storage.GetSnapshotContent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot content: %w", err)
	}
	return content, nil
}

// generateSnapshotID creates a unique snapshot ID
func generateSnapshotID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "snapshot-" + hex.EncodeToString(bytes), nil
}
