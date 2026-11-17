---
id: TASK-1
title: Merge api-gateway and shared into one server module
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T10:34:00Z'
labels:
  - server
milestone: m-0
dependencies: []
priority: high
type: chore
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-1 replaced the microservices plan with one server. Today the server is split across two Go modules (`services/api-gateway`, `services/shared`) joined by a replace directive, and the Dockerfile has to copy both. The shared module only holds a logger that is duplicated in `cli/pkg/log`. The api-gateway name and the stubbed `/internal/api/v1` routes still describe the old design.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The server lives in `server/` as one Go module with no replace directives
- [x] #2 `services/` is gone and the Dockerfile, skaffold.yaml and chart point at `server/`
- [x] #3 The stub routes are reduced to /healthz and /readyz, the snapshot routes come with TASK-3 and TASK-4
- [x] #4 Go version matches the CLI module
- [x] #5 The server shuts down gracefully: main waits for the HTTP server to finish before exiting
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. git mv services/api-gateway to server/, module github.com/3nd3r1/kubin/server, go 1.24 like cli.
2. Replace the shared zap logger with log/slog inside the server (shared only held the logger), delete services/shared.
3. Drop the router wrapper, the 501 snapshot handlers, domain stub and unused auth/cors middleware. Keep /healthz and /readyz on a chi router.
4. Config loaded once in main and passed down instead of package init globals. Keep env names so the chart still works.
5. server.Run(ctx) blocks until Shutdown finishes, main waits on it.
6. Dockerfile for one module, drop HEALTHCHECK that calls a missing subcommand (TASK-7/25 own probes). skaffold context to server/.
7. Tests for health routes and graceful shutdown. Makefile. Update docs paths.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The chart has no build paths. Its only services/ reference is the disabled ArgoCD app-of-apps template, and the subchart is still named api-gateway. Both left to TASK-7, which rebuilds the chart as one deployment. The Dockerfile HEALTHCHECK called a subcommand that never existed and was dropped; probes belong to TASK-7 and TASK-25. Env var names (SERVER_PORT, SERVER_ENV, LOG_LEVEL) kept so the chart values still apply. Timeouts are now durations (10s) instead of integer seconds.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Moved services/api-gateway to server/ as one Go module (go 1.24, no replace) and deleted services/shared. The zap logger is replaced by log/slog. Removed the router wrapper, the 501 snapshot handlers, the domain stub and unused auth/CORS middleware; the router now has /healthz and /readyz on chi with recovery, request logging and secure headers in production. Config is loaded in main and passed down, and parse errors now fail startup instead of being logged and ignored. server.Serve blocks until Shutdown finishes, so main waits for in-flight requests. Dockerfile and skaffold build from server/. Verified: go vet, golangci-lint (0 issues), go test -race -count=3 with tests for health routes, 404 on old routes, secure headers, a graceful-shutdown test with an in-flight request and listener errors. Built the image, curl /healthz and /readyz 200, old route 404, docker stop exits 0 with a stopped log line.
<!-- SECTION:FINAL_SUMMARY:END -->
