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
2. Collectors read resources and logs. Secret values, env var values and `managedFields` are stripped. The result is a versioned tarball.
3. The CLI uploads it. The server validates it, stores the raw tarball in the blob store and writes one row per resource to the database.
4. The server answers with a random, unguessable ID. The CLI prints the link.
5. The browser loads the UI from the same server. The UI calls the API for resource lists, single objects and logs. Logs are streamed from the blob store: at upload the server copies every log out of the tarball into one blob per snapshot and records where each starts, so serving a log is one ranged read.
6. Each upload gets an expiry (`SERVER_SNAPSHOT_TTL` by default, or `?expires_in=` up to `SERVER_SNAPSHOT_MAX_TTL`) and a random `delete_token` (only its SHA-256 hash is stored). Once `expires_at` passes, every read route returns `410` until a periodic job inside the server removes both blobs and the database row, after which reads return `404`. Owners can also delete early with `DELETE /api/v1/snapshots/{id}` and their bearer token. Snapshots stored before expiry existed have a `NULL` `expires_at` and never expire.

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

Go, cobra, client-go. `kubin create` uploads to `--server` or `KUBIN_SERVER` and prints the URL. `--output <file>` keeps the tarball instead, for air-gapped use. Collectors implement `collector.Collector` and return objects as `collector.Resource`. `collector.ResourcesCollector` lists every kind in `collector.Kinds`, the one allowlist, through discovery and the dynamic client, at the version the server prefers. A kind the user can't list is skipped with a warning. `collector.LogsCollector` then fetches the logs of the collected pods with a fixed worker pool, keeping the last `--log-tail` lines of each. `snapshot.Writer` puts them into the tarball layout, and `snapshot.Manager` runs the collectors and writes the manifest.

## Server (`server/`)

One Go binary with chi. The parts the old design made into services are packages:

| Package | Job |
|---|---|
| `api` | HTTP handlers, validation |
| `snapshot` | upload: validate the tarball, store it, index it |
| `store` | snapshots, resource index and where each log sits in the logs blob. SQLite by default, Postgres by config |
| `blob` | raw tarballs and logs blobs, read whole or by range. Local directory by default, S3 by config |
| `web` | the built UI, embedded with go:embed |

### API

```
POST   /api/v1/snapshots?expires_in=                   upload tarball, returns {id, url, expires_at, delete_token}
DELETE /api/v1/snapshots/{id}                          delete with Authorization: Bearer <delete_token> (204)
GET    /api/v1/snapshots/{id}                          manifest, counts and expires_at
GET    /api/v1/snapshots/{id}/resources?kind=&namespace= list
GET    /api/v1/snapshots/{id}/resources/{kind}/{ns}/{name}
GET    /api/v1/snapshots/{id}/logs/{ns}/{pod}/{container}?previous=true&q=   text/plain, q filters lines
GET    /healthz, /readyz
GET    /*                                              the UI
```

Read routes return `410` once a snapshot's `expires_at` has passed, and `404` after the cleanup job or a `DELETE` removes it.

## UI (`ui/`)

A Vite React app with Tailwind and shadcn/ui, built into the server (TASK-24). `/s/{id}` is a snapshot. The layout follows Lens: kinds in a sidebar, a table per kind, a detail panel with YAML, events and logs. Read-only.

URLs: `/s/{id}` is the problems-first overview. `/s/{id}/{resource}` lists one kind, `{resource}` being kubectl's lowercase plural (`pods`, `deployments`, `persistentvolumeclaims`). `/s/{id}/{resource}/{namespace}/{name}` is a resource's detail, with `_cluster` as the namespace of cluster-scoped kinds, as in the API. `?ns={namespace}` filters every namespaced kind and the sidebar links keep it; without it all namespaces show.

## Deploying

The same image covers three setups:

| Setup | For | Runs |
|---|---|---|
| `docker run -v kubin-data:/data` | a team, a homelab, one VM | kubin with SQLite and files in the volume |
| docker compose | the same, written down as a file | kubin with SQLite and files in the volume |
| Helm chart | Kubernetes shops | one replica, SQLite and files on a PVC, ingress |

The public instance runs as one container on Coolify (decision-12). SQLite and a local directory limit an install to one replica, which is enough ([capacity.md](capacity.md)). Postgres and S3 come back when one node runs out, and then more replicas need both. Install steps and configuration are in [self-hosting.md](self-hosting.md).

## Load

Upload is synchronous. A snapshot should be a few MB, and unpacking it and inserting the rows is cheap next to receiving the bytes. One container takes about 3 uploads a second of a 600-resource snapshot and serves about 2,000 reads a second ([capacity.md](capacity.md)). The server keeps no state outside the database and blob store, so it scales by adding replicas.

A queue would let the server answer before indexing finishes. It would not reduce the total work. If indexing ever gets slow enough to hold up uploads, the next step is a jobs table in the database and a worker pool inside the server, and after that letting the CLI upload straight to S3 with presigned URLs. Kafka is not planned.

## Security

A cluster dump can leak credentials.

- Every object goes through `redact.Redactor` in the CLI, whichever collector produced it. Secret values are emptied with their keys kept, env var values are dropped (decision-5, `--keep-env-values` opts out), and the `last-applied-configuration` annotation gets the same treatment. The manifest lists the rules that ran. Details in [snapshot-format.md](snapshot-format.md#redaction).
- Logs, container args and ConfigMaps are not redacted.
- Snapshot IDs are unguessable, since the link is the only access control until there is auth. Snapshots expire.
