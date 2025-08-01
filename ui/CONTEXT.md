# Kubin UI - Context

## What is the UI?

The Kubin UI is a web interface that provides a Lens-like experience for viewing Kubernetes cluster snapshots. It's like having a Kubernetes dashboard for historical cluster states, allowing users to explore what their cluster looked like at a specific point in time.

## Core Concept

Think of it as a "time machine" for your Kubernetes cluster. Instead of sharing cluster dumps as files in Jira tickets, you run one CLI command, get a shareable link, and anyone can explore the cluster state through a familiar Kubernetes dashboard interface.

## Core Responsibilities

### 1. Snapshot Exploration

- Displays cluster state as it existed at snapshot time
- Shows hierarchical view of namespaces and resources
- Provides detailed resource information and relationships
- Displays pod logs that were captured in the snapshot

### 2. Resource Navigation

- Hierarchical navigation like Lens (namespaces → resources → details)
- Focus on pods as the primary resource (MVP)
- Clean, structured display of resource data
- Easy navigation between related resources

### 3. Historical Analysis

- View cluster state at specific points in time
- Compare different snapshots to see changes
- Analyze what was happening when issues occurred
- Debug problems using historical context

### 4. Sharing & Collaboration

- Share snapshot links with team members
- Embed snapshots in documentation or tickets
- Provide read-only access to cluster state
- Enable remote debugging and support

## Key Features

- **Lens-like Interface**: Familiar Kubernetes dashboard experience
- **Historical Viewing**: Explore cluster state at snapshot time
- **Pod-focused**: Detailed pod information and logs (MVP)
- **Hierarchical Navigation**: Namespaces → Resources → Details
- **Shareable Links**: One-click sharing of cluster state
- **Clean Data Display**: Well-formatted resource information

## User Experience

Users can:

- Open a snapshot link and see the cluster as it was
- Navigate through namespaces and resources like in Lens
- Click on pods to see their details and captured logs
- Share the link with colleagues for debugging
- Use it as a replacement for cluster dump files

## Interface Design

The UI follows:

- **Lens-inspired**: Familiar Kubernetes dashboard layout
- **Snapshot-aware**: Clear indication of historical data
- **Clean Formatting**: Well-structured resource displays
- **Fast Loading**: Optimized for snapshot data access
- **Mobile-friendly**: Responsive design for different devices

## Deployment Model

- **Co-located**: UI is deployed alongside the Kubin server
- **Single Instance**: One UI per server instance
- **Organization-focused**: Each organization has their own server+UI
- **Public Service**: Also available as a hosted service

## Integration Points

- **Kubin Server**: Fetches snapshot data via REST API
- **Snapshot Data**: Displays captured cluster state and logs
- **Share Links**: Generates and handles shareable URLs
- **Authentication**: Integrates with server's auth system
