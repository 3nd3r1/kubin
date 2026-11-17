---
id: TASK-21
title: Expire and delete old snapshots
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-17T10:11:00Z'
labels:
  - server
milestone: m-3
dependencies:
  - TASK-3
priority: high
ordinal: 21000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Cluster dumps should not live forever, both for cost and because they hold cluster details.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Snapshots get an expiry at upload, with a default and a CLI flag within a server maximum
- [ ] #2 Expired snapshots return 410 and are removed from the database and the blob store by a periodic job inside the server
- [ ] #3 `kubin delete <url>` removes a snapshot using a delete token returned at upload
<!-- AC:END -->
