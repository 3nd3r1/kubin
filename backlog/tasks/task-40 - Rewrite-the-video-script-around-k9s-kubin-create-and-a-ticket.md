---
id: TASK-40
title: Rewrite the video script around k9s, kubin create and a ticket
status: To Do
assignee: []
created_date: '2026-11-19T14:46:00Z'
labels:
  - docs
  - tier:2
milestone: m-0
dependencies:
  - TASK-35
  - TASK-36
  - TASK-37
  - TASK-38
  - TASK-39
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

The user changed the demo story on 2026-11-19, and `demo/video-script.md` still tells the crashing `payment-api` story. The user records the narrated video from this script (TASK-30, owned by the user). New story: they have k9s open on their own machine and see problems. Instead of copying output into a ticket, they run one command, `kubin create`, open the link, click around, and paste the link into a ticket (Linear or Jira, which the user shows on camera).

## Where

`demo/video-script.md` only. Keep its shape: a short intro (length, face bubble, captions), a "Before recording" paragraph, numbered sections with time ranges, each with **Screen** and **Say**, and a "Don't claim" section at the end. Keep the plain voice of the current script.

Beats, in order:

1. k9s on namespace `checkout` shows the problems: `api` in CrashLoopBackOff, `postgres-0` Pending, `worker` failed, while `web` looks healthy.
2. The usual way: copying `kubectl describe` and logs out of the terminal into the ticket, and whoever picks it up only sees what you copied.
3. One command in a terminal: `kubin create`, which prints the link.
4. Click through the link: the overview with problems first; the api pod and its log line `cannot connect to postgres at db:5432`; owner links pod to ReplicaSet to Deployment to ConfigMap `api-config`, which says `DB_HOST=db` while the Service is called `postgres`; then the `web` Service whose selector matches no pods.
5. Paste the link into the ticket. Close on self-hosting: one container.

"Before recording" names the real commands: `./demo/setup.sh`, the `export` line it prints, `k9s -n checkout`, the browser zoom and history advice already there.

## Out of scope

- TASK-30 and its criteria: the user's task.
- README and GIF (the next task), demo scripts and manifests.

## Traps

- Every click the script asks for must exist in the UI. Check against the running demo (`./demo/setup.sh`, then `kubin create`) or the code in `ui/src/components/resource/related.tsx`; tear the demo down afterwards.
- "Don't claim" must match what is true today. Other kinds, Services, the overview and `--namespace` exist now (m-2, README "What works"); auth, private links, expiry, a hosted instance and a release image still don't. Read README "What does not".
- Keep it at about 75 seconds. More than three problems on screen makes the video a tour, not a story.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `demo/video-script.md` follows the five beats in the description, in order, with Screen and Say for each section
- [ ] #2 The section time ranges start at 0:00 and end between 1:10 and 1:20
- [ ] #3 The Say lines name at most three of the planted problems, and two of them are the api ConfigMap (`DB_HOST=db` against the `postgres` Service) and the `web` Service that selects no pods
- [ ] #4 Every resource name and command in the script exists in `demo/checkout.yaml` or `demo/setup.sh`, and `grep -n "payment-api\|run.sh" demo/video-script.md` finds nothing outside the closing self-host line
- [ ] #5 Every click the script asks for is a link the UI renders today, checked against a running demo or `ui/src/components/resource/related.tsx` and `ui/src/pages/overview.tsx`
- [ ] #6 The "Don't claim" section matches the README's "What does not" and drops things that now exist
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 The task's commit changes only `demo/video-script.md` (check with `git show --stat HEAD`)
- [ ] #2 If the demo was started to check the clicks, it is torn down (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`)
<!-- DOD:END -->
