---
id: TASK-6
title: Snapshot page with broken pods first and pod details
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T14:52:00Z'
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
- [x] #1 `/s/{id}` shows cluster name, capture time and every pod grouped by namespace
- [x] #2 Pods that are not Ready, are in CrashLoopBackOff or Error, or have restarts are listed first and marked
- [x] #3 Each pod shows phase, restarts and age at capture time
- [x] #4 Selecting a pod shows its containers, their state and last termination reason, and the full object as YAML
- [x] #5 An unknown ID shows a not-found page
- [x] #6 The UI calls the API on its own origin
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Pod status logic in ui/src/lib/pods.ts, following kubectl get pods: status reason from init and app container states, ready count, restarts, age at capture time, and a problem level (error for CrashLoopBackOff, Error, OOMKilled, image pull and config errors, Failed; warning for Pending, not Ready, restarts). Unit tests with vitest.
2. Routes: /s/:id loads the snapshot once and keeps a header with cluster, Kubernetes version and capture time. Index shows a Problems table first, then every pod in a table per namespace, problem rows marked. /s/:id/pods/:ns/:name shows containers (image, state, ready, restarts, last termination reason and exit code) and the object as YAML, so a pod can be linked directly.
3. Check against real CrashLoopBackOff, ImagePullBackOff, Pending and healthy pods on kind in Chrome.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Status logic in ui/src/lib/pods.ts follows the kubectl get pods STATUS column (init containers incl. sidecars, waiting and terminated reasons, Terminating) and has vitest tests. Problem levels: error for crash, image, config and exit-code failures; warning for Pending, not Ready or restarts. Pod detail is its own route, /s/:id/pods/:ns/:name, so a link can point straight at a broken pod. useFetch keys results by their inputs (React lint forbids resetting state in the effect). Checked in Chrome against a kind snapshot with a crash-looping Deployment pod, an ErrImagePull pod, an unschedulable Pending pod and healthy pods: statuses, ready counts and restarts match kubectl, problems are listed first with errors before warnings, the detail shows the exit code, last termination and the image pull error, unknown pods and snapshots show not-found pages, no console errors. The test pods stay in the kind cluster (namespace shop) for TASK-28, TASK-11 and TASK-29.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The snapshot page now opens on what is broken. /s/:id shows the cluster, Kubernetes version and capture time, then a table of pods that need attention (errors first, then warnings), then every pod in a table per namespace with status, ready, restarts and age at capture time, problem rows marked. Selecting a pod opens /s/:id/pods/:ns/:name with namespace, node, owner, each container with image, state, ready, restarts and last termination reason and exit code, and the full object as YAML. Unknown snapshots and pods get not-found pages. All calls go to the API on the same origin. Pod status logic has 16 unit tests (vitest, npm test).
<!-- SECTION:FINAL_SUMMARY:END -->
