---
id: TASK-22
title: Upload limits and rate limiting
status: Done
assignee:
  - '@agy'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-25T10:04:00Z'
labels:
  - server
  - tier:3
milestone: m-3
dependencies:
  - TASK-3
priority: medium
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The public instance (TASK-23) puts `POST /api/v1/snapshots` on the internet. Some limits exist: the compressed size (`SERVER_MAX_UPLOAD_BYTES`, 413) and the unpacked size (`snapshot.DefaultMaxUnpacked`, 1 GiB, a constant, rejected as a 400 "invalid snapshot"). Missing: a configurable unpacked limit that answers 413, a per-IP rate limit, and an upload timeout. `SERVER_READ_TIMEOUT` (10s) covers the whole upload today, so a 64 MiB snapshot needs more than 50 Mbit/s to get through.

## Where

- `server/internal/snapshot/format.go`: the `invalid("archive unpacks to more than %d bytes")` check. `snapshot.go`: `Service.MaxUnpacked`.
- `server/internal/api/snapshots.go` (`createSnapshot`) and `router.go`.
- `server/internal/config/config.go`, `server/cmd/main.go`, `docs/self-hosting.md`.

## Spec

- `SERVER_MAX_UNPACKED_BYTES`, default `1073741824`, feeds `Service.MaxUnpacked`. Going over it returns 413 with a JSON error naming the limit. Use an error type of its own, not `InvalidError`.
- `SERVER_UPLOAD_RATE` (uploads per client IP per hour, default 30) and `SERVER_UPLOAD_BURST` (default 10). 0 turns the limit off. Over the limit: 429, a `Retry-After` header, and a JSON error like "too many uploads from your address, try again in 2m". Use `golang.org/x/time/rate`, one limiter per client key, idle entries evicted so memory stays bounded.
- The client key is the IP from `RemoteAddr`, or the /64 prefix for IPv6. `SERVER_TRUSTED_PROXIES` (comma-separated CIDRs): when `RemoteAddr` is inside one, use the right-most `X-Forwarded-For` address that is not a trusted proxy. Otherwise ignore the header. Don't use chi's `middleware.RealIP`, which trusts any client.
- `SERVER_UPLOAD_TIMEOUT`, default `5m`: the upload handler extends its read and write deadlines with `http.NewResponseController`. Other routes keep `SERVER_READ_TIMEOUT` and `SERVER_WRITE_TIMEOUT`.
- Limits live in memory per replica. The docs say so. No Redis or shared store (CLAUDE.md).

## Out of scope

Rate limits on reads, auth, the CLI's `--max-size`, expiry (TASK-21). The CLI already shows the server's error message (`upload.serverMessage`), so it needs no change.

## Traps

- The write deadline counts from when the request headers were read, so after a long upload the response hits `WriteTimeout`. Extend both deadlines.
- Check the rate limit before reading the body, so a refused client doesn't send 64 MiB first.
- Test the limiter with an injected clock and `httptest` requests with `RemoteAddr` set.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 An archive that unpacks past SERVER_MAX_UNPACKED_BYTES gets 413 with a JSON error naming the limit; the compressed limit still gives 413
- [x] #2 Uploads over the limit from one client get 429 with Retry-After and a JSON error before the body is read; other clients are unaffected; a rate of 0 turns the limit off
- [x] #3 From a proxy in SERVER_TRUSTED_PROXIES the client comes from X-Forwarded-For; from anywhere else the header is ignored
- [x] #4 An upload may run up to SERVER_UPLOAD_TIMEOUT, while other routes keep SERVER_READ_TIMEOUT and SERVER_WRITE_TIMEOUT (test with a slow body)
- [x] #5 Idle clients are evicted from the limiter, so its memory stays bounded (test)
- [x] #6 docs/self-hosting.md documents the new variables and says the limits are per replica
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 cd server && make test && make lint passes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add golang.org/x/time/rate to server/go.mod.
2. In server/internal/snapshot/format.go, introduce TooLargeError (distinct from InvalidError) for unpacked size limit violations and guard against overflow/negative sizes. Update snapshot_test.go to verify TooLargeError.
3. Add SERVER_MAX_UNPACKED_BYTES, SERVER_UPLOAD_RATE, SERVER_UPLOAD_BURST, SERVER_TRUSTED_PROXIES, and SERVER_UPLOAD_TIMEOUT to server/internal/config/config.go, with CIDR parsing/validation for SERVER_TRUSTED_PROXIES.
4. Implement an in-memory per-client upload rate limiter in server/internal/api/ratelimit.go using golang.org/x/time/rate, IPv4 / IPv6 /64 keying, trusted-proxy X-Forwarded-For resolution, Retry-After calculation, injectable clock, and lazy eviction of idle clients.
5. Wire unpacked TooLargeError -> 413, rate limiting (before reading body) -> 429 with Retry-After, and UploadTimeout deadline extension via http.NewResponseController in server/internal/api/snapshots.go, router.go, and server/cmd/main.go.
6. Add unit and integration tests for unpacked 413, rate limiting (burst/429/body unread/other clients unaffected/rate=0 disabled), trusted proxy XFF handling, idle client eviction, and slow-body upload timeout vs normal route timeouts.
7. Update docs/self-hosting.md with the new environment variables, trusted proxies, and per-replica limit note.
8. Run cd server && make test && make lint and verify all acceptance criteria.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added configurable unpacked archive limits (SERVER_MAX_UNPACKED_BYTES -> 413 via snapshot.TooLargeError), per-client-IP upload rate limiting (SERVER_UPLOAD_RATE, SERVER_UPLOAD_BURST, SERVER_TRUSTED_PROXIES with IPv6 /64 grouping, right-most untrusted X-Forwarded-For resolution, Retry-After header, body-unread 429 rejection, and lazy idle client eviction), and upload-only read/write deadline extension (SERVER_UPLOAD_TIMEOUT via http.NewResponseController). Documented all new variables and per-replica limit semantics in docs/self-hosting.md. Verified with cd server && make test && make lint and go test -race -count=1 ./... (all passing, 0 lint issues).
<!-- SECTION:FINAL_SUMMARY:END -->
