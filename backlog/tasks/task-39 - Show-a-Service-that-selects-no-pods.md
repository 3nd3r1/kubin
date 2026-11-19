---
id: TASK-39
title: Show a Service that selects no pods
status: To Do
assignee: []
created_date: '2026-11-19T14:45:00Z'
updated_date: '2026-11-19T14:50:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies: []
references:
  - ui/src/lib/relations.ts
  - ui/src/components/resource/related.tsx
  - ui/src/pages/overview.tsx
documentation:
  - docs/architecture.md
priority: high
ordinal: 39000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

In the new m-0 demo (user, 2026-11-19) the `web` Service has a selector typo, so it selects none of the two running `web` pods and has no endpoints. Today that is invisible: the Service detail hides its Pods section when nothing matches (`Related` in `ui/src/components/resource/related.tsx` returns null on an empty list), and the overview (TASK-18) has no rule for Services. A Service that selects nothing is one of the most common "everything is running but nothing works" problems, and the demo shows it on camera.

## Where

- `ui/src/lib/relations.ts`: a pure function that says whether a Service has a non-empty selector and no pod in its namespace matches it. Reuse `selectorMatches` and `podsFor`. Tests in `ui/src/lib/relations.test.ts`.
- `ui/src/components/resource/related.tsx`: for a Service with a selector, the Pods section stays visible when nothing matches and says so with the selector, for example `No pods match app=wbe`.
- `ui/src/pages/overview.tsx`: list such Services among the problems as warnings, linking to the Service. `problem(kind, object)` in `ui/src/lib/columns.ts` sees one object at a time, so compute this in the overview from the Services and Pods lists instead.

## Out of scope

- EndpointSlice readiness (pods matched but not Ready). Matching by selector is enough here.
- Flagging Services in the kind table (`columns.ts`).
- Server, API, CLI, `docs/snapshot-format.md`.

## Traps

- A Service without a selector (the `kubernetes` Service in `default`, ExternalName Services, manually managed endpoints) selects nothing by design. Never flag it; `selectorMatches` already treats an empty selector as matching nothing, so check for the selector first.
- If the snapshot has no Pods (not in `snapshot.counts`, for example when RBAC denied listing them), don't flag any Service and don't say "No pods match".
- Pods match only in the Service's own namespace.
- The overview loads each section as its own request and renders it when it arrives (TASK-18); keep that. The one-line all-clear must stay false when a Service is flagged, and `?ns=` must narrow the new list.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The detail of a Service with a selector that matches no pod in its namespace shows its Pods section with a line naming the selector, for example `No pods match app=wbe`
- [ ] #2 The overview lists such Services as warnings with a link to each, and `?ns=` narrows the list
- [ ] #3 Services without a selector, Services that match at least one pod, and every Service in a snapshot without Pods are neither flagged nor described as matching no pods
- [ ] #4 `ui/src/lib/relations.test.ts` covers a selector matching nothing, an empty or missing selector, a matching pod in another namespace, and a matching pod in the same namespace
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [ ] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->
