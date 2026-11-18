---
id: TASK-30
title: Demo the crashing-pod story and rewrite the README around it
status: In Progress
assignee:
  - '@codex'
created_date: '2026-11-17T14:08:00Z'
updated_date: '2026-11-18T12:43:00Z'
labels:
  - docs
milestone: m-0
dependencies:
  - TASK-5
  - TASK-12
  - TASK-25
  - TASK-29
priority: high
ordinal: 30000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The README describes kubin but does not show why anyone would use it (user, 2026-11-17). decision-3 makes m-0 a demo: a crashing pod debugged from a link. Someone landing on the repo should understand the value from the first screen, before reading any prose. The demo has to be reproducible so it can be re-recorded as the UI changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A demo directory has a manifest whose pod crashes with a clear error in its logs, and one command that starts kind and kubin, applies it and runs kubin create
- [ ] #2 A recording (GIF or short video) of the story, from kubin create to seeing the error in the logs, is in the repo or linked from it
- [ ] #3 The README opens with one sentence, the recording, and the three commands to reproduce it, above everything else
- [ ] #4 The README says what works today and what does not, with no claims beyond what the demo shows
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add a crashing pod manifest and one script that creates kind, starts the server, applies the manifest and runs the CLI.
2. Capture the browser flow from upload through the error in the pod logs.
3. Rewrite the README so the demo and three reproduction commands come first, then state current limits.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-5: CLI progress on stderr is zap JSON lines. For the recording it should be short human-readable lines (or nothing but the link).

Added demo/crashing-pod.yaml and demo/run.sh. The runner builds the image and CLI, starts kind and Kubin, applies the pod, waits for a restart, and uploads a snapshot. Rewrote README around the demo and current limits. `bash -n demo/run.sh` and `git diff --check` pass.

End-to-end run is pending: Docker access is denied in this environment. Recording is pending: no browser or screen-recording tool is available here. Acceptance criteria remain unchecked.
<!-- SECTION:NOTES:END -->
