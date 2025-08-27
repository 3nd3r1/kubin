# Kubin Server - Context

## What is the Server?

The Kubin Server is a web service that stores, manages, and serves Kubernetes cluster snapshots. It provides the backend infrastructure for the entire Kubin platform, handling snapshot storage, user authentication, and API access.

## Core Responsibilities

### 1. Snapshot Storage
- Stores uploaded snapshots securely
- Manages snapshot metadata and indexing
- Handles snapshot compression and decompression
- Provides snapshot retrieval and download

### 2. API Management
- Exposes REST API for snapshot operations
- Handles upload, download, and management requests
- Provides search and filtering capabilities
- Manages API rate limiting and security

### 3. User Management
- Handles user authentication and authorization
- Manages user accounts and permissions
- Controls access to public and private snapshots
- Provides API key management for CLI access

### 4. Data Management
- Organizes snapshots by user and organization
- Implements snapshot versioning and history
- Manages snapshot retention and cleanup
- Provides backup and recovery capabilities

## Key Features

- **RESTful API**: Standard HTTP/JSON API for all operations
- **Authentication**: JWT-based user authentication
- **Authorization**: Role-based access control
- **Search**: Full-text search across snapshots
- **Sharing**: Public and private snapshot sharing
- **Versioning**: Track changes in snapshots over time

## Service Architecture

The server is designed as:
- **Stateless**: Can be scaled horizontally
- **Containerized**: Runs in Kubernetes or Docker
- **Database-backed**: Persistent storage for metadata
- **File storage**: Separate storage for snapshot files
- **Load balanced**: Supports multiple instances

## API Endpoints

Core endpoints include:
- `POST /snapshots` - Upload new snapshot
- `GET /snapshots` - List snapshots
- `GET /snapshots/{id}` - Get snapshot details
- `GET /snapshots/{id}/download` - Download snapshot
- `DELETE /snapshots/{id}` - Delete snapshot

## Integration Points

- **CLI Tool**: Receives uploaded snapshots
- **Web UI**: Provides data for the frontend
- **Database**: Stores metadata and user information
- **File Storage**: Stores actual snapshot files
- **Authentication**: Integrates with identity providers 