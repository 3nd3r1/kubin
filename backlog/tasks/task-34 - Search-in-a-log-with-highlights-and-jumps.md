---
id: TASK-34
title: Search in a log with highlights and jumps
status: Done
assignee:
  - '@claude'
created_date: '2026-11-19T11:42:00Z'
updated_date: '2026-11-24T08:04:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies:
  - TASK-17
documentation:
  - docs/architecture.md
priority: high
ordinal: 34000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Split from TASK-17. Finding the one error in thousands of log lines is the common case, and the browser's own find can't see lines the virtualized view (TASK-17) doesn't render.

## Where

- `ui/src/components/pod/logs.tsx`: a search box above the log, next/previous buttons, a "3 of 17" counter.
- `ui/src/lib/logs.ts`: a pure `findMatches(lines, query)` returning line index, start and end per match, case-insensitive, with the query taken literally. Tests in `ui/src/lib/logs.test.ts`.
- Scroll to a match with the scroll-to-line code TASK-17 added for `?line=`.
- The query goes in the pod URL as `q=`, next to TASK-17's `container`, `previous` and `line`.

Search runs in the browser over the log text the view already fetched. The server's `?q=` filter on the log endpoint (TASK-14, `server/internal/api/logs.go`) returns only matching lines and is not used here.

## Out of scope

- Regex search, filtering the log down to matching lines, searching across containers or pods.
- Server, API, CLI.

## Traps

- Compute matches once per query and log, over all lines. Render highlights only for the lines in view.
- Escape the query before building a RegExp, or don't use RegExp at all: `.`, `*`, `(` and `[` are common in logs.
- An empty query means no matches and no highlights, not a match at every position.
- Don't intercept Ctrl+F. Enter and Shift+Enter move between matches only while the search box has focus.
- Error-line highlighting from m-0 stays and must not hide the match highlight.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A search box above the log highlights every match, case-insensitive, and shows which match is current out of how many
- [x] #2 Enter, Shift+Enter and next/previous buttons move between matches and scroll the view to the match's line
- [x] #3 `findMatches` is tested for several matches on one line, no matches, an empty query, and regex characters taken literally
- [x] #4 The query is kept in the pod URL as `q`, so a shared link opens with the same search
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [x] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a search bar to LogView (ui/src/components/pod/logs.tsx): case-insensitive highlights over all lines (computed once per query/log, rendered only for visible lines), 'N of M' counter, Prev/Next buttons and Enter/Shift+Enter in the box, centring the current match. Query lives in the pod URL as q. findMatches in ui/src/lib/logs.ts is literal (no RegExp), empty query gives no matches; tested in logs.test.ts. Marks sit on top of the error-line and linked-line backgrounds. make lint/typecheck/test/build pass; nothing under server/, cli/ or docs touched. Not exercised in a browser; UI behaviour is verified by code and typecheck only.
<!-- SECTION:FINAL_SUMMARY:END -->
