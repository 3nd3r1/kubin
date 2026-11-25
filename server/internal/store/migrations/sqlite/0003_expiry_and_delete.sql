-- Snapshots expire after expires_at (NULL means never, for snapshots stored
-- before this migration). delete_token_hash holds the lowercase hex SHA-256 of
-- the bearer token returned at upload.
ALTER TABLE snapshots ADD COLUMN expires_at TIMESTAMP;
ALTER TABLE snapshots ADD COLUMN delete_token_hash TEXT NOT NULL DEFAULT '';

CREATE INDEX snapshots_by_expires_at ON snapshots (expires_at) WHERE expires_at IS NOT NULL;
