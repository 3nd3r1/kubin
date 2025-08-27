# Kubin Server

Store and serve Kubernetes cluster snapshots.

## Run

```bash
# Development
go run cmd/server/main.go

# Production
go build -o kubin-server
./kubin-server
```

## What it does

- Stores uploaded cluster snapshots
- Provides REST API for CLI and UI
- Handles user authentication
- Serves snapshot data to web UI

## API

```
POST /api/v1/snapshots     # Upload snapshot
GET  /api/v1/snapshots     # List snapshots
GET  /api/v1/snapshots/:id # Get snapshot
```

## Configuration

Environment variables:
```bash
KUBIN_SERVER_PORT=8080
KUBIN_DB_HOST=localhost
KUBIN_STORAGE_PATH=/var/lib/kubin
```

## Deploy

```yaml
# Kubernetes deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubin-server
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: kubin-server
        image: kubin/server:latest
        ports:
        - containerPort: 8080
```

## Build

```bash
go build -o kubin-server cmd/server/main.go
``` 