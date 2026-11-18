---
id: TASK-25
title: Ship kubin as one container that runs with docker run
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:11:00Z'
updated_date: '2026-11-18T08:25:00Z'
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
- [x] #1 One multi-stage Dockerfile builds the UI and the server into a single image
- [x] #2 `docker run -p 8080:8080 -v kubin-data:/data <image>` serves the UI and API, with SQLite and blobs under /data
- [x] #3 Snapshots survive a container restart
- [x] #4 All config is env vars with defaults, documented in docs/self-hosting.md
- [x] #5 The image runs as non-root
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Root Dockerfile with three stages: node builds ui/ into server/internal/web/dist, golang builds the server with the UI embedded, scratch runs it as 1001 with /data owned by 1001, SERVER_DATA_DIR=/data, SERVER_ENV=production, VOLUME /data.
2. Root .dockerignore that only sends ui/ and server/ sources. Remove server/Dockerfile and server/.dockerignore; point skaffold at the root context.
3. make image at the root.
4. docs/self-hosting.md: docker run, volume, every env var with its default, bind-mount ownership note. Update README, development.md, architecture.md.
5. Verify: build, run with a named volume, healthz/readyz, UI at / and /s/{id}, kubin create against kind, restart the container and reopen the snapshot, check uid.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-3: SERVER_DATA_DIR defaults to the relative path data; the image should set SERVER_DATA_DIR=/data and make /data writable for uid 1001 (the scratch image has no writable dirs). The server also needs a writable $SERVER_DATA_DIR/tmp for spooling uploads, which it creates itself.

Verified: docker build from the root makes a 12.6MB scratch image. docker run -p 8090:8080 -v kubin-test-data:/data served /healthz, /readyz, index.html, hashed assets and /s/{id}. kubin create against kind-kubin uploaded 118 resources; the snapshot read back after docker restart and after docker rm plus a new container on the same volume. Config.User is 1001:1001 and every file under /data is owned by 1001. A bind mount not owned by 1001 exits with mkdir /data/tmp: permission denied (documented); after chown it starts. server make test passes.
Risk: SERVER_READ_TIMEOUT 10s covers the whole upload body, so a large snapshot on a slow link times out. Relevant to TASK-22.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Root multi-stage Dockerfile: node builds the UI into server/internal/web/dist, golang builds the server with it embedded, scratch runs it as 1001 with SERVER_DATA_DIR=/data, SERVER_ENV=production and /data as a volume owned by 1001. Root .dockerignore sends only ui/ and server/. server/Dockerfile is gone; skaffold builds from the root. make image builds kubin:dev. docs/self-hosting.md covers docker run, bind-mount ownership, proxies and every env var with its default; README, development.md and architecture.md point to it. Verified end to end against kind with restarts. Follow-ups: release the image (TASK-19); the 10s read timeout may cut off slow uploads (TASK-22).
<!-- SECTION:FINAL_SUMMARY:END -->
