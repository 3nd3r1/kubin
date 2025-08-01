# Kubin Architecture

## Overview

Kubin is a modular, Kubernetes-native snapshot sharing platform with three main components: CLI, Server, and UI. The architecture prioritizes modularity, extensibility, and clean separation of concerns.

## System Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    CLI      │    │   Server    │    │     UI      │
│             │    │             │    │             │
│ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │
│ │Collector│ │    │ │   API   │ │    │ │  React  │ │
│ │Manager  │ │    │ │ Gateway │ │    │ │   App   │ │
│ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │
│ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │
│ │Snapshot │ │    │ │Snapshot │ │    │ │Snapshot │ │
│ │Manager  │ │    │ │Service  │ │    │ │Viewer   │ │
│ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │
└─────────────┘    │ ┌─────────┐ │    └─────────────┘
                   │ │Storage  │ │
                   │ │Service  │ │
                   │ └─────────┘ │
                   └─────────────┘
```

## Data Flow

### 1. Snapshot Creation Flow
```
CLI → K8s API → Collector → Snapshot Manager → Server API → Storage
```

### 2. Snapshot Viewing Flow
```
UI → Server API → Storage → Snapshot Service → UI Renderer
```

## Component Details

### CLI Architecture

```
CLI
├── Collector Interface
│   ├── PodCollector (MVP)
│   ├── LogCollector (MVP)
│   └── [Future: ServiceCollector, DeploymentCollector]
├── Snapshot Manager
│   ├── Data Packaging
│   ├── Compression
│   └── Upload Client
└── Configuration Manager
    ├── Kubeconfig
    ├── Server Settings
    └── Collection Filters
```

**Key Interfaces:**
```go
type Collector interface {
    Name() string
    Collect(ctx context.Context) ([]Resource, error)
}

type SnapshotManager interface {
    Create(collectors []Collector) (*Snapshot, error)
    Upload(snapshot *Snapshot) (string, error)
}
```

### Server Architecture

```
Server
├── API Gateway
│   ├── REST API (CLI endpoints)
│   ├── Internal API (UI endpoints)
│   └── Middleware (Auth, Logging, CORS)
├── Snapshot Service
│   ├── Metadata Management
│   ├── File Storage
│   └── Access Control
├── Storage Layer
│   ├── Database Interface
│   ├── File Storage Interface
│   └── Cache Interface
└── Plugin System
    ├── Auth Provider Interface
    ├── Storage Provider Interface
    └── Collector Interface
```

**Key Interfaces:**
```go
type StorageProvider interface {
    Store(snapshot *Snapshot) error
    Retrieve(id string) (*Snapshot, error)
    Delete(id string) error
}

type AuthProvider interface {
    Authenticate(token string) (*User, error)
    Authorize(user *User, resource string) bool
}
```

### UI Architecture

```
UI
├── Snapshot Viewer
│   ├── Resource Browser
│   ├── Pod Details
│   └── Log Viewer
├── Navigation
│   ├── Namespace Tree
│   ├── Resource List
│   └── Search
└── API Client
    ├── Snapshot API
    ├── Resource API
    └── Log API
```

## Data Models

### Snapshot Structure
```json
{
  "id": "snapshot-1234567890",
  "metadata": {
    "timestamp": "2024-01-01T12:00:00Z",
    "cluster": {
      "name": "prod-cluster",
      "version": "1.28.0",
      "context": "production"
    },
    "resources": {
      "pods": 25,
      "logs": 15
    },
    "filters": {
      "namespaces": ["default", "kube-system"],
      "labelSelectors": ["app=web"]
    }
  },
  "files": {
    "resources.tar.gz": "s3://bucket/snapshots/123/resources.tar.gz",
    "logs.tar.gz": "s3://bucket/snapshots/123/logs.tar.gz"
  }
}
```

### Resource Data Format
```json
{
  "kind": "Pod",
  "namespace": "default",
  "name": "web-app-123",
  "data": {
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": { ... },
    "spec": { ... },
    "status": { ... }
  },
  "logs": [
    {
      "container": "web",
      "logs": "2024-01-01T12:00:00Z INFO Starting application..."
    }
  ]
}
```

## API Design

### CLI API (External)
```
POST /api/v1/snapshots          # Upload snapshot
GET  /api/v1/snapshots/:id      # Get snapshot metadata
GET  /api/v1/snapshots/:id/download # Download snapshot
```

### UI API (Internal)
```
GET  /internal/api/v1/snapshots/:id/resources # Get resource list
GET  /internal/api/v1/snapshots/:id/pods      # Get pod details
GET  /internal/api/v1/snapshots/:id/logs      # Get pod logs
GET  /internal/api/v1/snapshots/:id/namespaces # Get namespace tree
```

## Storage Architecture

### Database Schema (PostgreSQL)
```sql
-- Snapshots table
CREATE TABLE snapshots (
    id VARCHAR(50) PRIMARY KEY,
    metadata JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    access_level VARCHAR(20) DEFAULT 'public'
);

