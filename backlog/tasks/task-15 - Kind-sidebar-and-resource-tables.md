---
id: TASK-15
title: Kind sidebar, kind routes and namespace filter
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-19T12:11:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies:
  - TASK-6
  - TASK-10
documentation:
  - docs/architecture.md
priority: high
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

m-2 (doc-1, milestone 3) makes browsing a snapshot feel like Lens, starting with kinds in a sidebar. Since TASK-10 a snapshot holds 16 kinds, but the UI from TASK-6 only shows pods. This task builds the frame the other m-2 tasks hang off: the sidebar, one route per kind and the namespace filter. The kubectl columns per kind are TASK-32. Background: `docs/architecture.md` (UI section).

## Where

- `ui/src/lib/kinds.ts` (new): one table of the kinds in `Kinds` in `cli/pkg/collector/resources.go`, with group as the API returns it (`core` for the core group), kind, URL segment, sidebar section and whether it is namespaced. URL segments are the lowercase plurals kubectl uses: `pods`, `deployments`, `replicasets`, `statefulsets`, `daemonsets`, `jobs`, `cronjobs`, `services`, `ingresses`, `endpointslices`, `configmaps`, `persistentvolumeclaims`, `persistentvolumes`, `nodes`, `namespaces`, `events`. Sections: Workloads (Pods, Deployments, DaemonSets, StatefulSets, ReplicaSets, Jobs, CronJobs), Network (Services, Ingresses, EndpointSlices), Config (ConfigMaps), Storage (PersistentVolumeClaims, PersistentVolumes), Cluster (Nodes, Namespaces, Events).
- URL helpers next to `podPath` in `ui/src/lib/snapshot.ts`.
- `ui/src/main.tsx`: routes under `/s/:id`.
- `ui/src/pages/snapshot.tsx` (`SnapshotLayout`): sidebar plus content.
- `ui/src/pages/kind.tsx` (new): the list page for one kind, a table with Name, Namespace (namespaced kinds only) and Age. Follow `PodTable` in `ui/src/pages/pods.tsx`, `useFetch`, `listResources` in `ui/src/lib/api.ts` and `Loading`/`LoadError`/`Empty` in `ui/src/components/common/states.tsx`. Pods keep the columns they have today (`summarize` in `ui/src/lib/pods.ts`).
- Tests as `ui/src/lib/*.test.ts` (vitest, no DOM), like `ui/src/lib/pods.test.ts`.
- `docs/architecture.md`, UI section: add the URL scheme below in two or three lines.

URL scheme. TASK-16, TASK-17 and TASK-18 rely on it:

- `/s/{id}` stays the current problems-first pod page until TASK-18 replaces it with the overview.
- `/s/{id}/{resource}` lists one kind, `{resource}` being the URL segment above.
- `/s/{id}/{resource}/{namespace}/{name}` is a resource's detail, with `_cluster` as the namespace of cluster-scoped kinds, as in the API. Pod links from m-0 (`/s/{id}/pods/{ns}/{name}`) must keep working. Until TASK-16 only pods have a detail page, so rows of other kinds are not links yet.
- `?ns={namespace}` is the namespace filter. It applies to every namespaced kind, sidebar links keep it, and no `ns` means all namespaces.

## Out of scope

- kubectl columns per kind (TASK-32), detail pages for kinds other than Pod (TASK-16), the overview (TASK-18).
- Server, API, CLI and the snapshot format. The API already filters by `kind=` and `namespace=` (`server/internal/api/read.go`); this task needs no server change.
- Sidebar counts that follow the namespace filter. Counts are snapshot totals.

## Traps

- Use `counts` from `GET /api/v1/snapshots/{id}` (`Snapshot.counts` in `ui/src/lib/api.ts`), not `manifest.resources`. Their group is `core` for the core group, not an empty string.
- A kind in `counts` that isn't in `kinds.ts` (a snapshot from a newer CLI) goes under an "Other" section with the generic table rather than disappearing. Kinds the snapshot doesn't have are hidden.
- The namespace choices come from the snapshot's Namespace resources. A snapshot taken with `kubin create --namespace` only has those.
- An unknown `{resource}` segment must render `NotFoundPage`, not crash.
- `ui/src/components/common/layout.tsx` caps the page at `max-w-6xl`. Snapshot pages need the full width for sidebar plus table; keep the cap on the home page.
- `useFetch` reloads when its deps change: put the namespace in the deps.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The sidebar lists every kind in the snapshot under Workloads, Network, Config, Storage and Cluster with its count from `snapshot.counts`, and kinds missing from `ui/src/lib/kinds.ts` under Other; the grouping is tested in `ui/src/lib/kinds.test.ts`
- [x] #2 `/s/{id}/{resource}` shows a table of that kind with Name, Namespace for namespaced kinds, and Age; Pods keep Status, Ready and Restarts; an unknown `{resource}` shows the not-found page
- [x] #3 `?ns=` filters every namespaced kind's table, sidebar links keep it, and reloading a URL restores the same view; the URL helpers are tested
- [x] #4 `/s/{id}/pods/{ns}/{name}` still opens the pod page and `/s/{id}` still shows the problems-first pod page
- [x] #5 `docs/architecture.md` describes the URL scheme
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [x] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added ui/src/lib/kinds.ts (kind table, sidebar grouping, segment resolution, Other section for unknown kinds), URL helpers in snapshot.ts, KindPage (generic Name/Namespace/Age table; pods reuse PodTable), sidebar and namespace select in SnapshotLayout, full-width snapshot pages, and the URL scheme in docs/architecture.md. Routes: :resource lists a kind (unknown segment gives NotFoundPage); pods/:ns/:name and the index page are unchanged. Verified with make lint, typecheck, test, build and new vitest tests; not exercised in a browser. Nothing under server/, cli/ or snapshot-format.md changed.
<!-- SECTION:FINAL_SUMMARY:END -->
