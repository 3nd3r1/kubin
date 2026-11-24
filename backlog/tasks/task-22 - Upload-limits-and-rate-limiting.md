---
id: TASK-22
title: Upload limits and rate limiting
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-24T09:37:00Z'
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
- [ ] #1 An archive that unpacks past SERVER_MAX_UNPACKED_BYTES gets 413 with a JSON error naming the limit; the compressed limit still gives 413
- [ ] #2 Uploads over the limit from one client get 429 with Retry-After and a JSON error before the body is read; other clients are unaffected; a rate of 0 turns the limit off
- [ ] #3 From a proxy in SERVER_TRUSTED_PROXIES the client comes from X-Forwarded-For; from anywhere else the header is ignored
- [ ] #4 An upload may run up to SERVER_UPLOAD_TIMEOUT, while other routes keep SERVER_READ_TIMEOUT and SERVER_WRITE_TIMEOUT (test with a slow body)
- [ ] #5 Idle clients are evicted from the limiter, so its memory stays bounded (test)
- [ ] #6 docs/self-hosting.md documents the new variables and says the limits are per replica
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd server && make test && make lint passes
<!-- DOD:END -->
