---
id: TASK-25
title: Ship kubin as one container that runs with docker run
status: To Do
assignee: []
created_date: '2026-11-17T10:11:00Z'
updated_date: '2026-11-17T14:00:00Z'
labels:
  - deploy
milestone: m-0
dependencies:
  - TASK-3
  - TASK-24
priority: high
ordinal: 6500
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Easy self-hosting is a main goal (doc-1, decision-2). The m-0 demo should run the way a self-hoster would run it: one image, one volume, no other services.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 One multi-stage Dockerfile builds the UI and the server into a single image
- [ ] #2 `docker run -p 8080:8080 -v kubin-data:/data <image>` serves the UI and API, with SQLite and blobs under /data
- [ ] #3 Snapshots survive a container restart
- [ ] #4 All config is env vars with defaults, documented in docs/self-hosting.md
- [ ] #5 The image runs as non-root
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-3: SERVER_DATA_DIR defaults to the relative path data; the image should set SERVER_DATA_DIR=/data and make /data writable for uid 1001 (the scratch image has no writable dirs). The server also needs a writable $SERVER_DATA_DIR/tmp for spooling uploads, which it creates itself.
<!-- SECTION:NOTES:END -->
