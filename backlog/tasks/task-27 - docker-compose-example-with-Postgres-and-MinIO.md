---
id: TASK-27
title: docker compose example
status: To Do
assignee: []
created_date: '2026-11-17T10:11:00Z'
updated_date: '2026-11-24T10:31:00Z'
labels:
  - deploy
  - tier:1
milestone: m-3
dependencies: []
priority: medium
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
docs/self-hosting.md says a compose file comes later. Some people run everything on a VM from compose files and want one to copy. Per decision-12 it runs kubin alone on SQLite and a named volume, like the `docker run` setup in docs/self-hosting.md: Postgres and S3 left m-3.

## Where

- New `deploy/compose/docker-compose.yml`.
- `docs/self-hosting.md`: a compose section after docker run, replacing the "come later" sentence for compose. `docs/architecture.md`: drop the "(TASK-27)" note in the Deploying table.

## Spec

- One service, `kubin`: `image: ${KUBIN_IMAGE:-ghcr.io/3nd3r1/kubin:latest}`, port 8080 published, a named volume at `/data`, `restart: unless-stopped`.
- Env vars from docs/self-hosting.md (Configuration) are shown as commented examples, `SERVER_PUBLIC_URL` among them.

## Traps

- The image is `scratch`: no shell, so no `healthcheck` inside it. Check `/readyz` from the host.
- A named volume at `/data` gets uid 1001 from the image. A bind mount would not (docs/self-hosting.md).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 In deploy/compose, `docker compose up -d` runs kubin and `curl -fsS localhost:8080/readyz` prints ok
- [ ] #2 Snapshots persist in the named volume across `docker compose down` and `docker compose up -d`
- [ ] #3 docs/self-hosting.md covers the compose setup and docs/architecture.md no longer marks it as TASK-27
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 docker compose -f deploy/compose/docker-compose.yml config -q passes
- [ ] #2 Manual: make image, then in deploy/compose KUBIN_IMAGE=kubin:dev docker compose up -d, curl -fsS localhost:8080/readyz, upload with cd cli && go run . create --server http://localhost:8080 -n kube-system against a kind cluster, then docker compose down && docker compose up -d and the link still loads
<!-- DOD:END -->
