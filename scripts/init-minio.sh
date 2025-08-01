#!/bin/sh
set -e

echo "Initializing MinIO..."

# Configure MinIO client
mc alias set local http://minio:9000 ${MINIO_ROOT_USER:-minioadmin} ${MINIO_ROOT_PASSWORD:-minioadmin}

# Create bucket for Kubin snapshots
mc mb local/kubin-snapshots --ignore-existing || true

# Set bucket policy for public read (for development)
mc anonymous set download local/kubin-snapshots

echo "Verifying setup:"
mc ls local/kubin-snapshots
mc anonymous get local/kubin-snapshots

echo "MinIO initialization completed!"
echo "Access MinIO Console at: http://localhost:9001"
echo "Login with: minioadmin / minioadmin" 
