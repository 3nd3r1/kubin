---
id: TASK-3
title: Store and index uploaded snapshots in the server
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T10:11:00Z'
labels:
  - server
milestone: m-0
dependencies:
  - TASK-1
  - TASK-2
priority: high
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Nothing can be shared until the server accepts a snapshot. Per decision-1 upload is synchronous: validate the tarball, keep it in the blob store, index each resource in the database. Per decision-2 the default database is SQLite so a self-hoster needs nothing but a volume; Postgres comes later as TASK-26, so keep SQL portable. The snapshot ID is the only access control until there is auth, so it has to be unguessable.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `POST /api/v1/snapshots` takes a v1 tarball and returns `{id, url}`
- [ ] #2 IDs are random with at least 96 bits of entropy
- [ ] #3 The raw tarball is stored through a blob interface with a filesystem implementation
- [ ] #4 A store interface backed by SQLite (pure Go driver, no CGO) holds a snapshots table and a resources table (snapshot, kind, namespace, name, JSON), created by migrations the server runs on start
- [ ] #5 Unknown format versions, oversized uploads and invalid archives return 4xx with a message
- [ ] #6 An integration test uploads a fixture tarball against a temporary SQLite file
<!-- AC:END -->
