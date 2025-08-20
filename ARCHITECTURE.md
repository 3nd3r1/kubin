# Kubin Architecture Documentation

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Principles](#architecture-principles)
3. [Service Architecture](#service-architecture)
4. [Data Architecture](#data-architecture)
5. [Communication Patterns](#communication-patterns)
6. [Upload Flow](#upload-flow)
7. [Query Flow](#query-flow)
8. [Performance Optimizations](#performance-optimizations)
9. [Operational Considerations](#operational-considerations)
10. [Scaling Strategy](#scaling-strategy)

## System Overview

Kubin is a Kubernetes snapshot sharing platform that enables development teams to capture, share, and analyze historical cluster states. The platform consists of a CLI tool for uploading cluster snapshots and a web interface for querying and visualizing cluster data.

### Core Requirements
- **Immediate Response**: CLI users receive shareable URLs within 2 seconds
- **Sub-second Queries**: Web UI loads snapshot data in under 2 seconds
- **High Throughput**: Support 100+ concurrent uploads of 100MB-1GB snapshots
- **Strong Consistency**: No eventual consistency - shared URLs work immediately
- **Rich Analytics**: Query billions of log entries with complex filtering

### Key Capabilities
- Historical cluster state preservation
- Cross-team snapshot sharing via URLs
- Advanced log analytics and search
- Lens-style Kubernetes object browsing
- Resource usage trend analysis

## Architecture Principles

### 1. Separation of Upload and Query Paths
The architecture maintains distinct optimized paths for data ingestion (upload) and data retrieval (query), preventing mutual interference and enabling independent scaling.

### 2. Immediate Response via Async Processing
CLI uploads receive immediate URL responses while heavy processing occurs asynchronously, satisfying user expectations without blocking operations.

### 3. Data Tier Optimization
Different data types are stored in purpose-built systems:
- Metadata: PostgreSQL for ACID transactions and complex queries
- Large files: S3 for durability and direct access
- Analytics: ClickHouse for columnar analytics at scale
- Caching: Redis for sub-second response times

### 4. Service Boundary Definition
Each service owns a specific data domain and business capability, following the single responsibility principle with clear interfaces.

### 5. Complexity Isolation
Complex orchestration logic is contained within dedicated services (Upload Orchestrator, Query Service) while keeping other services focused and simple.

## Service Architecture

### API Gateway (Kong/Envoy)
**Purpose**: Single entry point for all external traffic with cross-cutting concerns.

**Responsibilities**:
- TLS termination and SSL management
- Authentication and authorization (JWT validation)
- Rate limiting (100 concurrent uploads, 1000 queries/minute)
- Request routing based on path patterns
- Request/response logging and metrics collection
- CORS handling for web UI

**Technology Stack**: Kong Gateway with Redis for rate limiting state

**Scaling Strategy**: Horizontal scaling behind load balancer with session affinity disabled

### Upload Orchestrator Service
**Purpose**: Coordinates the complex multi-step upload workflow while providing immediate responses.

**Core Responsibilities**:
- Generate unique snapshot IDs (UUID v4)
- Create initial metadata records with "uploading" status
- Coordinate pre-signed URL generation across file types
- Return immediate shareable URLs to CLI clients
- Monitor upload completion via S3 events
- Orchestrate async processing pipeline
- Handle partial failures and retry logic
- Manage upload timeouts and cleanup

**State Management**:
```golang
type UploadSession struct {
    SnapshotID    string
    UserID        string
    Status        UploadStatus
    CreatedAt     time.Time
    ExpiresAt     time.Time
    FileManifest  []FileDescriptor
    UploadURLs    map[string]string
    ProcessingState ProcessingStep
}
```

**Workflow States**:
1. `INITIATED` - Snapshot ID generated, URLs created
2. `UPLOADING` - Files being uploaded to S3
3. `PROCESSING` - Background parsing and indexing
4. `COMPLETED` - Ready for queries
5. `FAILED` - Error state with cleanup

**Technology Stack**: Go with Gin framework, PostgreSQL for state persistence, Kafka for event publishing

### Metadata Service
**Purpose**: Fast storage and retrieval of Kubernetes object metadata for listing and filtering.

**Data Model**:
```sql
-- Primary metadata table
CREATE TABLE k8s_resources (
    id BIGSERIAL PRIMARY KEY,
    snapshot_id UUID NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    namespace VARCHAR(100),
    name VARCHAR(200) NOT NULL,
    uid VARCHAR(36),
    
    -- Common fields
    created_at TIMESTAMP NOT NULL,
    labels JSONB,
    annotations JSONB,
    status VARCHAR(50),
    
    -- Pod-specific fields
    node_name VARCHAR(200),
    restart_count INTEGER,
    phase VARCHAR(50),
    container_names TEXT[],
    
    -- Service-specific fields
    service_type VARCHAR(50),
    cluster_ip INET,
    ports JSONB,
    
    -- Deployment-specific fields
    replicas INTEGER,
    ready_replicas INTEGER,
    strategy_type VARCHAR(50),
    
    -- Reference to complete object
    s3_object_key VARCHAR(500) NOT NULL,
    object_size_bytes BIGINT,
    
    -- Indexing
    created_at_idx TIMESTAMP DEFAULT NOW()
);

-- Optimized indexes
CREATE INDEX idx_resources_snapshot_ns_type ON k8s_resources(snapshot_id, namespace, resource_type);
CREATE INDEX idx_resources_labels ON k8s_resources USING GIN(labels);
CREATE INDEX idx_resources_name_search ON k8s_resources(name text_pattern_ops);
CREATE INDEX idx_resources_status ON k8s_resources(resource_type, status) WHERE status != 'Running';
```

**Query Patterns**:
- Namespace listing: O(log n) with snapshot_id + namespace index
- Cross-resource search: O(log n) with label GIN index
- Status filtering: Partial index for non-Running resources
- Text search: Pattern matching with trigram indexes

**Caching Strategy**:
- Redis cache for frequently accessed namespaces (5-minute TTL)
- Application-level caching for search results (1-minute TTL)
- Prepared statement caching for common queries

**Technology Stack**: PostgreSQL 14+, Redis, Go with SQLX

### Storage Service
**Purpose**: Manages all interactions with object storage for large file operations.

**Responsibilities**:
- Generate pre-signed URLs for direct S3 uploads
- Organize files with consistent key patterns
- Handle multipart upload coordination for large files
- Generate download URLs for web UI access
- Implement lifecycle policies for cost optimization
- Monitor storage metrics and quotas

**File Organization**:
```
s3://kubin-snapshots/
├── snapshots/{snapshot_id}/
│   ├── metadata/
│   │   ├── cluster-info.json
│   │   ├── nodes.json
│   │   └── namespaces.json
│   ├── resources/
│   │   ├── pods/{namespace}/{pod-name}.yaml
│   │   ├── services/{namespace}/{service-name}.yaml
│   │   ├── deployments/{namespace}/{deployment-name}.yaml
│   │   └── configmaps/{namespace}/{cm-name}.yaml
│   └── logs/
│       ├── pods/{namespace}/{pod-name}/
│       │   ├── container1.log
│       │   └── container2.log
│       └── events/
│           └── events.json
```

**Upload Coordination**:
```golang
type UploadManifest struct {
    SnapshotID string
    Files []FileDescriptor
}

type FileDescriptor struct {
    Type        FileType  // metadata, resource, log
    Category    string    // pods, services, events
    Namespace   string
    Name        string
    SizeBytes   int64
    Checksum    string
}
```

**Technology Stack**: AWS S3, Go AWS SDK v2, Redis for URL caching

### Log Processing Service
**Purpose**: Transforms raw Kubernetes logs into searchable, indexed data for analytics.

**Processing Pipeline**:
1. **Raw Log Ingestion**: Consume S3 upload events from Kafka
2. **Log Parsing**: Extract structured data from various log formats
3. **Data Enrichment**: Add metadata context (pod, namespace, node)
4. **Indexing**: Insert parsed data into ClickHouse with optimizations
5. **Cleanup**: Remove temporary processing files

**Log Format Support**:
- Container logs (stdout/stderr)
- Kubernetes events
- Audit logs
- Custom application logs with JSON structure

**Data Transformation**:
```golang
type LogEntry struct {
    SnapshotID   string    `ch:"snapshot_id"`
    Timestamp    time.Time `ch:"timestamp"`
    Namespace    string    `ch:"namespace"`
    PodName      string    `ch:"pod_name"`
    ContainerName string   `ch:"container_name"`
    NodeName     string    `ch:"node_name"`
    LogLevel     string    `ch:"log_level"`
    Message      string    `ch:"message"`
    Labels       map[string]string `ch:"labels"`
    
    // Analytics fields
    MessageLength int32  `ch:"message_length"`
    ErrorKeywords []string `ch:"error_keywords"`
    RequestID     string `ch:"request_id"`
}
```

**Performance Optimizations**:
- Parallel processing of log files using worker pools
- Batch inserts to ClickHouse (10,000 rows per batch)
- Compression during transport (LZ4)
- Memory-mapped file reading for large logs

**Technology Stack**: Go, ClickHouse, Kafka, Prometheus for metrics

### Query Service
**Purpose**: Orchestrates complex queries across multiple data sources to provide unified responses.

**Core Capabilities**:
- Aggregate data from Metadata, Storage, and Analytics services
- Implement intelligent caching strategies
- Optimize response times through parallel requests
- Handle complex search and filtering operations
- Provide GraphQL-like field selection

**Request Processing**:
```golang
type SnapshotQuery struct {
    SnapshotID string
    Namespace  string
    ResourceTypes []string
    LabelSelector map[string]string
    TimeRange    *TimeRange
    IncludeFields []string
}

type QueryResponse struct {
    Metadata    ResourceMetadata
    Resources   []K8sResource
    LogSummary  LogAnalytics
    FileURLs    map[string]string
    CacheInfo   CacheMetadata
}
```

**Caching Layers**:
1. **Request Cache**: Complete responses cached for 15 minutes
2. **Component Cache**: Individual service responses cached for 5 minutes
3. **Search Cache**: Complex queries cached for 1 hour
4. **CDN Cache**: Static content cached globally

**Parallel Execution**:
```golang
func (q *QueryService) GetSnapshot(ctx context.Context, req *SnapshotQuery) (*QueryResponse, error) {
    var wg sync.WaitGroup
    var mu sync.Mutex
    resp := &QueryResponse{}
    
    // Parallel service calls
    wg.Add(3)
    
    // Metadata service
    go func() {
        defer wg.Done()
        metadata := q.metadataService.GetResources(ctx, req)
        mu.Lock()
        resp.Metadata = metadata
        mu.Unlock()
    }()
    
    // Storage service
    go func() {
        defer wg.Done()
        urls := q.storageService.GetDownloadURLs(ctx, req.SnapshotID)
        mu.Lock()
        resp.FileURLs = urls
        mu.Unlock()
    }()
    
    // Analytics service
    go func() {
        defer wg.Done()
        logs := q.analyticsService.GetLogSummary(ctx, req)
        mu.Lock()
        resp.LogSummary = logs
        mu.Unlock()
    }()
    
    wg.Wait()
    return resp, nil
}
```

**Technology Stack**: Go, Redis, HTTP/2 for service communication

### Analytics Service
**Purpose**: Provides advanced analytics and search capabilities over historical log data.

**Query Capabilities**:
- Full-text search across log messages
- Time-series analysis of error rates
- Resource usage trending
- Cross-pod correlation analysis
- Custom metric aggregations

**ClickHouse Schema Design**:
```sql
-- Main logs table with optimal partitioning
CREATE TABLE kubin_logs (
    snapshot_id String,
    timestamp DateTime64(3),
    namespace String,
    pod_name String,
    container_name String,
    node_name String,
    log_level Enum8('DEBUG'=1, 'INFO'=2, 'WARN'=3, 'ERROR'=4, 'FATAL'=5),
    message String,
    labels Map(String, String),
    
    -- Analytics columns
    message_length UInt32,
    error_keywords Array(String),
    request_id String,
    
    -- Partitioning
    date Date MATERIALIZED toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY (snapshot_id, toYYYYMM(timestamp))
ORDER BY (snapshot_id, namespace, pod_name, timestamp)
SETTINGS index_granularity = 8192;

-- Materialized views for common aggregations
CREATE MATERIALIZED VIEW error_summary
ENGINE = SummingMergeTree()
ORDER BY (snapshot_id, namespace, pod_name, date)
AS SELECT
    snapshot_id,
    namespace,
    pod_name,
    toDate(timestamp) as date,
    count() as total_logs,
    countIf(log_level >= 3) as error_count,
    countIf(log_level = 4) as fatal_count
FROM kubin_logs
GROUP BY snapshot_id, namespace, pod_name, date;
```

**Query Optimization**:
- Partition pruning by snapshot_id for query isolation
- Primary key optimization for time-range queries
- Materialized views for common aggregations
- Compression optimization (LZ4 for speed, ZSTD for storage)

**Technology Stack**: ClickHouse, Go, Prometheus

## Data Architecture

### Data Flow Overview
```
CLI Upload → S3 (Raw Files) → Kafka Events → Processing Services
                ↓
          Metadata Service → PostgreSQL (Structured Data)
                ↓
          Log Processing → ClickHouse (Analytics Data)
                ↓
          Query Service → Redis Cache → Web UI
```

### Data Consistency Model

**Strong Consistency Requirements**:
- Snapshot metadata must be immediately queryable after URL generation
- File availability must be guaranteed when download URLs are provided
- Cross-service data must remain synchronized

**Implementation**:
- PostgreSQL ACID transactions for metadata operations
- S3 strong read-after-write consistency
- Event-driven synchronization via Kafka
- Distributed transaction patterns for critical paths

### Data Retention Strategy

**Hot Data (0-30 days)**:
- Full metadata in PostgreSQL with all indexes
- Complete logs in ClickHouse with full search capability
- S3 objects in standard storage class

**Warm Data (30-365 days)**:
- Metadata retained with reduced indexing
- Log aggregations preserved, raw logs archived
- S3 objects moved to Intelligent-Tiering

**Cold Data (365+ days)**:
- Metadata summaries only
- Aggregated metrics preserved
- S3 objects in Glacier storage class

## Communication Patterns

### Synchronous Communication (HTTP/gRPC)
**Use Cases**:
- CLI to Upload Orchestrator (immediate response required)
- Web UI to Query Service (real-time user interaction)
- Service-to-service calls for request fulfillment

**Protocol Selection**:
- External APIs: REST/HTTP for simplicity and tooling support
- Internal services: gRPC for performance and type safety
- Load balancing: Client-side with circuit breakers

### Asynchronous Communication (Kafka)
**Use Cases**:
- Upload completion notifications
- Background processing coordination
- Cross-service event propagation
- Audit logging and metrics

**Topic Design**:
```
kubin.uploads.completed - S3 upload notifications
kubin.processing.status - Processing pipeline events  
kubin.analytics.requests - Log processing jobs
kubin.system.metrics - Operational metrics
```

**Event Schema**:
```json
{
  "event_type": "upload_completed",
  "timestamp": "2025-08-20T10:30:00Z",
  "snapshot_id": "abc123",
  "payload": {
    "file_count": 42,
    "total_size_bytes": 1073741824,
    "upload_duration_seconds": 45
  },
  "metadata": {
    "user_id": "user123",
    "cli_version": "v1.2.3"
  }
}
```

## Upload Flow

### Phase 1: Immediate Response (< 2 seconds)
```
1. CLI → API Gateway → Upload Orchestrator
   Request: POST /api/v1/uploads
   Body: {
     "cluster_metadata": {...},
     "file_manifest": [...]
   }

2. Upload Orchestrator Processing:
   - Generate UUID snapshot_id
   - Create PostgreSQL metadata record (status: "uploading")
   - Request pre-signed URLs from Storage Service
   - Return response immediately

3. Response to CLI:
   {
     "snapshot_id": "abc123",
     "share_url": "https://kubin.io/snapshots/abc123",
     "upload_urls": {
       "metadata": "https://s3.../metadata?signature=...",
       "resources": "https://s3.../resources?signature=...",
       "logs": "https://s3.../logs?signature=..."
     },
     "expires_at": "2025-08-20T11:30:00Z"
   }
```

### Phase 2: Parallel Upload (5-120 seconds)
```
CLI performs parallel uploads directly to S3:
- Metadata files (JSON): ~1MB, completes in 1-2 seconds
- Resource files (YAML): ~10-50MB, completes in 10-30 seconds  
- Log files: ~50MB-1GB, completes in 30-120 seconds

Upload optimizations:
- Multipart upload for files >100MB
- Parallel uploads with connection pooling
- Retry logic with exponential backoff
- Progress reporting to Upload Orchestrator
```

### Phase 3: Async Processing (30-300 seconds)
```
1. S3 Event → Kafka → Upload Orchestrator
   Event: "upload_completed" with file inventory

2. Upload Orchestrator → Processing Pipeline:
   - Update metadata status: "uploading" → "processing"
   - Trigger Log Processing Service via Kafka
   - Monitor processing completion

3. Log Processing Service:
   - Download log files from S3
   - Parse and structure log data
   - Batch insert to ClickHouse
   - Update processing status

4. Completion:
   - Metadata status: "processing" → "completed"
   - Query Service cache warming
   - Notification to monitoring systems
```

## Query Flow

### Snapshot List View
```
Web UI → API Gateway → Query Service

1. Cache Check:
   Redis key: "snapshots:user123:page1"
   TTL: 5 minutes

2. Cache Miss → Metadata Service:
   SQL: SELECT snapshot_id, cluster_name, created_at, status
        FROM snapshots 
        WHERE user_id = 'user123'
        ORDER BY created_at DESC
        LIMIT 20 OFFSET 0

3. Response Enhancement:
   - Add thumbnail generation status
   - Include sharing permissions
   - Calculate relative timestamps

4. Cache and Return:
   Store in Redis with 5-minute TTL
   Return paginated response
```

### Individual Snapshot View
```
Web UI → API Gateway → Query Service

1. Parallel Data Fetching:
   
   Thread 1: Metadata Service
   - Basic snapshot information
   - Resource counts by type
   - Processing status
   
   Thread 2: Storage Service  
   - Generate download URLs for YAML files
   - Check file availability
   - Calculate storage metrics
   
   Thread 3: Analytics Service
   - Recent error summary
   - Log volume statistics
   - Resource usage trends

2. Data Aggregation:
   Wait for all threads (timeout: 1.5 seconds)
   Combine responses into unified view
   Apply user permissions filtering

3. Response Caching:
   Redis key: "snapshot:abc123:full"
   TTL: 15 minutes (longer for historical data)
```

### Kubernetes Object Browsing (Lens-Style)
```
Web UI: Namespace View → Query Service

1. Fast Metadata Query:
   PostgreSQL: SELECT resource_type, name, status, labels
               FROM k8s_resources
               WHERE snapshot_id = 'abc123' 
                 AND namespace = 'production'
               ORDER BY resource_type, name

2. Response Processing:
   - Group by resource type
   - Apply label-based filtering
   - Add status indicators
   - Generate drill-down URLs

3. Progressive Enhancement:
   Background requests for:
   - Resource relationship mapping
   - Recent event summaries
   - Resource usage metrics
```

### Detailed Object View (kubectl describe equivalent)
```
Web UI: Pod Detail → Query Service → Storage Service

1. Storage Service:
   Generate pre-signed S3 URL:
   s3://kubin-snapshots/abc123/resources/pods/production/web-app-xyz.yaml

2. Direct Download:
   Web UI downloads YAML directly from S3
   Bypasses application servers entirely
   CDN caching for repeated access

3. Enhancement Data:
   Analytics Service provides:
   - Recent log entries for this pod
   - Error frequency analysis
   - Resource usage history
```

## Performance Optimizations

### Database Optimizations

**PostgreSQL Tuning**:
```sql
-- Connection pooling
max_connections = 200
shared_buffers = 4GB
effective_cache_size = 12GB

-- Query optimization
random_page_cost = 1.1  -- SSD optimization
effective_io_concurrency = 200

-- Indexing strategy
CREATE INDEX CONCURRENTLY idx_resources_composite 
ON k8s_resources(snapshot_id, namespace, resource_type, status);

-- Partial indexes for common filters
CREATE INDEX idx_failed_resources 
ON k8s_resources(snapshot_id, namespace)
WHERE status != 'Running';
```

**ClickHouse Optimizations**:
```sql
-- Table engine tuning
ENGINE = MergeTree()
ORDER BY (snapshot_id, namespace, pod_name, timestamp)
PARTITION BY (snapshot_id, toYYYYMM(timestamp))
SETTINGS index_granularity = 8192,
         merge_max_block_size = 8192;

-- Compression settings
ALTER TABLE kubin_logs 
MODIFY SETTING compress_marks = 1,
               compress_primary_key = 1;

-- Materialized views for aggregations
CREATE MATERIALIZED VIEW hourly_error_rates
ENGINE = SummingMergeTree()
ORDER BY (snapshot_id, namespace, hour)
AS SELECT
    snapshot_id,
    namespace, 
    toStartOfHour(timestamp) as hour,
    count() as total_logs,
    countIf(log_level >= 3) as errors
FROM kubin_logs
GROUP BY snapshot_id, namespace, hour;
```

### Caching Strategy

**Multi-Level Cache Hierarchy**:
```
Level 1: CDN (Global)
- Static assets: 24 hours
- Download URLs: 1 hour
- API responses: 5 minutes

Level 2: Redis (Regional) 
- Query results: 15 minutes
- Metadata summaries: 5 minutes
- User sessions: 24 hours

Level 3: Application (Local)
- Database connections: Connection pooling
- Prepared statements: In-memory cache
- Service discovery: 30 seconds
```

**Cache Invalidation**:
```golang
// Event-driven cache invalidation
func (c *CacheManager) HandleSnapshotUpdate(event SnapshotEvent) {
    patterns := []string{
        fmt.Sprintf("snapshot:%s:*", event.SnapshotID),
        fmt.Sprintf("user:%s:snapshots:*", event.UserID),
        "global:stats:*",
    }
    
    for _, pattern := range patterns {
        c.redis.DeletePattern(pattern)
    }
}
```

### Connection Pooling and Circuit Breakers

**Database Connection Management**:
```golang
// PostgreSQL connection pool
config := pgxpool.Config{
    MaxConns:        30,
    MinConns:        5,
    MaxConnLifetime: time.Hour,
    MaxConnIdleTime: time.Minute * 30,
}

// ClickHouse connection pool
clickhouseConn := clickhouse.OpenDB(&clickhouse.Options{
    Addr: []string{"clickhouse:9000"},
    Settings: clickhouse.Settings{
        "max_execution_time": 60,
        "max_memory_usage":   "10000000000",
    },
    MaxOpenConns: 20,
    MaxIdleConns: 5,
})
```

**Circuit Breaker Implementation**:
```golang
// Hystrix-style circuit breaker
type CircuitBreaker struct {
    failureThreshold int
    resetTimeout     time.Duration
    state           atomic.Value // Open/Closed/HalfOpen
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
    if cb.state.Load().(State) == Open {
        return ErrCircuitBreakerOpen
    }
    
    err := operation()
    if err != nil {
        cb.recordFailure()
    } else {
        cb.recordSuccess()
    }
    return err
}
```

## Operational Considerations

### Monitoring and Observability

**Metrics Collection**:
```golang
// Prometheus metrics
var (
    uploadDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "kubin_upload_duration_seconds",
            Help: "Time taken for snapshot uploads",
            Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
        },
        []string{"user_id", "size_category"},
    )
    
    queryLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "kubin_query_duration_seconds", 
            Help: "Query response time",
            Buckets: []float64{0.1, 0.5, 1, 2, 5},
        },
        []string{"query_type", "cache_hit"},
    )
)
```

**Distributed Tracing**:
```golang
// OpenTelemetry integration
func (s *QueryService) GetSnapshot(ctx context.Context, req *SnapshotQuery) {
    ctx, span := s.tracer.Start(ctx, "QueryService.GetSnapshot")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("snapshot.id", req.SnapshotID),
        attribute.String("user.id", req.UserID),
    )
    
    // Trace service calls
    metadata := s.getMetadata(ctx, req)  // Creates child span
    storage := s.getStorageURLs(ctx, req) // Creates child span
    analytics := s.getAnalytics(ctx, req) // Creates child span
}
```

**Health Checks**:
```golang
// Service health endpoint
type HealthChecker struct {
    postgres   *pgxpool.Pool
    redis      *redis.Client
    clickhouse *sql.DB
    s3         *s3.Client
}

func (h *HealthChecker) Check(ctx context.Context) HealthStatus {
    checks := map[string]bool{
        "postgres":   h.checkPostgres(ctx),
        "redis":      h.checkRedis(ctx),
        "clickhouse": h.checkClickHouse(ctx),
        "s3":         h.checkS3(ctx),
    }
    
    overall := "healthy"
    for service, status := range checks {
        if !status {
            overall = "unhealthy"
            break
        }
    }
    
    return HealthStatus{
        Status: overall,
        Checks: checks,
        Timestamp: time.Now(),
    }
}
```

### Security Considerations

**Authentication and Authorization**:
```yaml
# JWT token structure
{
  "sub": "user123",
  "iss": "kubin-auth",
  "exp": 1692547200,
  "iat": 1692543600,
  "permissions": [
    "snapshots:read:own",
    "snapshots:write:own", 
    "snapshots:share:team"
  ],
  "team_id": "team456"
}
```

**Data Protection**:
- TLS 1.3 for all external communication
- mTLS for internal service communication
- Encryption at rest for sensitive data in PostgreSQL
- S3 bucket encryption with customer-managed keys
- Secrets management via Kubernetes secrets or HashiCorp Vault

**Access Control**:
```golang
// Resource-based access control
type AccessPolicy struct {
    Resource    string   `json:"resource"`    // snapshots:abc123
    Actions     []string `json:"actions"`     // read, write, share
    Conditions  []Condition `json:"conditions"` // team_id, time_range
}

func (a *Authorizer) CanAccess(userID, resource, action string) bool {
    policies := a.getUserPolicies(userID)
    for _, policy := range policies {
        if policy.Matches(resource, action) {
            return policy.Evaluate()
        }
    }
    return false
}
```

### Disaster Recovery

**Backup Strategy**:
- PostgreSQL: Continuous WAL archiving + daily base backups
- ClickHouse: Daily snapshots with incremental backups
- S3: Cross-region replication with versioning enabled
- Redis: RDB snapshots every 6 hours

**Recovery Procedures**:
```bash
# PostgreSQL point-in-time recovery
pg_basebackup -h backup-server -D /var/lib/postgresql/backup
pg_ctl start -D /var/lib/postgresql/backup

# ClickHouse restoration
clickhouse-backup restore --table kubin_logs 2025-08-20T00:00:00

# S3 cross-region failover  
aws s3 sync s3://kubin-snapshots-primary s3://kubin-snapshots-dr
```

## Scaling Strategy

### Horizontal Scaling Plan

**Service Scaling**:
```yaml
# Kubernetes HorizontalPodAutoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: query-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: query-service
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

**Database Scaling**:

**PostgreSQL**:
- Read replicas for query load distribution
- Connection pooling with PgBouncer
- Partitioning by snapshot creation date
- Archive old snapshots to reduce working set

**ClickHouse**:
- Cluster setup with distributed tables
- Sharding by snapshot_id for query isolation
- ReplicatedMergeTree for high availability
- Materialized views for query acceleration

**Redis**:
- Redis Cluster for horizontal scaling
- Consistent hashing for data distribution
- Separate clusters for different cache tiers

### Performance Targets

**Current Scale (MVP)**:
- 1,000 snapshots per day
- 10,000 queries per day
- 100GB storage per day
- 10 concurrent uploads

**Target Scale (Year 1)**:
- 10,000 snapshots per day
- 100,000 queries per day
- 1TB storage per day
- 100 concurrent uploads

**Enterprise Scale (Year 2+)**:
- 100,000 snapshots per day
- 1,000,000 queries per day
- 10TB storage per day
- 1,000 concurrent uploads

### Cost Optimization

**Storage Tiering**:
```golang
// Automated lifecycle management
type StoragePolicy struct {
    SnapshotAge   time.Duration
    StorageClass  string
    IndexLevel    IndexLevel
    CompressionType string
}

var storagePolicies = []StoragePolicy{
    {0, "STANDARD", FullIndex, "LZ4"},           // 0-30 days: Fast access
    {30*24*time.Hour, "STANDARD_IA", ReducedIndex, "ZSTD"}, // 30-90 days: Reduced cost
    {90*24*time.Hour, "GLACIER", MetadataOnly, "ZSTD"},     // 90+ days: Archive
}

func (s *StorageManager) ApplyLifecyclePolicy(snapshotID string, age time.Duration) {
    for _, policy := range storagePolicies {
        if age >= policy.SnapshotAge {
            s.transitionStorage(snapshotID, policy)
        }
    }
}
```

**Compute Cost Management**:
```yaml
# Spot instances for batch processing
apiVersion: v1
kind: Node
metadata:
  labels:
    node.kubernetes.io/instance-type: "spot"
    workload-type: "batch-processing"
spec:
  taints:
  - key: "spot-instance"
    value: "true"
    effect: "NoSchedule"

# Log processing on spot instances
apiVersion: apps/v1
kind: Deployment
metadata:
  name: log-processor
spec:
  template:
    spec:
      tolerations:
      - key: "spot-instance"
        operator: "Equal"
        value: "true"
        effect: "NoSchedule"
      nodeSelector:
        workload-type: "batch-processing"
```

## Implementation Roadmap

### Phase 1: MVP (Months 1-3)
**Goal**: Basic functionality with immediate URL sharing

**Components**:
- API Gateway (Kong)
- Upload Orchestrator Service
- Metadata Service with PostgreSQL
- Storage Service with S3 integration
- Basic Query Service
- Simple web UI for browsing

**Features**:
- CLI snapshot upload with immediate URL response
- Basic web UI for viewing snapshot lists
- Simple Kubernetes object browsing
- File download via pre-signed URLs

**Success Metrics**:
- Upload response time < 2 seconds
- Support 10 concurrent uploads
- Basic query performance < 3 seconds

### Phase 2: Analytics & Performance (Months 4-6)
**Goal**: Add log analytics and optimize performance

**New Components**:
- Log Processing Service
- Analytics Service with ClickHouse
- Redis caching layer
- Enhanced Query Service with aggregation

**Features**:
- Log search and analytics
- Advanced Kubernetes resource relationships
- Performance optimization with caching
- User authentication and authorization

**Success Metrics**:
- Query performance < 1 second
- Support log analysis on 1GB+ files
- Handle 50 concurrent uploads

### Phase 3: Scale & Polish (Months 7-12)
**Goal**: Production-ready with enterprise features

**Enhancements**:
- Horizontal scaling capabilities
- Advanced monitoring and alerting
- Disaster recovery procedures
- Enterprise security features

**Features**:
- Team-based sharing and permissions
- Advanced analytics dashboards
- API rate limiting and quotas
- Comprehensive audit logging

**Success Metrics**:
- Support 100 concurrent uploads
- Sub-second query performance at scale
- 99.9% uptime SLA

## Technology Stack Summary

### Programming Languages
- **Go**: Primary language for all services (performance, concurrency)
- **TypeScript/React**: Web UI (developer productivity, type safety)
- **SQL**: Database queries and analytics

### Infrastructure
- **Kubernetes**: Container orchestration and service discovery
- **Kong Gateway**: API gateway and traffic management
- **Prometheus/Grafana**: Monitoring and alerting
- **Jaeger**: Distributed tracing

### Data Storage
- **PostgreSQL 14+**: Transactional data and metadata
- **ClickHouse**: Analytics and log storage
- **Redis**: Caching and session storage
- **AWS S3**: Object storage for large files

### Message Queue
- **Apache Kafka**: Event streaming and async processing
- **Schema Registry**: Event schema management

### Development Tools
- **Docker**: Containerization
- **Helm**: Kubernetes deployment management
- **GitLab CI/CD**: Automated testing and deployment
- **SonarQube**: Code quality and security scanning

## Conclusion

The Kubin architecture provides a robust, scalable platform for Kubernetes snapshot sharing that meets all core requirements:

**Immediate Response**: Upload Orchestrator provides URLs within 2 seconds while coordinating background processing
**Sub-second Queries**: Multi-tier caching and optimized data storage enable fast web UI responses
**High Throughput**: Parallel upload processing and direct S3 access support 100+ concurrent operations
**Strong Consistency**: Event-driven architecture ensures data synchronization without eventual consistency issues
**Rich Analytics**: ClickHouse enables complex queries over billions of log entries

The architecture follows proven patterns from Netflix, Amazon, and other hyperscale companies, with clear service boundaries, appropriate technology choices, and a focus on operational excellence. The phased implementation approach allows for rapid MVP delivery while building toward enterprise-scale capabilities.

Key architectural decisions prioritize:
- **User Experience**: Immediate feedback and fast queries
- **Scalability**: Horizontal scaling across all components  
- **Reliability**: Strong consistency and comprehensive monitoring
- **Cost Efficiency**: Intelligent storage tiering and compute optimization
- **Security**: Defense in depth with encryption and access controls

This foundation supports Kubin's growth from startup MVP to enterprise platform while maintaining the core value proposition of instant Kubernetes snapshot sharing.
