# Kubin Development Guide

## Quick Start

```bash
# 1. Start development environment
make dev

# 2. Install UI dependencies
make ui-install

# 3. Start server (in new terminal)
make server-run

# 4. Start UI (in new terminal)
make ui-dev
```

## Development Environment

The development environment includes:

- **PostgreSQL** (localhost:5432) - Database for metadata
- **Redis** (localhost:6379) - Caching layer
- **MinIO** (localhost:9000) - S3-compatible storage
- **pgAdmin** (localhost:5050) - Database management (optional)

## Makefile Commands

### Environment Management
```bash
make dev          # Start all services
make dev-down     # Stop all services
make dev-logs     # View service logs
make dev-reset    # Reset environment
make dev-status   # Check service status
```

### Database Management
```bash
make db           # Ensure database is running
make db-reset     # Reset database
make db-backup    # Create backup
make db-restore   # Restore from backup
```

### Storage Management
```bash
make storage      # Ensure storage is running
make storage-reset # Reset MinIO storage
make storage-ls   # List storage contents
```

### Development Tools
```bash
make tools        # Start pgAdmin
make tools-down   # Stop tools
```

### Component Development
```bash
# Server
make server-run   # Run server
make server-build # Build server
make server-test  # Test server

# CLI
make cli-build    # Build CLI
make cli-test     # Test CLI

# UI
make ui-dev       # Start UI dev server
make ui-build     # Build UI
make ui-test      # Test UI
```

### Full Development Setup
```bash
make full-dev     # Complete setup
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Server Configuration
KUBIN_SERVER_PORT=8080
KUBIN_SERVER_HOST=0.0.0.0

# Database Configuration
KUBIN_DB_TYPE=postgres
KUBIN_DB_HOST=localhost
KUBIN_DB_PORT=5432
KUBIN_DB_NAME=kubin
KUBIN_DB_USER=kubin
KUBIN_DB_PASSWORD=kubin
KUBIN_DB_SSL_MODE=disable

# Redis Configuration
KUBIN_REDIS_HOST=localhost
KUBIN_REDIS_PORT=6379

# Storage Configuration
KUBIN_STORAGE_TYPE=s3
KUBIN_STORAGE_S3_ENDPOINT=localhost:9000
KUBIN_STORAGE_S3_ACCESS_KEY=minioadmin
KUBIN_STORAGE_S3_SECRET_KEY=minioadmin
KUBIN_STORAGE_S3_BUCKET=kubin-snapshots
KUBIN_STORAGE_S3_REGION=us-east-1
KUBIN_STORAGE_S3_USE_SSL=false

# Authentication (MVP: none)
KUBIN_AUTH_TYPE=none

# UI Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Access Points

After starting the environment:

- **PostgreSQL**: `localhost:5432` (kubin/kubin)
- **Redis**: `localhost:6379`
- **MinIO API**: `localhost:9000`
- **MinIO Console**: `localhost:9001` (minioadmin/minioadmin)
- **pgAdmin**: `localhost:5050` (admin@kubin.local/admin)
- **UI**: `http://localhost:3000`
- **Server**: `http://localhost:8080`

## Development Workflow

1. **Start Environment**: `make dev`
2. **Start Server**: `make server-run`
3. **Start UI**: `make ui-dev`
4. **Test CLI**: `make cli-build`
5. **View Logs**: `make dev-logs`

## Troubleshooting

### Reset Everything
```bash
make clean        # Stop and remove everything
make full-dev     # Start fresh
```

### Database Issues
```bash
make db-reset     # Reset database
```

### Storage Issues
```bash
make storage-reset # Reset MinIO
```

### Service Status
```bash
make dev-status   # Check all services
``` 