---
id: TASK-18
title: Snapshot overview that puts problems first
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-19T13:08:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies:
  - TASK-16
  - TASK-32
documentation:
  - docs/architecture.md
priority: medium
ordinal: 18000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Someone opening a link usually wants to know what is wrong. doc-1 asks for unhealthy things first in m-2, so the landing page of a snapshot shows that before anyone starts clicking. Today `/s/{id}` is the m-0 pod page (`ui/src/pages/pods.tsx`): pods with problems, then every pod. With the kind tables from TASK-15 and the problem flags from TASK-32 it can cover the whole snapshot.

## Where

- `ui/src/pages/overview.tsx` (new) on the index route `/s/{id}` in `ui/src/main.tsx`, replacing `PodsPage`. The full pod list lives at `/s/{id}/pods` since TASK-15, so drop `PodsPage` if nothing else uses it.
- Pod problems: `summarize` and `problems` in `ui/src/lib/pods.ts`. Other kinds: `problem(kind, object)` in `ui/src/lib/columns.ts` (TASK-32).
- Warning events: a pure function in `ui/src/lib/events.ts`, tested in `ui/src/lib/events.test.ts`.
- Counts: `counts` on the snapshot (`ui/src/lib/api.ts`), already loaded by `SnapshotLayout`. Node count comes from there, without fetching Nodes.
- Detail links: helpers in `ui/src/lib/snapshot.ts` (TASK-15, TASK-16).

## Out of scope

- Server endpoints for problems or summaries. Everything is computed in the browser from `listResources`.
- New problem rules beyond `problems` in `lib/pods.ts` and `problem` in `lib/columns.ts`.
- Charts.

## Traps

- "The last hour" is the hour before capture (`manifest.created_at`), not before now. Events without any time are left out.
- The m-0 demo (README, TASK-30) relies on the crashing pod being at the top of the link. Keep pod problems first; read `README.md` and check that what it says about the landing page stays true.
- Fetch Pods, Events and the kinds with problem rules as separate requests and render each section when it arrives, so one slow list doesn't blank the page.
- The namespace filter `?ns=` (TASK-15) narrows problems and events; cluster-scoped problems (Nodes, PVs) still show.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `/s/{id}` shows the cluster name, Kubernetes version, capture time, node count and a count per kind that links to the kind's table
- [x] #2 Problems come first: pods flagged by `problems()` in `ui/src/lib/pods.ts`, then resources flagged by `problem()` in `ui/src/lib/columns.ts`, errors before warnings
- [x] #3 Warning events from the hour before capture are listed newest first; the selection is tested in `ui/src/lib/events.test.ts`, including events older than the hour and events without a time
- [x] #4 Every listed item links to its detail page, and `?ns=` narrows problems and events
- [x] #5 With nothing wrong the page says so in one line instead of showing empty sections
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [x] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added ui/src/pages/overview.tsx on /s/{id}, replacing PodsPage (PodTable stays in pods.tsx for the kind page). Facts and per-kind count links, pod problems first, then problem() hits for workloads, nodes, jobs and volumes (errors first), then warnings from the hour before capture via recentWarnings in lib/events.ts (tested). Pods, Events and each checked kind load as separate requests; ?ns= narrows them. One-line all-clear when nothing is wrong. Lint, typecheck, test, build pass; nothing under server/, cli/ or snapshot-format.md changed.
<!-- SECTION:FINAL_SUMMARY:END -->
