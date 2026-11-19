---
id: TASK-17
title: Log viewer that handles large logs and links to a line
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-19T11:42:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies:
  - TASK-14
documentation:
  - docs/architecture.md
  - docs/snapshot-format.md
priority: high
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Reading logs is the main thing people do with a shared snapshot. m-0 built the viewer (`ui/src/components/pod/logs.tsx`): a tab per container, a current/previous switch and error-line highlighting. It renders every line as a DOM node, so big logs are slow, and nobody can link to the line that matters. Search is TASK-34.

## Where

- `ui/src/components/pod/logs.tsx` (`PodLogs`, `LogView`).
- `ui/src/lib/logs.ts` and `ui/src/lib/logs.test.ts`.
- URL state on the pod detail URL through `useSearchParams` from react-router: `?container={name}&previous=1&line={n}`. Add parse and format helpers to `ui/src/lib/logs.ts`.
- Windowing: render only the lines in view plus a margin. Use fixed-height lines without wrapping (horizontal scroll), and a pure `visibleRange(scrollTop, viewportHeight, lineHeight, lineCount, overscan)` in `ui/src/lib/logs.ts`. No new npm dependency.

## Out of scope

- Search (TASK-34).
- The server's log endpoint and its `q` filter (`server/internal/api/logs.go`), the log collector, and the snapshot format.
- Download, follow mode, timestamps.

## Traps

- The CLI keeps the last `--log-tail` lines and at most `manifest.logs.max_bytes` per file, but users can raise both. Don't assume a small log.
- Line numbers count lines in the captured file, not in the container's whole log. When the file is `truncated`, the existing notice already says older lines are missing; keep it.
- `PodLogs` picks the log to open with `defaultLog` and scrolls a crash-looping pod's log section into view. A `container`, `previous` or `line` in the URL wins over both.
- Keep `useFetch` deps on container and previous only; changing `line` must not refetch the log.
- One very long line must not change the row height. That is why lines don't wrap.
- `-mx-4`/`px-4` on error lines in the current `<pre>` assume padding on the container; recheck once rows are positioned.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The log view renders only the lines in view plus a margin; `visibleRange` in `ui/src/lib/logs.ts` is tested for the top, middle and end of a 50,000-line log and for an empty log
- [ ] #2 Each line shows its number; clicking it puts `container`, `previous` and `line` in the pod URL, and opening such a URL selects that container and run, scrolls to the line and highlights it; the URL helpers are tested
- [ ] #3 Without `line` in the URL the viewer opens the log `defaultLog` picks, scrolled to the end, as today
- [ ] #4 Error-line highlighting and the truncated notice still work
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [ ] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->
