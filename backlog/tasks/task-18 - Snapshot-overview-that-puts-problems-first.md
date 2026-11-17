---
id: TASK-18
title: Snapshot overview that puts problems first
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
labels:
  - ui
milestone: m-2
dependencies:
  - TASK-15
priority: medium
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Someone opening a link usually wants to know what is wrong. The landing page of a snapshot should show that before they start clicking.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The overview shows cluster version, capture time, node count and resource counts
- [ ] #2 Pods not Ready, in CrashLoopBackOff or with restarts are listed first
- [ ] #3 Warning events from the last hour before capture are listed
- [ ] #4 Every item links to its detail
<!-- AC:END -->
