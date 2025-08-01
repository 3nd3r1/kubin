# Kubin - Context & Overview

## What is Kubin?

Kubin is a "pastebin for Kubernetes" - a tool that allows you to capture, store, and share snapshots of your Kubernetes clusters. Think of it like taking a screenshot of your entire cluster state and being able to share it with others or reference it later.

## Core Concept

When you run Kubin, it:
1. Connects to your Kubernetes cluster
2. Collects information about all your resources (pods, services, deployments, etc.)
3. Packages everything into a compressed snapshot
4. Uploads it to a server where you can view, share, or download it

## Project Structure

The project is divided into three main components:

### CLI Tool (`/cli`)
A command-line tool that runs on your local machine or in CI/CD pipelines. It connects to your Kubernetes cluster, collects the snapshot data, and uploads it to the server.

### Server (`/server`) 
A web service that stores and serves the snapshots. It provides a REST API for uploading, downloading, and managing snapshots. This runs as a service that can be deployed in your own cluster or hosted externally.

### Web UI (`/ui`)
A web interface for browsing, viewing, and sharing snapshots. This is where users can see the contents of snapshots, compare different versions, and share links with others.

## How It Works

1. **Capture**: CLI tool connects to your cluster and collects resource information
2. **Package**: Data is compressed and packaged with metadata
3. **Upload**: Snapshot is sent to the server via API
4. **Store**: Server stores the snapshot and makes it accessible
5. **Share**: Users can view snapshots through the web UI or download them

## Use Cases

- **Debugging**: Capture cluster state when troubleshooting issues
- **Documentation**: Save cluster configurations for reference
- **Sharing**: Share cluster states with team members or support
- **Backup**: Quick backup of cluster configurations
- **Audit**: Track changes in cluster state over time

## Key Benefits

- **Simple**: One command to capture entire cluster state
- **Portable**: Snapshots can be shared and viewed anywhere
- **Secure**: Control who can access your snapshots
- **Fast**: Quick capture and upload process
- **Searchable**: Find and compare snapshots easily 