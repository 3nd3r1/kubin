---
id: TASK-41
title: Re-record the demo GIF and update the README for the checkout demo
status: Done
assignee:
  - '@claude'
created_date: '2026-11-19T14:46:00Z'
updated_date: '2026-11-19T15:11:00Z'
labels:
  - docs
  - tier:3
milestone: m-0
dependencies:
  - TASK-35
  - TASK-36
  - TASK-40
references:
  - README.md
  - demo/kubin-demo.gif
  - demo/video-script.md
priority: high
ordinal: 41000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

The README's GIF and opening sentence still show the crashing `payment-api` pod. The demo story is now (user, 2026-11-19, decision-8): a broken cluster, one `kubin create` that captures everything non-secret and prints a link, the link goes into a ticket, then a quick look at the UI. The viewer has grown since the README was written: kind sidebar, detail pages and the problems-first overview landed in m-2 (TASK-15 to TASK-18, TASK-32, TASK-33), but "What does not" still says the page shows only pods.

## Where

- `demo/kubin-demo.gif`: re-record. The recipe from TASK-30's notes: run `./demo/run.sh`, take full-size browser screenshots of each view, crop to the content column, join them with ffmpeg (palettegen, about 10 fps, 880 px wide). Follow `demo/video-script.md` (TASK-40), without the k9s and ticket parts: the terminal with `kubin create` and its link, the overview with problems first, the api pod's logs, the kind sidebar.
- `README.md`: the one-sentence opener (capture a broken cluster in one command, share the link in a ticket instead of a log dump), the GIF's alt text, the paragraph under the three commands (mention `demo/setup.sh` for running `kubin create` yourself, with k9s from `nix-shell`), and "What works" / "What does not". Say that secrets and env var values stay out.

## Out of scope

- The narrated-video comment placeholder at the top of the README and TASK-30: the user's.
- The three reproduction commands (`git clone`, `cd kubin`, `./demo/run.sh`): keep them as they are.
- Any code, scripts or manifests. If something in the demo is broken, write a draft (`backlog draft create`) or set this task to Blocked.

## Traps

- There is no screen recorder here. Earlier runs used browser screenshots; Chrome tools or headless chromium from `nix-shell -p chromium` work. ffmpeg comes from `nix-shell -p ffmpeg`.
- Claim nothing the demo or the code doesn't show. No ConfigMap links or "Service selects no pods" (m-2), no auth, no expiry, no release image, no hosted instance.
- Tear the demo down afterwards (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `demo/kubin-demo.gif` shows, in order: the terminal with `kubin create` and its link, the overview with problems first, the api pod log with `cannot connect to postgres at db:5432`, and the kind sidebar
- [x] #2 The GIF is 880 px wide and under 1 MB (`nix-shell -p ffmpeg --run "ffprobe -v error -show_entries stream=width -of csv=p=0 demo/kubin-demo.gif"` and `du -h demo/kubin-demo.gif`)
- [x] #3 The README's opening sentence and the GIF's alt text tell the new story, the three reproduction commands are unchanged, and the paragraph below them mentions `demo/setup.sh` for running `kubin create` yourself
- [x] #4 "What works" and "What does not" are true today: no "shows only pods" claim, nothing the demo or code doesn't show, and `grep -n "payment-api\|crashing pod" README.md` finds nothing
- [x] #5 The narrated-video placeholder comment at the top of the README is still there
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Ran `./demo/run.sh` from a clean state to record, and tore it down afterwards (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`)
- [x] #2 The task's commits change only `README.md` and `demo/kubin-demo.gif` (check with `git log --stat`)
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Re-recorded demo/kubin-demo.gif (880x570, 148K): terminal with kubin create and link, overview problems first, api pod log, kind sidebar. Built from headless chromium screenshots and ffmpeg after a clean ./demo/run.sh; demo torn down. README opener, alt text, setup.sh paragraph, What works / What does not updated; placeholder comment kept.
<!-- SECTION:FINAL_SUMMARY:END -->
