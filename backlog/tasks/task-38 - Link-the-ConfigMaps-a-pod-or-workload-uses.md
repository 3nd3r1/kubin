---
id: TASK-38
title: Link the ConfigMaps a pod or workload uses
status: Done
assignee:
  - '@claude'
created_date: '2026-11-19T14:45:00Z'
updated_date: '2026-11-24T08:06:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies: []
references:
  - ui/src/lib/relations.ts
  - ui/src/components/resource/related.tsx
documentation:
  - docs/architecture.md
priority: high
ordinal: 38000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

In the new m-0 demo (user, 2026-11-19) the `api` Deployment crash-loops because its ConfigMap `api-config` says `DB_HOST=db` while the database Service is called `postgres`. On camera the viewer clicks from the api pod to its ReplicaSet, to the Deployment, to the ConfigMap. TASK-33 linked owners, owned objects, Services and Nodes, and left ConfigMap links out, so the last click doesn't exist yet. decision-5 drops env var values but keeps names and `valueFrom` references, and ConfigMap data is collected as is, so the snapshot has what this needs.

## Where

- `ui/src/lib/relations.ts`: a pure function that takes a pod spec and returns the names of the ConfigMaps it references, each once, sorted. Sources, in `containers` and `initContainers`: `envFrom[].configMapRef.name`, `env[].valueFrom.configMapKeyRef.name`; and in `volumes`: `configMap.name` and `projected.sources[].configMap.name`.
- A helper (same file) that finds the pod spec of an object: `spec` for a Pod; `spec.template.spec` for Deployment, ReplicaSet, StatefulSet, DaemonSet and Job; `spec.jobTemplate.spec.template.spec` for CronJob.
- `ui/src/lib/relations.test.ts`: tests for both.
- `ui/src/components/resource/related.tsx`: a "ConfigMaps" section in `RelatedResources`, which both `ui/src/pages/pod.tsx` and `ui/src/pages/resource.tsx` render. Load ConfigMaps with `listResources` (kind `ConfigMap`, the object's namespace) only when the object references any and the snapshot's `counts` has ConfigMaps. Reuse `Links`/`itemPath`; widen the local `Obj` type as needed.
- `ui/src/lib/k8s.ts`: add the pod spec fields you read, if typed there.

## Out of scope

- Secret, ServiceAccount and PVC links. Same idea, not needed for the demo.
- The reverse direction (a ConfigMap listing who uses it).
- Server, API, CLI, `docs/snapshot-format.md`.

## Traps

- A referenced ConfigMap may be missing from the snapshot (not collected, filtered by namespace, or really missing in the cluster). Show the name as plain text with "not in snapshot", never a link to a 404. Missing is itself a debugging clue.
- `RelatedResources` on `resource.tsx` gets `kind` from the route (`{ ...obj, kind }`); on `pod.tsx` the object is the pod. Check both pages show the section.
- An object that references no ConfigMap shows no section and makes no request.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The detail pages of a Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, Job and CronJob show a ConfigMaps section that lists each ConfigMap the pod spec references once, as a link to the ConfigMap's detail
- [x] #2 A referenced ConfigMap that is not in the snapshot shows as plain text marked "not in snapshot", never as a link
- [x] #3 An object that references no ConfigMap shows no ConfigMaps section and makes no ConfigMap request
- [x] #4 `ui/src/lib/relations.test.ts` covers envFrom, env valueFrom configMapKeyRef, configMap volumes, projected volumes, initContainers, a name referenced twice, the pod spec lookup for each kind, and a spec with no references
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [x] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added podSpecOf and configMapNames to relations.ts with tests, a ConfigMaps section in RelatedResources (links found ConfigMaps, plain text '(not in snapshot)' for missing ones, nothing and no request when no references or no ConfigMap in counts), and the pod spec fields in k8s.ts. ui lint, typecheck, test and build pass. The component itself has no test and was not checked in a browser.
<!-- SECTION:FINAL_SUMMARY:END -->
