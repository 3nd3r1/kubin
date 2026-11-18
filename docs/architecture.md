# Architecture

Kubin is a CLI and one server. The CLI packs a cluster into a tarball and uploads it. The server stores it, indexes it and serves both the API and the web UI from the same binary. A self-hosted install is one container and one volume.

```
  laptop / CI                              kubin (one Go binary, one image)
 ┌───────────────────┐                   ┌──────────────────────────────────┐
 │ kubin create      │  POST tarball     │  /api/v1/...  REST API           │
 │  collect, redact, │ ────────────────▶ │  /s/{id}      web UI (embedded)  │
 │  pack, upload     │ ◀── {id, url} ─── │                                  │
 └───────────────────┘                   │  blob ──▶ ./data or S3           │
                                         │  store ─▶ SQLite or Postgres     │
 browser ── https://kubin.you/s/7fK2q ─▶ └──────────────────────────────────┘
```

The earlier design had seven services plus Kafka, ClickHouse and Redis. Why it changed: `backlog decision view decision-1` and `decision-2`.

## Flow

1. `kubin create` uses the current kubeconfig context, like kubectl.
2. Collectors read resources and logs. Secret values and `managedFields` are stripped. The result is a versioned tarball.
3. The CLI uploads it. The server validates it, stores the raw tarball in the blob store and writes one row per resource to the database.
4. The server answers with a random, unguessable ID. The CLI prints the link.
5. The browser loads the UI from the same server. The UI calls the API for resource lists, single objects and logs. Logs are streamed from the blob store.
6. Snapshots expire and a job inside the server deletes them.

## Snapshot

A gzipped tar, and the contract between the CLI and the server. Its layout is versioned in `manifest.json`.

```
manifest.json                                 format version, cluster info, CLI version, counts
resources/<group>/<version>/<Kind>/<namespace>/<name>.json
resources/<group>/<version>/<Kind>/_cluster/<name>.json     cluster-scoped kinds
logs/<namespace>/<pod>/<container>.log        and <container>.previous.log after a restart
```

The core group is written as `core`. The full spec is [snapshot-format.md](snapshot-format.md).

## CLI (`cli/`)

Go, cobra, client-go. `kubin create` uploads to `--server` or `KUBIN_SERVER` and prints the URL. `--output <file>` keeps the tarball instead, for air-gapped use. Collectors implement `collector.Collector` and return objects as `collector.Resource`. `collector.LogsCollector` then fetches the logs of the collected pods with a fixed worker pool, keeping the last `--log-tail` lines of each. `snapshot.Writer` puts them into the tarball layout, and `snapshot.Manager` runs the collectors and writes the manifest.

## Server (`server/`)

One Go binary with chi. The parts the old design made into services are packages:

| Package | Job |
|---|---|
| `api` | HTTP handlers, validation |
| `snapshot` | upload: validate the tarball, store it, index it |
| `store` | snapshots and resource index. SQLite by default, Postgres by config |
| `blob` | raw tarballs. Local directory by default, S3 by config |
| `web` | the built UI, embedded with go:embed |

### API

```
POST /api/v1/snapshots                                 upload tarball, returns {id, url}
GET  /api/v1/snapshots/{id}                            manifest and counts
GET  /api/v1/snapshots/{id}/resources?kind=&namespace= list
GET  /api/v1/snapshots/{id}/resources/{kind}/{ns}/{name}
GET  /api/v1/snapshots/{id}/logs/{ns}/{pod}/{container}
GET  /healthz, /readyz
GET  /*                                                the UI
```

## UI (`ui/`)

A Vite React app with Tailwind and shadcn/ui, built into the server (TASK-24). `/s/{id}` is a snapshot. The layout follows Lens: kinds in a sidebar, a table per kind, a detail panel with YAML, events and logs. Read-only.

## Deploying

The same image covers three setups:

| Setup | For | Runs |
|---|---|---|
| `docker run -v kubin-data:/data` | a team, a homelab, one VM | kubin with SQLite and files in the volume |
| docker compose (TASK-27) | a small company on one VM | kubin, Postgres, MinIO |
| Helm chart (TASK-7) | Kubernetes shops, the public instance | several replicas, Postgres, S3, ingress |

SQLite and a local directory limit an install to one replica. More replicas need Postgres and S3. Install steps and configuration are in [self-hosting.md](self-hosting.md).

## Load

Upload is synchronous. A snapshot should be a few MB, and unpacking it and inserting the rows is cheap next to receiving the bytes. Nobody has measured this yet. The server keeps no state outside the database and blob store, so it scales by adding replicas.

A queue would let the server answer before indexing finishes. It would not reduce the total work. If indexing ever gets slow enough to hold up uploads, the next step is a jobs table in the database and a worker pool inside the server, and after that letting the CLI upload straight to S3 with presigned URLs. Kafka is not planned.

## Security

A cluster dump can leak credentials.

- Secret objects are collected without `data` and `stringData`.
- Env var values on pods can hold secrets. The default is open (TASK-12).
- Snapshot IDs are unguessable, since the link is the only access control until there is auth. Snapshots expire.
