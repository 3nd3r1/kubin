---
id: decision-3
title: 'm-0 is a demo: a crashing pod debugged from a link'
date: '2026-11-17T14:07:00Z'
status: accepted
---
## Context

On 2026-11-17 the user said the README doesn't make the value clear and asked to aim at a milestone that can be shown in a short demo video. m-0 as planned ended with a link that lists namespaces and pods. That proves the plumbing but doesn't show why anyone would use kubin.

## Decision

The user chose this on 2026-11-17. m-0 is renamed "Demo: crashing pod to a link" and ends with this story, short enough for a ~30 second recording:

1. A kind cluster runs a pod in CrashLoopBackOff.
2. `kubin create` against a kubin started with `docker run` prints a link.
3. The link opens on the broken pod first, with its events and its container logs showing the actual error.

Secret values stay out of the snapshot, since the demo invites sharing links. The README is rewritten around this story, with the recording, once it works.

## Consequences

Events (TASK-28), pod logs (TASK-11, TASK-14) and secret stripping (TASK-12) move into m-0. A pod events and logs view (TASK-29) and the demo and README work (TASK-30) are added. CI (TASK-8) and the GitHub issue setup (TASK-9) move to m-3. The full Lens layout and the complete overview (TASK-15 to TASK-18) stay in m-2. doc-1 still describes the old m-0 and needs the user's update.
