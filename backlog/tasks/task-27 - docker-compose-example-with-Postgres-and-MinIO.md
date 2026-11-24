---
id: TASK-27
title: docker compose example with Postgres and MinIO
status: To Do
assignee: []
created_date: '2026-11-17T10:11:00Z'
updated_date: '2026-11-24T09:37:00Z'
labels:
  - deploy
  - tier:2
milestone: m-3
dependencies:
  - TASK-20
  - TASK-26
priority: medium
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The middle self-hosting option in docs/architecture.md (Deploying): a small company on one VM, with Postgres and S3-compatible storage instead of SQLite and a local directory. It uses the config TASK-26 (`SERVER_DATABASE_URL`) and TASK-20 (`SERVER_S3_*`) added; read `docs/self-hosting.md` for the final names.

## Where

- New `deploy/compose/docker-compose.yml` and `deploy/compose/.env.example`.
- `docs/self-hosting.md`: replace "Docker compose ... come later" with a compose section. `docs/architecture.md`: drop the "(TASK-27)" note.

## Spec

- `kubin`: `image: ${KUBIN_IMAGE:-ghcr.io/3nd3r1/kubin:latest}`, port 8080 published, a named volume at `/data` (uploads spool to `/data/tmp` even with Postgres and S3), `depends_on` Postgres healthy and the bucket job completed successfully.
- `postgres`: `postgres:17-alpine`, a named volume, a `pg_isready` healthcheck.
- `minio`: the same pinned `minio/minio:RELEASE.*` tag TASK-20's tests use (its implementation notes name it), a named volume.
- A one-shot service with a pinned `minio/mc` image that creates the bucket. The server never creates buckets (TASK-20).
- Passwords and keys come from `.env`. The committed `.env.example` has placeholders, and the docs say to copy it and change them.
- Only kubin's port is published. Postgres and MinIO stay on the compose network.

## Out of scope

The Helm chart (TASK-7), TLS and a reverse proxy (the docs can point at "Behind a proxy"), backups.

## Traps

- The kubin image is `scratch`: no shell, so no `healthcheck` inside it. Check `/readyz` from the host.
- MinIO stopped publishing community images in late 2025, and `minio/mc` too. Pin tags that `docker pull` still fetches.
- A named volume at `/data` gets uid 1001 from the image. A bind mount wouldn't (docs/self-hosting.md).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 After `cp .env.example .env`, `docker compose up -d` in deploy/compose runs kubin, Postgres and MinIO, and kubin /readyz answers ok
- [ ] #2 kubin stores into Postgres and MinIO, not SQLite or the local directory: after an upload the snapshot row is in Postgres and both blobs are in the bucket
- [ ] #3 Data persists in named volumes across `docker compose down` and `docker compose up -d`
- [ ] #4 Only kubin publishes a port
- [ ] #5 docs/self-hosting.md covers the compose setup
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 docker compose -f deploy/compose/docker-compose.yml config -q passes
- [ ] #2 Manual: make image, then in deploy/compose cp .env.example .env && KUBIN_IMAGE=kubin:dev docker compose up -d, curl -fsS localhost:8080/readyz, upload with cd cli && go run . create --server http://localhost:8080 -n kube-system against a kind cluster, then docker compose down && docker compose up -d and the link still loads
<!-- DOD:END -->
