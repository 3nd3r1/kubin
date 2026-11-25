---
id: TASK-21
title: Expire snapshots and delete them by token in the server
status: Done
assignee:
  - '@agy'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-25T11:56:00Z'
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
- [x] #1 An upload gets its expiry from the request, or the server default when none is given, and is refused with 400 above SERVER_SNAPSHOT_MAX_TTL
- [x] #2 The upload response has expires_at and delete_token, and GET /api/v1/snapshots/{id} has expires_at
- [x] #3 Every read route of an expired snapshot returns 410 with a JSON error
- [x] #4 The cleanup job removes expired snapshots, their rows and both blobs, and leaves unexpired and no-expiry snapshots alone (test with an injected clock)
- [x] #5 DELETE /api/v1/snapshots/{id} with the right bearer token removes the snapshot and its blobs with 204; a wrong or missing token gives 403 and removes nothing
- [x] #6 Only a hash of the delete token is stored (test reads the row)
- [x] #7 docs/architecture.md (API, Flow) and docs/self-hosting.md (SERVER_SNAPSHOT_TTL, SERVER_SNAPSHOT_MAX_TTL, SERVER_CLEANUP_INTERVAL) describe it
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 cd server && make test && make lint passes (CI=true make test too if migrations/postgres exists)
- [x] #2 Manual: cd server && make run; write a tarball with cd cli && go run . create -o /tmp/s.tar.gz against a kind cluster; curl --data-binary @/tmp/s.tar.gz -H "Content-Type: application/gzip" "localhost:8080/api/v1/snapshots?expires_in=1m"; the snapshot answers 200, then 410 after a minute, and its row and blobs are gone after the next cleanup pass
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Database migration and store (server/internal/store/):
   - Add migrations/sqlite/0003_expiry_and_delete.sql (migrations/postgres/ does not exist yet): add nullable expires_at TIMESTAMP (NULL means never expires for pre-migration snapshots), delete_token_hash TEXT NOT NULL DEFAULT '' (hex SHA-256), and partial index snapshots_by_expires_at ON snapshots (expires_at) WHERE expires_at IS NOT NULL.
   - Update store.Snapshot with ExpiresAt *time.Time and DeleteTokenHash string. Add ExpiredSnapshot{ID, BlobKey, LogsBlobKey string} and Store methods ListExpired(ctx, before, limit) ([]ExpiredSnapshot, error) and DeleteSnapshot(ctx, id) error.
   - Update BeginSnapshot and GetSnapshot in sql.go/read.go for the new columns (sql.NullTime for expires_at). Normalize ExpiresAt and ListExpired's before with .UTC().Round(0) so SQLite's time.Time.String() lexicographical expires_at <= $1 comparison is exact and portable to Postgres.
   - ListExpired selects WHERE expires_at IS NOT NULL AND expires_at <= $1 ORDER BY expires_at ASC LIMIT $2. DeleteSnapshot executes DELETE FROM snapshots WHERE id = $1 on wdb (idempotent; resources and logs cascade via ON DELETE CASCADE).

2. Configuration (server/internal/config/):
   - Add SnapshotTTL (SERVER_SNAPSHOT_TTL, default 168h), SnapshotMaxTTL (SERVER_SNAPSHOT_MAX_TTL, default 720h), and CleanupInterval (SERVER_CLEANUP_INTERVAL, default 10m).
   - Validate in Load() that all three are > 0 and SnapshotTTL <= SnapshotMaxTTL.

