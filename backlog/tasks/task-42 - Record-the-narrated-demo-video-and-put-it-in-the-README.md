---
id: TASK-42
title: Record the narrated demo video and put it in the README
status: Done
assignee:
  - '@human'
created_date: '2026-11-19T14:57:00Z'
updated_date: '2026-11-24T08:16:00Z'
labels:
  - human
milestone: m-0
dependencies:
  - TASK-35
  - TASK-36
  - TASK-37
  - TASK-40
  - TASK-41
ordinal: 42000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The README should open with a narrated video by the user, face bubble included, above the GIF (split from TASK-30). Only the user can do this. Workers must not pick it up, so it has no tier.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A free throwaway Linear workspace exists with an example ticket
- [x] #2 k9s is installed or opened from nix-shell and ./demo/setup.sh has run
- [x] #3 A narrated video of about 75 seconds with a face bubble and burned-in captions is recorded from the Say lines of demo/video-script.md (TASK-40)
- [x] #4 The kubin link is pasted into the Linear ticket on camera
- [x] #5 The final GIF from TASK-41 is reviewed
- [x] #6 The mp4 is dragged into the GitHub README editor and plays inline above the GIF at the placeholder comment
<!-- AC:END -->
