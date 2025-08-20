# Kubin CLI - Context

## What is the CLI?

The Kubin CLI is a command-line tool that captures Kubernetes cluster snapshots and generates shareable links. It's the entry point for users to create snapshots of their clusters and get a URL that can be shared with team members for debugging and analysis.

## Distribution & Installation

The CLI is distributed as a standalone binary (like kubectl, docker, or git):

- Binary releases for Linux, macOS, Windows
- Package managers: brew install kubin, apt install kubin
- Go install: go install github.com/kubin/cli
- Container image for CI/CD pipelines only

## Core Responsibilities

### 1. Cluster Connection

- Connects to Kubernetes clusters using standard kubeconfig
- Supports multiple cluster contexts
- Handles authentication and authorization

### 2. Resource Collection

- Collects Kubernetes resources (pods, services, deployments, etc.)
- Captures pod logs and resource states
- Gathers cluster metadata (version, context, etc.)
- Supports filtering by namespace, labels, or resource types

### 3. Snapshot Creation

- Packages collected data into compressed snapshots
- Creates metadata about the snapshot (timestamp, cluster info, etc.)
- Validates snapshot integrity
- Ensures all data needed for UI viewing is captured

### 4. Upload & Link Generation

- Uploads snapshots to the Kubin server
- Handles upload progress and error recovery
- Manages authentication with the server
- Returns shareable URL for the snapshot

## Key Features

- **Simple Commands**: One command to capture entire cluster
- **Flexible Filtering**: Choose what resources to include
- **Progress Feedback**: Shows collection and upload progress
- **Error Handling**: Graceful handling of connection issues
- **Configuration**: Supports config files and environment variables

## User Experience

Users will typically run commands like:

```bash
kubin create                    # Capture entire cluster and get shareable link
kubin create --namespace prod   # Capture specific namespace
kubin create --upload           # Capture and upload immediately
```

The workflow is: run command → get URL → share with team → they can view cluster state in browser

## Architecture Overview

The CLI is built with:

- **Collectors**: Modules that gather different types of resources
- **Kubernetes Client**: Interface to interact with clusters
- **Snapshot Manager**: Orchestrates the collection and packaging
- **Upload Client**: Handles communication with the server

## Integration Points

- **Kubernetes API**: Reads cluster resources
- **Kubin Server**: Uploads snapshots via REST API
- **Local Storage**: Temporarily stores snapshots before upload
- **Configuration**: Reads user preferences and settings

