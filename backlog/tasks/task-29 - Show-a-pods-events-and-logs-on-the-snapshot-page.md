---
id: TASK-29
title: Show a pod's events and logs on the snapshot page
status: To Do
assignee: []
created_date: '2026-11-17T14:08:00Z'
updated_date: '2026-11-18T08:33:00Z'
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
- [ ] #1 A pod's detail lists its events, newest first, with type, reason, message and count
- [ ] #2 Each container's log is shown, with the previous log for containers that restarted
- [ ] #3 A pod in CrashLoopBackOff opens with the crashing container's previous log in view
- [ ] #4 A pod with no events or logs says so instead of showing an empty panel
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-11: for a CrashLoopBackOff container the crash output is usually in <container>.log (the last run), not <container>.previous.log. The kubelet keeps one dead container per name, so the run before is often gone and the CLI skips it. Pick the file to show from manifest logs.files rather than assuming previous exists.
<!-- SECTION:NOTES:END -->
