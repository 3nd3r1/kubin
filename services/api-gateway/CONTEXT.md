# API Gateway - Context

## What is the API Gateway?

The API Gateway is the single entry point for all external requests to the Kubin platform. It handles routing, authentication, rate limiting, and serves as the orchestration layer between the CLI, Web UI, and the underlying microservices.

## Core Responsibilities

### 1. Request Routing

- Routes requests to appropriate backend services
- Load balances traffic across service instances
- Handles service discovery and health checking
- Provides unified API versioning (`/api/v1`, `/internal/api/v1`)

### 2. Security & Authentication

- TLS termination and security headers
- JWT-based authentication validation
- API rate limiting (100 concurrent uploads, 1000 queries/min)
- CORS handling for web UI integration

### 3. Traffic Management

- Request/response middleware stack
- Error handling and recovery
- Request logging and monitoring
- Graceful shutdown coordination

## Architecture Role

The API Gateway serves as the orchestration layer in Kubin's microservices architecture:

**Upstream Services:**

- **Upload Orchestrator**: Handles snapshot upload coordination
- **Metadata Service**: Manages PostgreSQL-backed snapshot metadata
- **Storage Service**: Coordinates S3 file operations
- **Query Service**: Orchestrates data retrieval from multiple services
- **Analytics Service**: Provides ClickHouse-powered log analytics

**Client Interfaces:**

- **CLI Tool**: Direct API access for snapshot uploads
- **Web UI**: Internal API endpoints for frontend functionality

## API Structure

### External API (`/api/v1`)

- `POST /api/v1/snapshots` - Initiate snapshot upload (→ Upload Orchestrator)
- `GET /api/v1/snapshots` - List snapshots (→ Metadata Service)

### Internal API (`/internal/api/v1`)

- `GET /internal/api/v1/snapshots/{id}/resources` - Get snapshot resources
- `GET /internal/api/v1/snapshots/{id}/pods` - Get pod information
- `GET /internal/api/v1/snapshots/{id}/logs` - Get pod logs (→ Analytics Service)
- `GET /internal/api/v1/snapshots/{id}/namespaces` - Get namespaces

## Communication Patterns

### Synchronous Routing

- CLI upload requests → Upload Orchestrator (immediate response)
- Web UI queries → Query Service (real-time data)
- File operations → Storage Service (direct S3 coordination)

### Middleware Stack

1. **Recovery**: Panic recovery and error handling
2. **Logging**: Request/response logging
3. **Security**: Headers, CORS, authentication
4. **JSON**: Content-type management

## Integration with Microservices Architecture

The API Gateway is part of Kubin's distributed system that separates concerns:

- **Upload Path**: API Gateway → Upload Orchestrator → Storage Service → Kafka → Log Processor
- **Query Path**: API Gateway → Query Service → (Metadata Service + Analytics Service + Storage Service)
- **Caching**: Redis integration for performance optimization

## Key Features

- **High Throughput**: Optimized for concurrent snapshot uploads
- **Fault Tolerance**: Circuit breakers and graceful degradation
- **Observability**: Comprehensive logging and health checks
- **Scalability**: Stateless design for horizontal scaling

## Health & Monitoring

- `/healthz` - Basic health check
- `/readyz` - Readiness check (validates downstream services)
- Structured logging with request tracing
- Error reporting and recovery patterns

## Configuration

Environment-based configuration supporting:

- Server timeouts and connection limits
- Authentication providers and JWT settings
- Rate limiting policies
- Allowed origins for CORS
- Service discovery endpoints

