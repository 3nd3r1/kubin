package domain

import "time"

type SnapshotMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ClusterName string `json:"cluster_name"`
	CreatedBy   string `json:"created_by"`
}

type Snapshot struct {
	ID        string            `json:"id"`
	Metadata  *SnapshotMetadata `json:"metadata"`
	Size      int64             `json:"size"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