-- Resources table (for fast querying)
CREATE TABLE resources (
    id SERIAL PRIMARY KEY,
    snapshot_id VARCHAR(50) REFERENCES snapshots(id),
    kind VARCHAR(50) NOT NULL,
    namespace VARCHAR(100),
    name VARCHAR(200) NOT NULL,
    data JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### File Storage
- **Local**: `/var/lib/kubin/snapshots/{id}/`
- **S3**: `s3://bucket/snapshots/{id}/`
- **Modular**: Interface allows switching storage backends

## Deployment Architecture

### Kubernetes Deployment
```yaml
# Server Deployment
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
        env:
        - name: KUBIN_DB_HOST
          value: "postgres-service"
        - name: KUBIN_STORAGE_TYPE
          value: "s3"

# UI Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubin-ui
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: kubin-ui
        image: kubin/ui:latest
        ports:
        - containerPort: 3000
```

### Helm Chart Structure
```
kubin/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── server-deployment.yaml
│   ├── ui-deployment.yaml
│   ├── postgres-deployment.yaml
│   ├── ingress.yaml
│   └── service.yaml
└── charts/
    └── postgresql/
```

## Modularity & Extensibility

### Plugin Architecture
```go
// Collector Plugin Interface
type CollectorPlugin interface {
    Name() string
    Version() string
    Collect(ctx context.Context, config map[string]interface{}) ([]Resource, error)
}

// Storage Plugin Interface
type StoragePlugin interface {
    Name() string
    Initialize(config map[string]interface{}) error
    Store(path string, data []byte) error
    Retrieve(path string) ([]byte, error)
    Delete(path string) error
}

// Auth Plugin Interface
type AuthPlugin interface {
    Name() string
    Authenticate(credentials interface{}) (*User, error)
    Authorize(user *User, resource string, action string) bool
}
```

### Configuration Management
```yaml
# Server Configuration
server:
  port: 8080
  host: "0.0.0.0"

storage:
  type: "s3"  # or "local", "gcs"
  config:
    bucket: "kubin-snapshots"
    region: "us-west-2"

auth:
  type: "none"  # or "jwt", "oauth", "ldap"
  config: {}

collectors:
  - name: "pods"
    enabled: true
  - name: "logs"
    enabled: true
  - name: "services"
    enabled: false
```

## Security Considerations

### MVP (No Auth)
- Public snapshots by default
- No authentication required
- Rate limiting on API endpoints

### Future Extensions
- JWT-based authentication
- Role-based access control
- Organization-based isolation
- API key management
- Audit logging

## Performance Considerations

### Caching Strategy
- Snapshot metadata in Redis
- Resource lists cached in memory
- Log data streamed on-demand

### Scaling Strategy
- Horizontal scaling of server instances
- Database connection pooling
- CDN for UI static assets
- Load balancing for API requests

## Monitoring & Observability

### Metrics
- Snapshot creation rate
- API response times
- Storage usage
- Error rates

### Logging
- Structured logging with correlation IDs
- Request/response logging
- Error tracking and alerting

### Health Checks
- Database connectivity
- Storage availability
- External service dependencies 