---
id: TASK-26
title: Postgres as an optional database
status: To Do
assignee: []
created_date: '2026-11-17T10:11:00Z'
labels:
  - server
milestone: m-3
dependencies:
  - TASK-3
priority: medium
ordinal: 26000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
SQLite is the default (decision-2) but cannot be shared by several server replicas. Larger installs and the Helm chart need Postgres.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A database URL in config picks Postgres, otherwise SQLite is used
- [ ] #2 The same migrations and store tests run against both
- [ ] #3 docs/self-hosting.md explains when to use which
<!-- AC:END -->
