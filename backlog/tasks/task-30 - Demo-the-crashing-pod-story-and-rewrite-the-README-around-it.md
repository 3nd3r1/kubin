---
id: TASK-30
title: Demo the crashing-pod story and rewrite the README around it
status: Done
assignee:
  - '@human'
created_date: '2026-11-17T14:08:00Z'
updated_date: '2026-11-19T14:57:00Z'
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
- [x] #1 A demo directory has a manifest whose pod crashes with a clear error in its logs, and one command that starts kind and kubin, applies it and runs kubin create
- [x] #2 A recording (GIF or short video) of the story, from kubin create to seeing the error in the logs, is in the repo or linked from it
- [x] #3 The README opens with one sentence, the recording, and the three commands to reproduce it, above everything else
- [x] #4 The README says what works today and what does not, with no claims beyond what the demo shows
- [x] #5 A script for the narrated video is in demo/
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add a crashing pod manifest and one script that creates kind, starts the server, applies the manifest and runs the CLI.
2. Capture the browser flow from upload through the error in the pod logs.
3. Rewrite the README so the demo and three reproduction commands come first, then state current limits.

4. User wants a narrated face-bubble video as the top of the README (2026-11-18), with the GIF kept below it. Claude writes the video script; the user records.

5. kubin create prints plain lines instead of zap JSON so the terminal reads well on camera.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-5: CLI progress on stderr is zap JSON lines. For the recording it should be short human-readable lines (or nothing but the link).

Added demo/crashing-pod.yaml and demo/run.sh. The runner builds the image and CLI, starts kind and Kubin, applies the pod, waits for a restart, and uploads a snapshot. Rewrote README around the demo and current limits. `bash -n demo/run.sh` and `git diff --check` pass.

End-to-end run is pending: Docker access is denied in this environment. Recording is pending: no browser or screen-recording tool is available here. Acceptance criteria remain unchecked.

Claude ran demo/run.sh end to end on 2026-11-18 and fixed four bugs: the pod command was parsed as a YAML map, the pod apply raced the namespace's default service account, the readiness curl gave up on a connection reset, and a failed docker run left a Created container behind. Added KUBIN_PORT. The snapshot carries the crash line in both the current and previous logs. Recording still pending.

Claude, 2026-11-18: kubin create now prints plain stderr lines ("Collecting the cluster...", "Uploading N resources from C to S...") instead of zap JSON; zap is gone from the CLI. run.sh waits for the node (kind --wait 2m), so the snapshot no longer carries a FailedScheduling warning. Recorded demo/kubin-demo.gif (180 KB, about 11s): a terminal card with the link, the snapshot page, the pod's logs and the previous run. To re-record: run demo/run.sh, take full-size browser screenshots of the same four views, crop to the content column and join them with ffmpeg (palettegen, 10 fps, 880 px wide). Restored the Kubin title and logo that a0d85af dropped. Wrote demo/video-script.md. AC #5 waits on the user's narrated video.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Demo (demo/run.sh, crashing-pod.yaml), GIF (demo/kubin-demo.gif), README rewrite and the video script (demo/video-script.md) are done. The narrated face-bubble video moved to its own human task.
<!-- SECTION:FINAL_SUMMARY:END -->
