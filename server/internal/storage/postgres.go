package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/3nd3r1/kubin/server/internal/config"
	"github.com/3nd3r1/kubin/server/internal/domain"
	"github.com/3nd3r1/kubin/server/internal/log"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	cfg := config.Get()
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) StoreSnapshot(ctx context.Context, snapshot *domain.Snapshot, content []byte) error {
	// Store metadata in PostgreSQL
	metadataJSON, err := json.Marshal(snapshot.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO snapshots (id, metadata, size, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			metadata = EXCLUDED.metadata,
			size = EXCLUDED.size,
			updated_at = EXCLUDED.updated_at
	`

	_, err = p.db.ExecContext(ctx, query,
		snapshot.ID,
		metadataJSON,
		snapshot.Size,
		snapshot.CreatedAt,
		snapshot.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store snapshot metadata: %w", err)
	}

	// Store content in file storage (this would be implemented separately)
	// For now, we'll just log that we would store the content
	log.With("id", snapshot.ID, "size", snapshot.Size).Debug("Would store snapshot content")

	return nil
}

func (p *PostgresStorage) GetSnapshot(ctx context.Context, id string) (*domain.Snapshot, error) {
	query := `
		SELECT id, metadata, size, created_at, updated_at
		FROM snapshots
		WHERE id = $1
	`

	var snapshot domain.Snapshot
	var metadataJSON []byte

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&snapshot.ID,
		&metadataJSON,
		&snapshot.Size,
		&snapshot.CreatedAt,
		&snapshot.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("snapshot not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Unmarshal metadata
	var metadata domain.SnapshotMetadata
	if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	snapshot.Metadata = &metadata

	return &snapshot, nil
}

func (p *PostgresStorage) ListSnapshots(ctx context.Context, limit, offset int) ([]*domain.Snapshot, error) {
	query := `
		SELECT id, metadata, size, created_at, updated_at
		FROM snapshots
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := p.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*domain.Snapshot
	for rows.Next() {
		var snapshot domain.Snapshot
		var metadataJSON []byte

		err := rows.Scan(
			&snapshot.ID,
			&metadataJSON,
			&snapshot.Size,
			&snapshot.CreatedAt,
			&snapshot.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}

		// Unmarshal metadata
		var metadata domain.SnapshotMetadata
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		snapshot.Metadata = &metadata

		snapshots = append(snapshots, &snapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating snapshots: %w", err)
	}

	return snapshots, nil
}

func (p *PostgresStorage) DeleteSnapshot(ctx context.Context, id string) error {
	query := `DELETE FROM snapshots WHERE id = $1`

	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("snapshot not found: %s", id)
	}

	return nil
}

func (p *PostgresStorage) GetSnapshotContent(ctx context.Context, id string) ([]byte, error) {
	// This would retrieve content from file storage
	// For now, return a placeholder
	return []byte("Snapshot content placeholder"), nil
}

func (p *PostgresStorage) Health(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}
