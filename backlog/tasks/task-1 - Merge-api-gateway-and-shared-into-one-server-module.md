---
id: TASK-1
title: Merge api-gateway and shared into one server module
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
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
- [ ] #1 The server lives in `server/` as one Go module with no replace directives
- [ ] #2 `services/` is gone and the Dockerfile, skaffold.yaml and chart point at `server/`
- [ ] #3 The stub routes are reduced to /healthz and /readyz, the snapshot routes come with TASK-3 and TASK-4
- [ ] #4 Go version matches the CLI module
- [ ] #5 The server shuts down gracefully: main waits for the HTTP server to finish before exiting
<!-- AC:END -->
