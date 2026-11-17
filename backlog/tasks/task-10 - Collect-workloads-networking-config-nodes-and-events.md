---
id: TASK-10
title: Collect workloads, networking, config and nodes
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-17T14:08:00Z'
labels:
  - cli
milestone: m-1
dependencies:
  - TASK-2
  - TASK-12
  - TASK-28
priority: high
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Pods, namespaces and events alone do not explain a broken cluster. Debugging needs owners (Deployments, ReplicaSets, StatefulSets, DaemonSets, Jobs, CronJobs), Services, Endpoints, Ingresses, ConfigMaps, PVCs, PVs and Nodes. Writing a collector per kind does not scale, so collection should go through discovery and the dynamic client with an allowlist. Events were pulled into m-0 as TASK-28 for the demo (decision-3) and move into the allowlist here. New kinds go through the redaction from TASK-12.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All listed kinds are in the snapshot with correct group and version paths
- [ ] #2 Kinds the user has no RBAC for are skipped with a warning, not a failure
- [ ] #3 The kind list is one place in the code
- [ ] #4 Tests cover discovery, allowlist and the RBAC skip with a fake client
<!-- AC:END -->
