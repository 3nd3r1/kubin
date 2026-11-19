---
id: TASK-33
title: Link related resources from the detail page
status: To Do
assignee: []
created_date: '2026-11-19T11:41:00Z'
updated_date: '2026-11-19T11:42:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies:
  - TASK-16
documentation:
  - docs/architecture.md
priority: medium
ordinal: 33000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Split from TASK-16. Debugging means jumping from a failing pod to its ReplicaSet, Deployment, Service and Node, and back. The detail pages from TASK-16 show one object; this task links it to the objects around it.

## Where

- `ui/src/lib/relations.ts` (new) with pure functions, tested in `ui/src/lib/relations.test.ts`.
- The detail page `ui/src/pages/resource.tsx` (TASK-16) and the pod page `ui/src/pages/pod.tsx`, whose `Owner` and `Node` fields become links.
- `ui/src/lib/k8s.ts`: `ObjectMeta.ownerReferences` lacks `apiVersion`, `uid` and `controller`; add them. Add `spec.selector` for Services.
- Detail URLs come from the kinds table and helpers in `ui/src/lib/kinds.ts` and `ui/src/lib/snapshot.ts` (TASK-15).

Relations to show:

- Owners: each owner reference links to the owner's detail.
- Owned: resources in the same namespace whose owner references contain this object's UID. Only these pairs: Deployment to ReplicaSet; ReplicaSet, StatefulSet, DaemonSet and Job to Pod; CronJob to Job. Fetch only the owned kind, in the object's namespace (`listResources` with `kind` and `namespace`).
- A Pod links to its Node (`spec.nodeName`) and to the Services in its namespace whose selector matches its labels.
- A Service lists the Pods its selector matches.
- A Node lists the Pods with `spec.nodeName` equal to its name.

## Out of scope

- Ingress to Service, PVC to PV, EndpointSlice and ConfigMap links. Leave them for later.
- Server, API, CLI. Everything is computed in the browser.
- A graph view.

## Traps

- Match owned objects by owner UID, not name: a Deployment recreated under the same name must not claim the old ReplicaSets.
- A Service without a selector (ExternalName, manual endpoints) selects nothing. An empty selector must not match every pod.
- The owner or node may not be in the snapshot (a namespace-scoped snapshot has no Nodes, an owner kind may not be collected). Show plain text, never a link to a 404.
- Pods on a Node can be in any namespace: list Pods without a namespace filter for that one.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Owner references are links to the owner's detail, or plain text when the owner isn't in the snapshot
- [ ] #2 Deployments, ReplicaSets, StatefulSets, DaemonSets, Jobs and CronJobs list the resources they own, matched by owner UID, as links
- [ ] #3 A Pod links to its Node and to the Services whose selector matches its labels; a Service lists the Pods it selects; a Node lists the Pods scheduled on it
- [ ] #4 `ui/src/lib/relations.test.ts` covers UID matching (same name, other UID: no match), selector matching (no selector matches nothing) and owners missing from the snapshot
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [ ] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->
