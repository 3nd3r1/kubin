---
id: TASK-30
title: Demo the crashing-pod story and rewrite the README around it
status: To Do
assignee: []
created_date: '2026-11-17T14:08:00Z'
updated_date: '2026-11-17T14:23:00Z'
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-5: CLI progress on stderr is zap JSON lines. For the recording it should be short human-readable lines (or nothing but the link).
<!-- SECTION:NOTES:END -->
