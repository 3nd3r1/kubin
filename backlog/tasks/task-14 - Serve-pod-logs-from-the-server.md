---
id: TASK-14
title: Serve pod logs from the server
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-17T14:08:00Z'
labels:
  - server
milestone: m-0
dependencies:
  - TASK-4
  - TASK-11
priority: high
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Logs stay in the blob store per decision-1, so the UI needs an endpoint to stream them out of the stored tarball or an extracted copy.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `GET /api/v1/snapshots/{id}/logs/{ns}/{pod}/{container}` streams the log
- [ ] #2 A `previous=true` parameter returns the previous container log when present
- [ ] #3 A `q` parameter returns only matching lines
- [ ] #4 Large logs stream without loading the file into memory
<!-- AC:END -->