3. Snapshot service and cleanup loop (server/internal/snapshot/):
   - Add DefaultTTL (168h), DefaultMaxTTL (720h), DefaultCleanupBatch (100), ErrInvalidDeleteToken, and UploadResult{ID, ExpiresAt, DeleteToken}. Add DefaultTTL, MaxTTL, CleanupBatch fields on Service.
   - Add NewDeleteToken() (32 random bytes via crypto/rand, base64.RawURLEncoding), HashDeleteToken(token) (lowercase hex SHA-256), and VerifyDeleteToken(storedHash, token) (subtle.ConstantTimeCompare; rejects empty hash or token).
   - Update Service.Upload(ctx, r, ttl) to compute uploadedAt := now().UTC().Round(0) and expiresAt := uploadedAt.Add(ttl), store only HashDeleteToken(token), and return UploadResult.
   - Add deleteSnapshot helper that deletes BlobKey (if non-empty), then LogsBlobKey (if non-empty, avoiding FS.Delete("") error on pre-TASK-14 snapshots), then Store.DeleteSnapshot(ctx, id) so blob failures leave a retryable row and never orphan blobs.
   - Add Service.Delete(ctx, id, token): loads snapshot (store.ErrNotFound -> 404), verifies token (ErrInvalidDeleteToken -> 403), then calls deleteSnapshot (works on already-expired unswept snapshots too).
   - Add cleanup.go with Service.DeleteExpired(ctx) (draining in CleanupBatch-sized passes, stopping the current pass on error so a failing blob doesn't infinite-loop) and Service.RunCleanup(ctx, interval, log) using time.NewTicker and stopping promptly on ctx.Done().

4. HTTP API and router (server/internal/api/):
   - Extend Uploader with Upload(ctx, r, ttl) (snapshot.UploadResult, error) and Delete(ctx, id, token) error (update stubUploader in ratelimit_test.go).
   - Add SnapshotTTL and SnapshotMaxTTL to Options (defaulting to snapshot defaults when 0). Reuse Options.Now for expiry checks as well as rate limiting.
   - Register DELETE /api/v1/snapshots/{id} in router.go.
   - POST /api/v1/snapshots: after rate-limit check and before reading the body, parse optional ?expires_in= (Go duration); reject invalid, <= 0, or > SnapshotMaxTTL with 400 JSON error naming the maximum. Add expires_at (RFC 3339 UTC) and delete_token to uploadResponse.
   - Read routes: add ExpiresAt *time.Time (omitempty) to snapshotResponse in GET /api/v1/snapshots/{id}. In h.snapshot(w, r) (shared by all four read routes: metadata, list resources, get resource, get log), return 410 {"error":"snapshot expired"} when snap.ExpiresAt != nil && !h.now().Before(*snap.ExpiresAt).
   - DELETE /api/v1/snapshots/{id}: require Authorization: Bearer <token> (case-insensitive scheme); missing/malformed -> 403 without touching storage; ErrNotFound -> 404; ErrInvalidDeleteToken -> 403; success -> 204 No Content.

5. Server wiring (server/cmd/main.go):
   - Wire DefaultTTL/MaxTTL into snapshot.Service and SnapshotTTL/SnapshotMaxTTL into api.Options.
   - Start snapshots.RunCleanup in a goroutine with a cancellable child context; cancel and await it in a defer registered after defer db.Close() so cleanup exits cleanly before SQLite closes.

6. Documentation (docs/architecture.md, docs/self-hosting.md):
   - docs/architecture.md: update Flow step 6 and API section with ?expires_in=, expires_at, delete_token, DELETE /api/v1/snapshots/{id}, 410-then-404 semantics, and NULL expires_at behavior.
   - docs/self-hosting.md: document SERVER_SNAPSHOT_TTL, SERVER_SNAPSHOT_MAX_TTL, SERVER_CLEANUP_INTERVAL and note that pre-migration snapshots (NULL expires_at) never expire.

7. Tests and verification:
   - store: migration/reopen, expires_at + delete_token_hash round-trip (non-null and NULL), ListExpired boundary/monotonic/timezone/NULL/limit ordering, DeleteSnapshot cascade (resources + logs) and idempotency.
   - config: defaults and invalid TTL/interval rejection.
   - snapshot: raw DB check that only SHA-256 hash is stored, Delete valid/wrong/empty/missing + empty LogsBlobKey + blob-failure-keeps-row, DeleteExpired with injected clock (batches, leaves unexpired & NULL-expiry untouched), RunCleanup cancellation.
   - api: upload default/custom/invalid/over-max expires_in (without reading body), upload and GET metadata expires_at (including NULL omission), all 4 read routes returning 410 when expired and 404 after removal, DELETE 204/403/404.
   - Run cd server && make test && make lint (and go test -race) plus DoD #2 manual verification after plan approval.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-14: a snapshot now has two blobs, snapshots.blob_key (tarball) and snapshots.logs_blob_key (logs, may be empty for old snapshots). Delete both; the logs rows go with ON DELETE CASCADE.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented server-side snapshot expiration, periodic cleanup, and bearer-token deletion:
- Store & migration (server/internal/store/): added 0003_expiry_and_delete.sql (nullable expires_at, delete_token_hash, partial index), updated Snapshot/BeginSnapshot/GetSnapshot with UTC round(0) timestamp normalization, and added ListExpired and idempotent DeleteSnapshot.
- Config (server/internal/config/): added SERVER_SNAPSHOT_TTL (168h), SERVER_SNAPSHOT_MAX_TTL (720h), and SERVER_CLEANUP_INTERVAL (10m) with validation in Load().
- Snapshot service & cleanup (server/internal/snapshot/): added UploadResult, crypto/rand delete token generation, SHA-256 hashing and constant-time verification, Service.Delete, blob-before-row deleteSnapshot helper, batched Service.DeleteExpired, and Service.RunCleanup.
- HTTP API & main (server/internal/api/, server/cmd/main.go): added ?expires_in= validation before reading the upload body, expires_at and delete_token in upload responses, expires_at in GET /api/v1/snapshots/{id} (omitted when NULL), 410 on expired snapshots across all four read routes, DELETE /api/v1/snapshots/{id} with bearer token auth (204/403/404), and graceful cleanup goroutine lifecycle before db.Close().
- Docs: updated docs/architecture.md and docs/self-hosting.md.
- Verified with make test, make lint (0 issues), go test -race -count=1 ./..., and manual kind-kubin-demo end-to-end check.
<!-- SECTION:FINAL_SUMMARY:END -->
