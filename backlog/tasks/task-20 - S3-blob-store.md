---
id: TASK-20
title: S3 blob store
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-24T10:31:00Z'
labels:
  - server
  - tier:3
dependencies:
  - TASK-3
priority: high
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With more than one replica, tarballs and logs blobs can't stay on local disk. The chart (TASK-7), compose (TASK-27) and the public instance (TASK-23) keep them in S3 or MinIO. The local directory stays the default (decision-2).

## Where

- `server/internal/blob/blob.go`: the `Store` interface. `fs.go` is the existing implementation and `fs_test.go` its tests.
- New `server/internal/blob/s3.go` on `github.com/minio/minio-go/v7`, which talks to AWS S3 and MinIO and is pure Go.
- `server/internal/config/config.go`, these names, which TASK-7 and TASK-27 use:
  - `SERVER_S3_BUCKET`: set means S3, empty means the local directory
  - `SERVER_S3_ENDPOINT`: empty means AWS (`s3.amazonaws.com`); for MinIO like `http://minio:9000`
  - `SERVER_S3_REGION`: default `us-east-1`
  - `SERVER_S3_ACCESS_KEY_ID`, `SERVER_S3_SECRET_ACCESS_KEY`: empty means the standard AWS chain (`AWS_*` env, web identity, instance role)
  - `SERVER_S3_PATH_STYLE`: `true` for MinIO
  - `SERVER_S3_PREFIX`: optional key prefix
- `server/cmd/main.go`: pick the store. `Ready` is `db.Ping` today; make `/readyz` check the blob store too, for example a `Ping(ctx) error` on `blob.Store` (the FS store stats its directory, S3 checks the bucket).
- `docs/self-hosting.md`: config table.

## Spec

- The server checks the bucket at startup and exits with a clear message when it is missing or the credentials are refused. It never creates buckets.
- Keys go under `SERVER_S3_PREFIX` when set. Keep the key checks `fs.go` does.

## Out of scope

Postgres (TASK-26), the chart and compose, moving existing blobs between stores, presigned uploads.

## Traps

- `OpenRange` has to behave like `fs.go`, which uses `io.NewSectionReader`: a length of 0 gives an empty reader (empty logs exist), reading past the end stops at the end, and an offset at or past the end gives an empty reader. S3's Range header is inclusive (`bytes=start-end`), and S3 answers 416 for a start past the end. Run one behavior test table against both stores.
- `Open` and `OpenRange` on a missing key return `blob.ErrNotFound`. minio-go reports a missing object lazily on the first Read, so stat first or check for `NoSuchKey`.
- `Put`: "a reader that fails part way leaves nothing behind". Abort failed multipart uploads. `snapshot.Service` passes an `*os.File`, so pass its size; for unknown sizes set a part size (like 16 MiB) so memory stays bounded.
- MinIO needs path-style addressing; AWS uses virtual-hosted style.
- MinIO stopped publishing community images in late 2025. Pin a `minio/minio:RELEASE.*` tag that `docker pull` still fetches, and write it in the implementation notes so TASK-27 uses the same one.
- Docker-backed tests may skip only when Docker is unreachable and `CI` is unset; with `CI=true` they fail. If TASK-26 already added a testcontainers helper, reuse it.
- Never log the secret key.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 With `SERVER_S3_BUCKET` set the server keeps blobs in that bucket; unset, it uses the local directory as today
- [ ] #2 One behavior test table runs against the FS store and against a MinIO container the test starts: round trip, missing key gives ErrNotFound, OpenRange with empty, in-range and past-the-end ranges, Delete of a missing key, and a failed Put leaving nothing
- [ ] #3 Config covers AWS S3 (virtual-hosted style, standard credential chain when no keys are set) and MinIO (endpoint, path style), with a test of how config maps to client options
- [ ] #4 /readyz answers 503 when the blob store is unreachable, for both stores
- [ ] #5 The server exits at startup with a clear message when the bucket is missing or the credentials are refused
- [ ] #6 docs/self-hosting.md documents every SERVER_S3_ variable
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd server && CI=true make test && make lint passes, with Docker running so the MinIO tests run rather than skip
- [ ] #2 Manual: a MinIO container with a bucket, cd server && make run with SERVER_S3_* set, an upload with cd cli && go run . create --server http://localhost:8080 -n kube-system, and a pod log from the link loads
- [ ] #3 The pinned MinIO image tag is written in the implementation notes
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-14: blob.Store has OpenRange(key, offset, length) for serving logs; map it to a ranged GetObject.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
created: 2026-11-24T10:31:00Z

---

Left m-3 with decision-12: one node with a local directory is enough for now (docs/capacity.md). Comes back when one node runs out.

---
<!-- COMMENTS:END -->
