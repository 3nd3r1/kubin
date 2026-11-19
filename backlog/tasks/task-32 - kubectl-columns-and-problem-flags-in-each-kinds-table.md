---
id: TASK-32
title: kubectl columns and problem flags in each kind's table
status: To Do
assignee: []
created_date: '2026-11-19T11:41:00Z'
updated_date: '2026-11-19T11:42:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies:
  - TASK-15
documentation:
  - docs/architecture.md
priority: high
ordinal: 32000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Split from TASK-15. A Lens or kubectl table tells you a resource's state at a glance: READY 0/3, STATUS NotReady. TASK-15 gives every kind a table with Name, Namespace and Age; this task adds the columns `kubectl get` shows for each kind and flags the rows that are unhealthy, which doc-1 asks to see first in m-2.

## Where

- `ui/src/lib/columns.ts` (new): per kind, a list of columns, each a header and a pure function from the object (and capture time) to the cell text. Plus an exported `problem(kind, object)` that returns `"error" | "warning" | null` like `Problem` in `ui/src/lib/pods.ts`. TASK-18 reuses it.
- `ui/src/lib/columns.test.ts` (new).
- `ui/src/lib/k8s.ts`: add the fields the columns read. Every field optional, as the file's header says.
- The kind table in `ui/src/pages/kind.tsx` from TASK-15. Pods already have their columns in `ui/src/lib/pods.ts`; reuse them, don't duplicate.

Columns, after Name and Namespace and before Age as TASK-15 shows them:

| Kind | Columns |
|---|---|
| Deployment | READY (ready/replicas), UP-TO-DATE, AVAILABLE |
| ReplicaSet | DESIRED, CURRENT, READY |
| StatefulSet | READY |
| DaemonSet | DESIRED, CURRENT, READY, UP-TO-DATE, AVAILABLE, NODE SELECTOR |
| Job | STATUS (Complete, Failed, Suspended, Running), COMPLETIONS (succeeded/completions), DURATION |
| CronJob | SCHEDULE, TIMEZONE, SUSPEND, ACTIVE, LAST SCHEDULE |
| Service | TYPE, CLUSTER-IP, EXTERNAL-IP, PORT(S) |
| Ingress | CLASS, HOSTS, ADDRESS, PORTS |
| EndpointSlice | ADDRESSTYPE, PORTS, ENDPOINTS |
| ConfigMap | DATA |
| PersistentVolumeClaim | STATUS, VOLUME, CAPACITY, ACCESS MODES, STORAGECLASS |
| PersistentVolume | CAPACITY, ACCESS MODES, RECLAIM POLICY, STATUS, CLAIM, STORAGECLASS |
| Node | STATUS, ROLES, VERSION |
| Namespace | STATUS |
| Event | LAST SEEN, TYPE, REASON, OBJECT, MESSAGE, newest first |

Problems: workloads (Deployment, StatefulSet, DaemonSet, ReplicaSet) with fewer ready than desired, Nodes not Ready, failed Jobs, PVCs not Bound. Use "error" for a Node not Ready, a failed Job, and a workload with 0 ready of more than 0 desired; "warning" for the rest.

## Out of scope

- Sorting, column pickers, filtering by text.
- Server, API, CLI. Everything is computed in the browser from the objects `listResources` returns.
- The detail page (TASK-16).

## Traps

- Ages, LAST SEEN, LAST SCHEDULE and DURATION are relative to capture time (`manifest.created_at`), never `Date.now()`. `age()` in `ui/src/lib/pods.ts` already does this.
- Missing fields show `<none>` or `0` as kubectl does, never `undefined` or `NaN`. Status is often missing on objects that never reconciled.
- Event times: core Events and events.k8s.io Events set different fields. Export and reuse `lastSeen` from `ui/src/lib/events.ts`.
- ConfigMap DATA counts the keys of `data` and `binaryData` together.
- Node ROLES come from `node-role.kubernetes.io/<role>` labels, `<none>` if there are none. STATUS appends `,SchedulingDisabled` when `spec.unschedulable` is true.
- ACCESS MODES use kubectl's short forms: RWO, ROX, RWX, RWOP.
- A ReplicaSet scaled to 0 (old revision of a Deployment) is not a problem.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Each kind in the table in the description shows its kubectl columns, built by functions in `ui/src/lib/columns.ts`
- [ ] #2 `ui/src/lib/columns.test.ts` covers every kind, including an object without `status` that renders `<none>` or `0`, never `undefined` or `NaN`
- [ ] #3 Ages, LAST SEEN, LAST SCHEDULE and DURATION are computed against capture time, with a test
- [ ] #4 An exported `problem(kind, object)` flags workloads with fewer ready than desired, Nodes not Ready, failed Jobs and PVCs not Bound, is tested, and flagged rows are tinted like problem pods
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [ ] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->
