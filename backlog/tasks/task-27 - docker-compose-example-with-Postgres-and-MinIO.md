---
id: TASK-27
title: docker compose example with Postgres and MinIO
status: To Do
assignee: []
created_date: '2026-11-17T10:11:00Z'
labels:
  - deploy
milestone: m-3
dependencies:
  - TASK-20
  - TASK-26
priority: medium
ordinal: 27000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The middle self-hosting option between docker run and Helm: a small company running kubin on one VM with real storage.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `deploy/compose/docker-compose.yml` runs kubin, Postgres and MinIO with `docker compose up`
- [ ] #2 Data persists in named volumes
- [ ] #3 docs/self-hosting.md covers it
<!-- AC:END -->
