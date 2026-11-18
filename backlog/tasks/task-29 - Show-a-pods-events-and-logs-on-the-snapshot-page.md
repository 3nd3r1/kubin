---
id: TASK-29
title: Show a pod's events and logs on the snapshot page
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T14:08:00Z'
updated_date: '2026-11-18T11:20:00Z'
labels:
  - ui
milestone: m-0
dependencies:
  - TASK-6
  - TASK-14
  - TASK-28
priority: high
ordinal: 29000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The demo (decision-3) ends when the viewer sees why the pod is crashing without access to the cluster: its events and the error in its logs. TASK-16 and TASK-17 build the full detail panel and log viewer in m-2. This is the part the demo needs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A pod's detail lists its events, newest first, with type, reason, message and count
- [x] #2 Each container's log is shown, with the previous log for containers that restarted
- [x] #3 A crash-looping pod opens with the crashing container's log of its last crashed run in view, scrolled to the end
- [x] #4 A pod with no events or logs says so instead of showing an empty panel
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. api.ts: Manifest gets logs and redactions; getLog fetches one log as text.
2. lib/events.ts: events for one pod (involvedObject kind Pod, name, uid when both have one), newest first by lastTimestamp, eventTime or firstTimestamp, count from count or series.count. Vitest.
3. lib/logs.ts: log files per container in spec order from manifest logs.files; the default container is the first one crashing (waiting CrashLoopBackOff, terminated non-zero, or restarted), its file is the current log, else previous. Vitest.
4. Pod page: Events section (table: type, reason, message, count, last seen relative to capture) and Logs section (container tabs, current/previous toggle when both exist, truncation note, error lines highlighted, scrolled to the end). A crashing pod scrolls the Logs section into view. Empty states for no events, no logs, and snapshots without logs.
5. Check in Chrome against kind: crash-looping, image-pull, healthy pod, and an old snapshot.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-11: for a CrashLoopBackOff container the crash output is usually in <container>.log (the last run), not <container>.previous.log. The kubelet keeps one dead container per name, so the run before is often gone and the CLI skips it. Pick the file to show from manifest logs.files rather than assuming previous exists.

AC3 reworded (was: previous log in view). TASK-11 found that the kubelet keeps one dead container per name, so a crash-looping container's crash is in <container>.log and its previous log is usually gone. The UI picks the file from the manifest's logs.files.

Logs come before Events on the pod page, so a crashing pod scrolls straight to the log. Default log (lib/logs.ts): a crash-looping or failed container opens its current log (the last run), or previous if only that was captured; a container that failed and is running again opens its previous run; otherwise the first container. Error-looking lines are highlighted. Events are matched by kind, name and uid, sorted by parsed time across event versions.
Verified: 11 new vitest tests (events and logs helpers), typecheck, eslint and prettier clean. In Chrome against a kind snapshot served by a server from this tree: the CrashLoopBackOff api pod opens scrolled to Logs with the panic line highlighted and scrolled to the end, Current/Previous run switch, events newest first with type, reason, message, count and last seen; the ImagePullBackOff and Pending pods say no logs were captured and list their events; a restarted web pod opens its previous run and says it wrote nothing, and says no events were recorded; a snapshot uploaded before TASK-14 says the log is not stored. No console errors.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
The pod page now shows Logs and Events. Logs: one tab per container with captured logs, a Current/Previous run switch when both exist, a note when the log was cut to the tail limit, error-looking lines highlighted, scrolled to the end. A crash-looping pod opens on the crashing container's last run and scrolls the Logs section into view, so the crash is on screen right away. Events: the pod's events newest first with type, reason, message, count and how long before capture. Every empty case says why (no events kept, container never started, old snapshot, empty log). AC3 was reworded to match how the kubelet keeps logs (TASK-11). Verified with unit tests and in Chrome on kind.
<!-- SECTION:FINAL_SUMMARY:END -->
