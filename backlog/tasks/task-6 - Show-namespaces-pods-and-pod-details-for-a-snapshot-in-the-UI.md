---
id: TASK-6
title: Snapshot page with broken pods first and pod details
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T14:09:00Z'
labels:
  - ui
milestone: m-0
dependencies:
  - TASK-4
  - TASK-24
priority: high
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The link has to open something, and per decision-3 it should open on what is broken. This is the first real page of the embedded UI (TASK-24). Events and logs for a pod come in TASK-29. The full Lens-like layout and overview come in m-2 (TASK-15, TASK-18).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `/s/{id}` shows cluster name, capture time and every pod grouped by namespace
- [ ] #2 Pods that are not Ready, are in CrashLoopBackOff or Error, or have restarts are listed first and marked
- [ ] #3 Each pod shows phase, restarts and age at capture time
- [ ] #4 Selecting a pod shows its containers, their state and last termination reason, and the full object as YAML
- [ ] #5 An unknown ID shows a not-found page
- [ ] #6 The UI calls the API on its own origin
<!-- AC:END -->
