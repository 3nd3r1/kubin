---
id: TASK-21
title: Expire snapshots and delete them by token in the server
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-24T09:37:00Z'
labels:
  - server
  - tier:4
milestone: m-3
dependencies:
  - TASK-3
priority: high
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
doc-1 m-3: snapshots expire. A dump holds cluster details, the link is the only access control (docs/architecture.md, Security), and a public instance can't keep dumps forever. This task is the server side and defines the API that TASK-45 (CLI) and TASK-46 (UI) build on.

## Where

- `server/internal/store/`: a new migration in `migrations/sqlite/` (and `migrations/postgres/` if TASK-26 has landed), `store.go`, `sql.go`, `read.go`.
- `server/internal/snapshot/snapshot.go` (`Service.Upload`): expiry and delete token.
- `server/internal/api/`: `snapshots.go` (upload response, new DELETE), `read.go` (the `snapshot` helper every read route uses, which is where 410 goes), `router.go`.
- A cleanup loop started from `server/cmd/main.go` that stops when its context ends, like `server.Serve`.
- `server/internal/config/config.go`, `docs/architecture.md` (API section, Flow step 6), `docs/self-hosting.md`.

## Proposed contract (the plan confirms or changes it)

- `POST /api/v1/snapshots?expires_in=72h` (Go duration). Without it: `SERVER_SNAPSHOT_TTL`, default `168h`. Above `SERVER_SNAPSHOT_MAX_TTL` (default `720h`) or not positive: 400 naming the maximum. The response adds `expires_at` (RFC 3339) and `delete_token`.
- `GET /api/v1/snapshots/{id}` adds `expires_at`.
- Once `expires_at` has passed, every `/api/v1/snapshots/{id}...` read returns 410 `{"error":"snapshot expired"}`, even before the job has removed it. After removal it is a 404.
- `DELETE /api/v1/snapshots/{id}` with `Authorization: Bearer <delete_token>`: 204 and the snapshot is gone. A wrong or missing token gives 403 and removes nothing.
- Every `SERVER_CLEANUP_INTERVAL` (default `10m`) a loop removes expired snapshots: both blobs first (`BlobKey` and `LogsBlobKey`; a missing blob is fine), then the row. The resources and logs rows go with `ON DELETE CASCADE`.

## Out of scope

- CLI flags and `kubin delete` (TASK-45), the UI (TASK-46), auth.
- The tarball format. Expiry is not in the manifest, so the format version doesn't change.
- The upload response is a contract with the CLI: only add fields.

## Traps

- Store only a SHA-256 hash of the delete token and compare in constant time. Make the token with `crypto/rand`, like `NewID`.
- Snapshots stored before this migration have no expiry. NULL `expires_at` means it never expires, and the job skips it. Document that.
- With several replicas every one runs the job at once, so each step must be idempotent. `Blobs.Delete` of a missing key is already not an error.
- Delete blobs before the row. A failure then leaves a row the next run retries, never a blob nothing points at.
- Remove expired snapshots in batches (a LIMIT per pass), so a backlog doesn't become one long transaction.
- Inject the clock (`Service.Now` exists) so tests don't sleep.
- Until TASK-46 lands, the UI shows a 410 as a generic error (`ui/src/lib/api.ts` maps only 404). That's fine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An upload gets its expiry from the request, or the server default when none is given, and is refused with 400 above SERVER_SNAPSHOT_MAX_TTL
- [ ] #2 The upload response has expires_at and delete_token, and GET /api/v1/snapshots/{id} has expires_at
- [ ] #3 Every read route of an expired snapshot returns 410 with a JSON error
- [ ] #4 The cleanup job removes expired snapshots, their rows and both blobs, and leaves unexpired and no-expiry snapshots alone (test with an injected clock)
- [ ] #5 DELETE /api/v1/snapshots/{id} with the right bearer token removes the snapshot and its blobs with 204; a wrong or missing token gives 403 and removes nothing
- [ ] #6 Only a hash of the delete token is stored (test reads the row)
- [ ] #7 docs/architecture.md (API, Flow) and docs/self-hosting.md (SERVER_SNAPSHOT_TTL, SERVER_SNAPSHOT_MAX_TTL, SERVER_CLEANUP_INTERVAL) describe it
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd server && make test && make lint passes (CI=true make test too if migrations/postgres exists)
- [ ] #2 Manual: cd server && make run; write a tarball with cd cli && go run . create -o /tmp/s.tar.gz against a kind cluster; curl --data-binary @/tmp/s.tar.gz -H "Content-Type: application/gzip" "localhost:8080/api/v1/snapshots?expires_in=1m"; the snapshot answers 200, then 410 after a minute, and its row and blobs are gone after the next cleanup pass
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-14: a snapshot now has two blobs, snapshots.blob_key (tarball) and snapshots.logs_blob_key (logs, may be empty for old snapshots). Delete both; the logs rows go with ON DELETE CASCADE.
<!-- SECTION:NOTES:END -->
