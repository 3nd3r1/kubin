# Kubin CLI

Capture Kubernetes clusters and get shareable links.

## Install

```bash
go install github.com/3nd3r1/kubin/cli@latest
```

## Usage

```bash
# Capture entire cluster
kubin create

# Capture specific namespace
kubin create --namespace prod
```

## What it does

1. Connects to your Kubernetes cluster
2. Collects resources (pods, services, etc.) and logs
3. Packages everything into a snapshot
4. Uploads to server and returns a shareable URL

## Commands

- `kubin create` - Capture cluster and get shareable link
- `kubin list` - List your snapshots
- `kubin get <id>` - Get snapshot details

## Configuration

Set server URL:
```bash
export KUBIN_SERVER_URL=https://kubin.example.com
```

Or use config file `~/.kubin/config.yaml`:
```yaml
server:
  url: https://kubin.example.com
```

## Build

```bash
go build -o kubin
``` 