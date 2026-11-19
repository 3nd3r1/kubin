---
id: TASK-10
title: Collect workloads, networking, config and nodes
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-19T09:04:00Z'
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
Pods, namespaces and events alone do not explain a broken cluster. Debugging needs owners (Deployments, ReplicaSets, StatefulSets, DaemonSets, Jobs, CronJobs), Services, EndpointSlices, Ingresses, ConfigMaps, PVCs, PVs and Nodes. Writing a collector per kind does not scale, so collection should go through discovery and the dynamic client with an allowlist. Events were pulled into m-0 as TASK-28 for the demo (decision-3) and move into the allowlist here. New kinds go through the redaction from TASK-12.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All listed kinds are in the snapshot with correct group and version paths
- [x] #2 Kinds the user has no RBAC for are skipped with a warning, not a failure
- [x] #3 The kind list is one place in the code
- [x] #4 Tests cover discovery, allowlist and the RBAC skip with a fake client
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. kube.Client gets discovery (preferred resources) and a paginated dynamic list; drop the typed namespace/pod/event lists.
2. collector.ResourcesCollector replaces CoreCollector and EventsCollector: one Kinds allowlist (GroupKinds), version from discovery, Forbidden lists skipped with a warning, partial discovery failures warned.
3. kube/kubetest builds a typed + dynamic fake with built-in discovery for tests.
4. Update manager/create tests, docs (architecture, snapshot-format coverage note). No format bump: the layout is unchanged.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-12: every collected object already goes through redact.Redactor in snapshot.Manager, so Secrets collected here get secret-data, and workload templates get env-values, with no extra work. ConfigMaps and container command/args are not redacted; say in the collector how that is handled (TASK-12 docs list them as not covered).

Claude, 2026-11-19: collector.ResourcesCollector replaces CoreCollector and EventsCollector. collector.Kinds is the one allowlist (GroupKinds); the version comes from discovery.ServerPreferredResources, subresources and kinds without list are ignored, kinds the cluster doesn't serve are left out. A Forbidden list is skipped with a warning; a partial discovery failure warns and keeps the groups it got; any other error fails. kube.Client lists through the dynamic client in pages of 500. EndpointSlices instead of core Endpoints: Endpoints are deprecated since v1.33 and client-go prints a warning on every list. Secrets are not collected. ConfigMaps go out unredacted, as snapshot-format.md already says. New kube/kubetest builds typed + dynamic fakes with plain-cluster discovery. No format bump: the layout did not change.
Verified on the kind-kubin-demo cluster (v1.37): 166 resources across 10 kinds, no deprecation warnings; as a service account allowed only pods, namespaces and deployments, every other kind printed one skip warning and the snapshot was written (exit 0); uploaded to a local server, and /resources?kind=Deployment returns the objects. go test ./... and golangci-lint pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
kubin create now collects Nodes, workloads (Deployments, ReplicaSets, StatefulSets, DaemonSets, Jobs, CronJobs), Services, EndpointSlices, Ingresses, ConfigMaps, PVCs and PVs as well as namespaces, pods and events. Everything goes through discovery and the dynamic client, driven by one allowlist in collector.Kinds. Kinds the user can't list are skipped with a warning. Tests use a fake cluster with plain-cluster discovery and cover the allowlist, the RBAC skip, partial and failed discovery. It was checked against kind with full and restricted RBAC and uploaded to a real server. README and architecture are updated. Follow-up: the snapshot page still shows only pods (TASK-15).
<!-- SECTION:FINAL_SUMMARY:END -->
