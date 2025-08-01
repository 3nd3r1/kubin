-- Kubin Database Initialization Script

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Snapshots table
CREATE TABLE IF NOT EXISTS snapshots (
    id VARCHAR(50) PRIMARY KEY,
    metadata JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    access_level VARCHAR(20) DEFAULT 'public',
    created_by VARCHAR(100),
    tags TEXT[]
);

-- Resources table for fast querying
CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    snapshot_id VARCHAR(50) REFERENCES snapshots(id) ON DELETE CASCADE,
    kind VARCHAR(50) NOT NULL,
    namespace VARCHAR(100),
    name VARCHAR(200) NOT NULL,
    data JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_snapshots_created_at ON snapshots(created_at);
CREATE INDEX IF NOT EXISTS idx_snapshots_access_level ON snapshots(access_level);
CREATE INDEX IF NOT EXISTS idx_resources_snapshot_id ON resources(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_resources_kind ON resources(kind);
CREATE INDEX IF NOT EXISTS idx_resources_namespace ON resources(namespace);
CREATE INDEX IF NOT EXISTS idx_resources_name ON resources(name);

-- Create GIN indexes for JSONB fields
CREATE INDEX IF NOT EXISTS idx_snapshots_metadata ON snapshots USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_resources_data ON resources USING GIN (data);

-- Insert some sample data for development
INSERT INTO snapshots (id, metadata, access_level, created_by, tags) VALUES
(
    'sample-snapshot-1',
    '{
        "timestamp": "2024-01-01T12:00:00Z",
        "cluster": {
            "name": "dev-cluster",
            "version": "1.28.0",
            "context": "development"
        },
        "resources": {
            "pods": 5,
            "logs": 3
        },
        "filters": {
            "namespaces": ["default", "kube-system"],
            "labelSelectors": ["app=web"]
        }
    }'::jsonb,
    'public',
    'developer@example.com',
    ARRAY['development', 'sample']
) ON CONFLICT (id) DO NOTHING;

-- Sample resources
INSERT INTO resources (snapshot_id, kind, namespace, name, data) VALUES
(
    'sample-snapshot-1',
    'Pod',
    'default',
    'web-app-123',
    '{
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
            "name": "web-app-123",
            "namespace": "default"
        },
        "spec": {
            "containers": [
                {
                    "name": "web",
                    "image": "nginx:latest"
                }
            ]
        },
        "status": {
            "phase": "Running"
        }
    }'::jsonb
) ON CONFLICT DO NOTHING; 