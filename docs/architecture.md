# Architecture

Kubin is three parts: a CLI that packs a cluster into one file, a server that stores it and answers queries about it, and a web UI that reads from the server. One server binary, one Postgres, one blob store.

```
kubin create ──upload .tar.gz──▶ server ──▶ blob store (disk, later S3)
                                   │
                                   └──────▶ Postgres (snapshots, resource index)
browser ──▶ ui (Next.js) ──REST──▶ server
```

The earlier design had seven services plus Kafka, ClickHouse and Redis. It was replaced before any of it was built. Why is in `backlog decision view decision-1`.

## Snapshot

A snapshot is a gzipped tar. It is the contract between the CLI and the server, so its layout is versioned in `manifest.json`.

```
manifest.json                                 format version, cluster info, CLI version, counts
resources/<group>/<version>/<kind>/<namespace>/<name>.json
resources/<group>/<version>/<kind>/_cluster/<name>.json     cluster-scoped kinds
logs/<namespace>/<pod>/<container>.log
```

The core group is written as `core`. The CLI strips `managedFields` and never writes Secret data. The full spec will live in `docs/snapshot-format.md` once TASK-2 lands.

## CLI (`cli/`)

Go, cobra, client-go. `kubin create` reads the current kubeconfig context, runs each collector, writes the tarball and uploads it. It prints the snapshot URL. `--output <file>` keeps the tarball instead of uploading, for air-gapped use.

Collectors implement `collector.Collector`. Each returns resources, the persister writes them into the tarball layout.

## Server (`services/api-gateway/` today, moving to `server/` in TASK-1)

One Go binary with chi. Everything the old design split into services is a package here:

| Package | Job |
|---|---|
| `api` | HTTP handlers, request validation |
| `snapshot` | upload: validate the tarball, store it, index it |
| `store` | Postgres: snapshots table and resource index |
| `blob` | raw tarballs and logs. Filesystem first, S3 later |

Upload is synchronous. A snapshot is a few MB, so validating and indexing it inside the request is fine. If it stops being fine, the indexing moves to a goroutine and the snapshot gets a status. It doesn't need a queue.

Postgres holds one row per snapshot and one row per resource (kind, namespace, name, the JSON). That is enough for listing, filtering and showing any resource. Logs stay as files in the blob store and are streamed to the UI. Log search starts as a scan of one log file. ClickHouse only comes back if that gets slow on real snapshots.

### API

```
POST /api/v1/snapshots                                 upload tarball, returns {id, url}
GET  /api/v1/snapshots/{id}                            manifest and counts
GET  /api/v1/snapshots/{id}/resources?kind=&namespace= list
GET  /api/v1/snapshots/{id}/resources/{kind}/{ns}/{name}
GET  /api/v1/snapshots/{id}/logs/{ns}/{pod}/{container}
GET  /healthz, /readyz
```

Snapshot IDs are random and unguessable, since the link is the only access control until there is auth.

## UI (`ui/`)

Next.js, Tailwind, shadcn/ui. `/s/{id}` is a snapshot. Layout follows Lens: kinds in a sidebar, a table per kind, a detail panel with YAML, events and logs. Read-only.

## Deploy (`charts/`, `skaffold.yaml`)

An umbrella Helm chart with a subchart per deployable (server, ui) and Bitnami Postgres as a dependency. `skaffold dev` against a local kind cluster is the dev loop. The ArgoCD app-of-apps template stays off until there is a hosted instance.

## Security

A cluster dump can leak credentials. The rules:

- Secret objects are collected without `data` and `stringData`.
- Env var values on pods can hold secrets. How to handle them is open (TASK-12).
- Snapshot IDs are unguessable and snapshots expire.
