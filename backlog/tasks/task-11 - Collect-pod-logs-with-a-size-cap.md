---
id: TASK-11
title: Collect pod logs with a size cap
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
labels:
  - cli
milestone: m-1
dependencies:
  - TASK-2
priority: high
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Logs are usually the reason someone shares a cluster. `kube.Client.GetPodLogs` exists but nothing calls it, and it reads the full log of one container only. A large cluster can produce gigabytes of logs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Current logs for every container, including init containers, land in logs/<ns>/<pod>/<container>.log
- [ ] #2 Previous logs are collected for containers that restarted
- [ ] #3 Each container log is capped by a tail limit with a flag to change it, and truncation is recorded in the manifest
- [ ] #4 Logs are fetched in parallel with a bounded worker count
<!-- AC:END -->
