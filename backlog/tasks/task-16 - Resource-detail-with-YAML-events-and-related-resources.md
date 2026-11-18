---
id: TASK-16
title: Resource detail with YAML, events and related resources
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-18T08:50:00Z'
labels:
  - ui
milestone: m-2
dependencies:
  - TASK-15
priority: high
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Debugging means jumping from a failing pod to its ReplicaSet, Deployment, Service and the events about each.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Detail shows the object as YAML with copy
- [ ] #2 Events whose involvedObject is the resource are listed with time, reason and message
- [ ] #3 Owner references and owned resources are links
- [ ] #4 Pods link to their node and to Services whose selector matches them
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-12: when the manifest redactions list has env-values, show env entries without value or valueFrom as redacted, not as empty. Secret data values are empty strings under secret-data.
<!-- SECTION:NOTES:END -->
