# Architecture

## System Design

Kubin uses a microservices architecture optimized for high-throughput uploads and sub-second queries. The system separates upload and query paths to prevent interference and enable independent scaling.

### Diagram

![Architecture](./docs/assets/kubin_architecture.png)

## Core Services

### API Gateway

- Single entry point with TLS, auth, rate limiting
- Routes traffic to appropriate services
- 100 concurrent uploads, 1000 queries/min limits

### Upload Orchestrator

- Coordinates multi-step upload workflow
- Generates snapshot IDs and immediate shareable URLs
- Manages S3 pre-signed URLs and async processing
- States: `INITIATED` → `UPLOADING` → `PROCESSING` → `COMPLETED`

### Metadata Service

- Fast PostgreSQL storage for K8s resource metadata
- Optimized for listing, filtering, and search operations
- Handles user management and sharing permissions

### Storage Service

- Manages S3 file operations and organization
- Generates pre-signed URLs for direct uploads/downloads
- Implements lifecycle policies and multipart uploads

### Log Processor

- Background parsing of raw logs into structured data
- Batch inserts to ClickHouse for analytics
- Handles various log formats with parallel processing

### Query Service

- Orchestrates data from multiple services
- Parallel requests with intelligent caching (Redis)
- Provides unified responses for web UI

### Analytics Service

- Advanced log search and filtering on ClickHouse
- Time-series analysis and aggregations
- Materialized views for common queries

## Data Storage

**PostgreSQL**: User data, snapshot metadata, K8s resources

- ACID transactions for immediate consistency
- Optimized indexes for fast queries
- Handles user auth and sharing permissions

**ClickHouse**: Log data and analytics

- Columnar storage for fast aggregations
- Partitioned by snapshot_id and time
- Materialized views for common patterns

**S3**: Raw files and YAML resources

- Direct upload/download via pre-signed URLs
- Organized hierarchy: `snapshots/{id}/{type}/...`
- Lifecycle policies for cost optimization

**Redis**: Multi-layer caching

- Request cache (15min), component cache (5min)
- Search cache (1hr), CDN cache (global)

## Communication Patterns

**Synchronous (HTTP/gRPC)**:

- CLI ↔ Upload Orchestrator (immediate response)
- Web UI ↔ Query Service (real-time interaction)
- Inter-service calls for request fulfillment

**Asynchronous (Kafka)**:

- S3 upload completion events
- Background processing coordination
- Cross-service event propagation

## Upload Flow

### Phase 1: Immediate Response (< 2s)

1. CLI sends cluster metadata to Upload Orchestrator
2. Generate snapshot ID, create PostgreSQL record
3. Request S3 pre-signed URLs from Storage Service
4. Return shareable URL and upload URLs to CLI

### Phase 2: Parallel Upload (5-120s)

1. CLI uploads files directly to S3 using pre-signed URLs
2. Multipart uploads for large files, progress reporting
3. S3 events trigger Kafka notifications

### Phase 3: Background Processing (30-300s)

1. Log Processor parses and indexes logs
2. Metadata status updates: `uploading` → `processing` → `completed`
3. Query Service cache warming

## Query Flow

**List View**: Redis cache → PostgreSQL metadata query
**Detail View**: Parallel calls to Metadata + Storage + Analytics services  
**File Access**: Direct S3 downloads with pre-signed URLs
**Log Search**: ClickHouse queries with partition pruning

## Performance Optimizations

- **Parallel Processing**: Worker pools, goroutines, async patterns
- **Intelligent Caching**: Multi-layer Redis + CDN caching
- **Direct S3 Access**: Bypass servers for file operations
- **Database Optimization**: Partitioning, indexing, materialized views
- **Connection Pooling**: HTTP/2, gRPC with connection reuse

## Scaling Strategy

**Horizontal Scaling**: All services are stateless and horizontally scalable
**Database Scaling**: Read replicas for PostgreSQL, ClickHouse cluster sharding
**Storage Scaling**: S3 auto-scales, Redis cluster for cache
**Load Balancing**: API Gateway with health checks and circuit breakers
