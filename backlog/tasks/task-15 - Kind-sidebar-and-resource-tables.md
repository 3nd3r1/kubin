---
id: TASK-15
title: Kind sidebar and resource tables
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
labels:
  - ui
milestone: m-2
dependencies:
  - TASK-6
  - TASK-10
priority: high
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Lens is the reference for how browsing a cluster should feel. With more kinds in the snapshot, the m-0 namespace/pod page stops scaling.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A sidebar groups kinds as Lens does (Workloads, Network, Config, Storage, Cluster) with counts
- [ ] #2 Each kind has a table with the columns kubectl get shows for it
- [ ] #3 A namespace filter applies across all tables and is kept in the URL
- [ ] #4 Every view has a shareable URL
<!-- AC:END -->
