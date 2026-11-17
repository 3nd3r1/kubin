---
id: TASK-3
title: Store and index uploaded snapshots in the server
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T14:00:00Z'
labels:
  - server
milestone: m-0
dependencies:
  - TASK-1
  - TASK-2
priority: high
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Nothing can be shared until the server accepts a snapshot. Per decision-1 upload is synchronous: validate the tarball, keep it in the blob store, index each resource in the database. Per decision-2 the default database is SQLite so a self-hoster needs nothing but a volume; Postgres comes later as TASK-26, so keep SQL portable. The snapshot ID is the only access control until there is auth, so it has to be unguessable.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `POST /api/v1/snapshots` takes a v1 tarball and returns `{id, url}`
- [x] #2 IDs are random with at least 96 bits of entropy
- [x] #3 The raw tarball is stored through a blob interface with a filesystem implementation
- [x] #4 A store interface backed by SQLite (pure Go driver, no CGO) holds a snapshots table and a resources table (snapshot, kind, namespace, name, JSON), created by migrations the server runs on start
- [x] #5 Unknown format versions, oversized uploads and invalid archives return 4xx with a message
- [x] #6 An integration test uploads a fixture tarball against a temporary SQLite file
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. blob package: Store interface (Put/Open/Delete) and FS implementation writing via temp file + rename.
2. store package: Store interface with BeginSnapshot returning a tx that streams resource rows; SQLite implementation on modernc.org/sqlite; embedded per-dialect migrations tracked in schema_migrations, run on open. Queries use $N placeholders so they work on Postgres too.
3. snapshot package: spool the upload to a temp file, read it as a v1 archive (manifest first, known version, layout paths only, regular files only, no duplicates, unpacked size cap), stream resources into the store tx, put the raw tarball in the blob store, commit. IDs are 15 random bytes in lowercase base32 (120 bits).
4. api: POST /api/v1/snapshots with MaxBytesReader, 201 {id,url}, 400/413 JSON errors. readyz pings the database.
5. config: SERVER_DATA_DIR, SERVER_PUBLIC_URL, SERVER_MAX_UPLOAD_BYTES. main wires SQLite at data/kubin.db and blobs at data/blobs.
6. snapshottest package builds a fixture tarball; integration test uploads it through the router against a temp SQLite file, plus rejection cases.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Upload spools to SERVER_DATA_DIR/tmp, is read once to validate and stream rows into one store transaction, then the raw tarball goes to the blob store before commit, so a committed snapshot always has its blob; a failed commit deletes the blob. Log entries are validated and size-counted but not indexed (TASK-14 reads them from the blob). Cluster-scoped objects get namespace "". Resource JSON is compacted before storing. Archive limits: SERVER_MAX_UPLOAD_BYTES (64 MiB) on the body, 1 GiB unpacked including 512 bytes per entry header. go.mod moved to go 1.26 for modernc.org/sqlite v1.62; Dockerfile builder bumped to golang:1.26-alpine. snapshottest.Fixture is the shared fixture for TASK-4. Verified: go test -race ./..., golangci-lint clean, and a tarball from the real cli snapshot.Writer uploaded to the built binary (201, blob and rows written).
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
POST /api/v1/snapshots accepts a v1 tarball and returns 201 {id, url}. New packages: blob (Store interface, filesystem implementation with atomic writes), store (Store interface, SQLite via modernc.org/sqlite with embedded per-dialect migrations run on start, $N placeholders so queries carry over to Postgres), snapshot (format checks from docs/snapshot-format.md, 120-bit base32 IDs, ingest). Invalid archives and unknown versions return 400, oversized bodies 413, all with a JSON error message. readyz now pings the database. Config adds SERVER_DATA_DIR, SERVER_PUBLIC_URL and SERVER_MAX_UPLOAD_BYTES. Risk: SERVER_READ_TIMEOUT (10s) caps how long an upload may take; revisit with TASK-22.
<!-- SECTION:FINAL_SUMMARY:END -->
