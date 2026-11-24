---
id: TASK-46
title: Show when a snapshot expires and an expired page in the UI
status: To Do
assignee: []
created_date: '2026-11-24T09:36:00Z'
labels:
  - ui
  - tier:2
milestone: m-3
dependencies:
  - TASK-21
priority: medium
type: feature
ordinal: 7600
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
After TASK-21 an expired snapshot answers 410, and the UI shows that as a generic error because `ui/src/lib/api.ts` maps only 404 to `NotFoundError`. People who open an old link should see that it expired, and the snapshot header should say when a live one will.

## Where

- `ui/src/lib/api.ts`: an `ExpiredError` for 410, next to `NotFoundError`. The snapshot type gets an optional `expires_at: string` (TASK-21 adds it to `GET /api/v1/snapshots/{id}`).
- `ui/src/lib/use-fetch.ts`: an `expired` state next to `missing`.
- `ui/src/pages/snapshot.tsx`: renders `NotFoundPage` for `missing` today. Render a page saying the snapshot expired for `expired`. The header there shows the capture time in a `<time>` element; put the expiry next to it, like "expires in 3 days", with the exact time in `title`.
- Tests next to the code in `ui/src/lib/*.test.ts`, with vitest fake timers for "now".

## Out of scope

Server and CLI changes, a delete button (the token stays with the CLI), restyling the header.

## Traps

- Snapshots stored before TASK-21 have no `expires_at`. Show nothing for them.
- An expiry already in the past while the page is open: show "expired", don't crash.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A 410 from any snapshot API call shows a page saying the snapshot expired, not a generic error
- [ ] #2 The snapshot header shows when the snapshot expires, relative, with the exact time in a title, and nothing when `expires_at` is missing
- [ ] #3 Unit tests cover the 410 mapping and the expiry text, including a missing and a past expiry
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd ui && make lint && make typecheck && make test passes
<!-- DOD:END -->
