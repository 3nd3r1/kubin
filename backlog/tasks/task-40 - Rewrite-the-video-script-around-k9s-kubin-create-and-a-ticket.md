---
id: TASK-40
title: 'Rewrite the video script: is this familiar, kubin create, a ticket link'
status: To Do
assignee: []
created_date: '2026-11-19T14:46:00Z'
updated_date: '2026-11-19T14:50:00Z'
labels:
  - docs
  - tier:2
milestone: m-0
dependencies:
  - TASK-35
  - TASK-36
  - TASK-37
references:
  - demo/video-script.md
  - demo/checkout.yaml
  - demo/setup.sh
priority: high
ordinal: 40000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

The user changed the demo story again on 2026-11-19: nobody debugs on camera. `demo/video-script.md` still tells the crashing `payment-api` story. The user records the narrated video from this script (TASK-30, owned by the user). The video asks "is this familiar?": k9s or kubectl shows a broken cluster, you need a ticket with logs, and you'd paste a huge log dump. Instead one `kubin create` captures everything non-secret and prints a link in seconds, the link goes into a Linear ticket, then a quick look at the UI. No root-cause walkthrough. Background: decision-8, decision-5 (env var values are dropped).

## Where

`demo/video-script.md` only. Keep its shape: a short intro (length, face bubble, captions), a "Before recording" paragraph, numbered sections with time ranges, each with **Screen** and **Say**, and a "Don't claim" section at the end. Keep the plain voice of the current script.

Beats, in order:

1. "Is this familiar?" k9s (or kubectl) on namespace `checkout` shows a broken cluster: `api` in CrashLoopBackOff, `postgres-0` Pending, `worker` failed.
2. You need a ticket with logs, so you'd copy `kubectl describe` and a huge log dump into it.
3. Instead: `kubin create` in a terminal. It captures everything non-secret and prints a link in seconds. Keep the line that secrets and env values stay out of the snapshot.
4. Paste the link into a Linear ticket.
5. A quick look at the link: the overview with problems first, one pod's logs (the api pod), the kind sidebar. No root cause, no click chain.
6. Close on self-hosting: one container.

"Before recording" names the real commands: `./demo/setup.sh`, the `export` line it prints, `k9s -n checkout`, the browser zoom and history advice already there.

## Out of scope

- TASK-30 and its criteria: the user's task.
- README and GIF (TASK-41), demo scripts and manifests.

## Traps

- Every view the script shows must exist in the UI today (overview, pod logs, kind sidebar). Check against the running demo (`./demo/setup.sh`, then `kubin create`) or `ui/src/pages/overview.tsx`; tear the demo down afterwards.
- No ConfigMap links or "Service selects no pods": those moved to m-2 (TASK-38, TASK-39) and are not in the video.
- "Don't claim" must match README "What does not": auth, private links, expiry, a hosted instance and a release image don't exist yet.
- Keep it at about 75 seconds.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `demo/video-script.md` follows the beats in the description, in order, with Screen and Say for each section
- [ ] #2 The section time ranges start at 0:00 and end between 1:10 and 1:20
- [ ] #3 A Say line states that secrets and env var values stay out of the snapshot
- [ ] #4 The UI part shows only the overview, one pod's logs and the kind sidebar; `grep -niE "configmap|root cause|selects no pods" demo/video-script.md` finds no walkthrough of a root cause
- [ ] #5 Every resource name and command in the script exists in `demo/checkout.yaml` or `demo/setup.sh`, and `grep -n "payment-api\|run.sh" demo/video-script.md` finds nothing outside the closing self-host line
- [ ] #6 The "Don't claim" section matches the README's "What does not" and drops things that now exist
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 The task's commit changes only `demo/video-script.md` (check with `git show --stat HEAD`)
- [ ] #2 If the demo was started to check the clicks, it is torn down (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`)
<!-- DOD:END -->
