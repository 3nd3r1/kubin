---
id: TASK-41
title: Re-record the demo GIF and update the README for the checkout demo
status: To Do
assignee: []
created_date: '2026-11-19T14:46:00Z'
labels:
  - docs
  - tier:3
milestone: m-0
dependencies:
  - TASK-35
  - TASK-36
  - TASK-38
  - TASK-39
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

The README's GIF and opening sentence still show the crashing `payment-api` pod. The demo moved to the broken checkout app (user, 2026-11-19), and the viewer has grown since the README was written: kind sidebar, detail pages, related links and the problems-first overview landed in m-2 (TASK-15 to TASK-18, TASK-32, TASK-33), but "What does not" still says the page shows only pods. Someone landing on the repo should see the new story in the GIF and read claims that are true.

## Where

- `demo/kubin-demo.gif`: re-record. The recipe from TASK-30's notes: run `./demo/run.sh`, take full-size browser screenshots of each view, crop to the content column, join them with ffmpeg (palettegen, about 10 fps, 880 px wide). The terminal frame with `kubin create` and the link comes first. Follow the beats of `demo/video-script.md` (rewritten in the task before this one), without the k9s and ticket parts.
- `README.md`: the one-sentence opener, the GIF's alt text, the paragraph under the three commands (mention `demo/setup.sh` for running `kubin create` yourself, with k9s from `nix-shell`), and "What works" / "What does not".

## Out of scope

- The narrated-video comment placeholder at the top of the README and TASK-30: the user's.
- The three reproduction commands (`git clone`, `cd kubin`, `./demo/run.sh`): keep them as they are.
- Any code, scripts or manifests. If something in the demo is broken, write a draft (`backlog draft create`) or set this task to Blocked.

## Traps

- There is no screen recorder here. Earlier runs used browser screenshots; Chrome tools or headless chromium from `nix-shell -p chromium` work. ffmpeg comes from `nix-shell -p ffmpeg`.
- Claim nothing the demo or the code doesn't show. No auth, no expiry, no release image, no hosted instance yet.
- Tear the demo down afterwards (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `demo/kubin-demo.gif` shows, in order: the terminal with `kubin create` and its link, the overview with problems first, the api pod log with `cannot connect to postgres at db:5432`, the `api-config` ConfigMap with `DB_HOST: db`, and the `web` Service with no matching pods
- [ ] #2 The GIF is 880 px wide and under 1 MB (`nix-shell -p ffmpeg --run "ffprobe -v error -show_entries stream=width -of csv=p=0 demo/kubin-demo.gif"` and `du -h demo/kubin-demo.gif`)
- [ ] #3 The README's opening sentence and the GIF's alt text tell the new story, the three reproduction commands are unchanged, and the paragraph below them mentions `demo/setup.sh` for running `kubin create` yourself
- [ ] #4 "What works" and "What does not" are true today: no "shows only pods" claim, nothing the demo or code doesn't show, and `grep -n "payment-api\|crashing pod" README.md` finds nothing
- [ ] #5 The narrated-video placeholder comment at the top of the README is still there
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Ran `./demo/run.sh` from a clean state to record, and tore it down afterwards (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`)
- [ ] #2 The task's commits change only `README.md` and `demo/kubin-demo.gif` (check with `git log --stat`)
<!-- DOD:END -->
