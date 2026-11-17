---
id: TASK-4
title: Serve the snapshot read API
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
labels:
  - server
milestone: m-0
dependencies:
  - TASK-3
priority: high
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The UI needs to read a snapshot back. The endpoints are listed in docs/architecture.md under API.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `GET /api/v1/snapshots/{id}` returns the manifest and counts
- [ ] #2 `GET /api/v1/snapshots/{id}/resources` lists resources, filterable by kind and namespace
- [ ] #3 `GET /api/v1/snapshots/{id}/resources/{kind}/{ns}/{name}` returns one object
- [ ] #4 Unknown IDs return 404 and CORS allows the UI origin from config
- [ ] #5 Handlers are tested against the fixture from TASK-3
<!-- AC:END -->
