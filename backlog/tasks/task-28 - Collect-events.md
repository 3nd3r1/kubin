---
id: TASK-28
title: Collect events
status: To Do
assignee: []
created_date: '2026-11-17T14:08:00Z'
labels:
  - cli
milestone: m-0
dependencies:
  - TASK-2
priority: high
ordinal: 28000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The demo (decision-3) needs the events of a crashing pod, like Back-off restarting failed container, because they say what Kubernetes saw before anyone opens the logs. This pulls events forward from TASK-10. When TASK-10 moves collection to discovery and an allowlist, events move into that list.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Events from core/v1 in every collected namespace are in the snapshot under the v1 layout
- [ ] #2 A user without RBAC for events gets a warning and a snapshot without them, not a failure
- [ ] #3 Tests use a fake client
<!-- AC:END -->
