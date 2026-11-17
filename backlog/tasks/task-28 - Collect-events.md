---
id: TASK-28
title: Collect events
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T14:08:00Z'
updated_date: '2026-11-17T14:55:00Z'
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
- [x] #1 Events from core/v1 in every collected namespace are in the snapshot under the v1 layout
- [x] #2 A user without RBAC for events gets a warning and a snapshot without them, not a failure
- [x] #3 Tests use a fake client
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. kube.Client gets GetEvents (core/v1, all namespaces when namespace is empty).
2. collector.EventsCollector lists events across all namespaces (the CLI collects every namespace) and returns them as core/v1 Event resources. A Forbidden error logs a warning to stderr and returns no events; other errors fail the snapshot.
3. snapshot.Manager runs it after the core collector. No layout change, so no format bump.
4. Tests with the fake clientset: events land under resources/core/v1/Event/<ns>/<name>.json, forbidden is skipped with a warning, other errors propagate. Then a real run against kind.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
EventsCollector lists core/v1 events across all namespaces (the CLI collects every namespace), so they land under resources/core/v1/Event/<ns>/<name>.json; no layout change, no format bump. apierrors.IsForbidden logs a warning (new log.Warn) and returns no events; other list errors fail the snapshot. Tests: fake clientset for collection, a Forbidden reactor for the skip with the warning captured, a generic error for propagation, and the manager tarball test now includes an event path and its manifest count. Real run on kind: 90 events indexed, including BackOff restarting failed container for the crash-looping api pod, FailedScheduling for the Pending pod and Failed pull for the ErrImagePull pod. Kubernetes keeps events for about an hour by default, so older problems can show no events.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
kubin create now collects core/v1 events from every namespace into the snapshot, next to namespaces and pods. A user without RBAC for events gets a warning on stderr and a snapshot without them; any other error still fails the run. Verified with fake-client tests and a run against kind, where the crash-looping, unschedulable and image-pull pods each have the Warning event that explains them.
<!-- SECTION:FINAL_SUMMARY:END -->
