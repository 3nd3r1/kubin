# Kubin (WIP)

![Kubin Logo](docs/assets/logo.png)

**Pastebin for Kubernetes clusters**

Capture your cluster state with one command, get a shareable link, and explore it in a Lens-like interface.

## Quick Start

```bash
# Install CLI
go install github.com/3nd3r1/kubin/cli@latest

# Capture cluster and get shareable link
kubin create

# Share the link with your team
# They can view the cluster state in their browser
```

## What it does

Instead of sharing cluster dumps as files in Jira tickets:
1. Run `kubin create`
2. Get a shareable URL
3. Anyone can explore the cluster state in a familiar Kubernetes dashboard

## Components

- **CLI** (`/cli`) - Capture clusters and get shareable links
- **Server** (`/server`) - Store and serve snapshots
- **UI** (`/ui`) - Lens-like interface for viewing snapshots

## Use Cases

- Debug cluster issues with your team
- Document cluster state for reference
- Share cluster state with support
- Track changes over time

## Development

```bash
# Build CLI
cd cli && go build

# Build Server
cd server && go build

# Run Server
cd server && go run cmd/server/main.go
```

## License

MIT 
