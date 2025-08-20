# Kubin - Context & Overview

## What is Kubin?

Kubin is a "pastebin for Kubernetes" - a tool that allows you to capture, store, and share snapshots of your Kubernetes clusters. Think of it like taking a screenshot of your entire cluster state and being able to share it with others or reference it later.

## Core Concept

When you run Kubin, it:

1. Connects to your Kubernetes cluster
2. Collects information about all your resources (pods, services, deployments, etc.)
3. Packages everything into a compressed snapshot
4. Uploads it to a microservices platform for storage and processing
5. Provides a shareable URL for viewing the snapshot in a web interface

## Project Structure

The project uses a modern microservices architecture with three main user-facing components:

### CLI Tool (`/cli`)

A command-line tool that runs on your local machine or in CI/CD pipelines. It connects to your Kubernetes cluster, collects the snapshot data, and uploads it to the platform. Distributed as a binary like `kubectl` or `docker`.

### Microservices Platform (`/services`)

A distributed backend system consisting of specialized services:

- **API Gateway**: Routes requests and handles authentication
- **Log Service**: Processes and stores log data for fast analytics
- **Storage Service**: Handles file operations and compression

### Web UI (`/ui`)

A React-based web interface for browsing, viewing, and sharing snapshots. Provides a Lens-like experience for exploring historical cluster states. Deployed as a containerized application.

## How It Works

1. **Capture**: CLI tool connects to your cluster and collects resource information
2. **Package**: Data is compressed and packaged with metadata
3. **Upload**: Snapshot is sent to the platform via API Gateway
4. **Process**: Services handle the data asynchronously - metadata goes to PostgreSQL, logs to ClickHouse, files to S3
5. **Share**: Users can view snapshots through the web UI using shareable links

## Use Cases

- **Debugging**: Capture cluster state when troubleshooting issues
- **Documentation**: Save cluster configurations for reference
- **Sharing**: Share cluster states with team members or support
- **Backup**: Quick backup of cluster configurations
- **Audit**: Track changes in cluster state over time
- **Analytics**: Query and analyze historical log data

## Key Benefits

- **Simple**: One command to capture entire cluster state
- **Fast**: Optimized storage and retrieval with sub-second queries
- **Scalable**: Microservices architecture handles high load
- **Portable**: Snapshots can be shared and viewed anywhere
- **Secure**: Control who can access your snapshots
- **Searchable**: Find and compare snapshots easily with powerful analytics

## Architecture Highlights

- **Hybrid Communication**: Synchronous for user operations, asynchronous for heavy processing
- **Optimized Storage**: PostgreSQL for metadata, ClickHouse for logs, S3 for files
- **High Performance**: Sub-second search on massive datasets
- **Cloud Native**: Kubernetes deployment with auto-scaling and monitoring

