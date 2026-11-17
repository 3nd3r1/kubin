---
id: TASK-6
title: Show namespaces, pods and pod details for a snapshot in the UI
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T10:11:00Z'
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
The link has to open something. This is the first real page of the embedded UI (TASK-24); the full Lens-like layout comes in m-2.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `/s/{id}` shows cluster name, capture time and namespaces
- [ ] #2 Selecting a namespace lists its pods with phase, restarts and age at capture time
- [ ] #3 Selecting a pod shows its containers, status and full JSON or YAML
- [ ] #4 An unknown ID shows a not-found page
- [ ] #5 The UI calls the API on its own origin
<!-- AC:END -->
