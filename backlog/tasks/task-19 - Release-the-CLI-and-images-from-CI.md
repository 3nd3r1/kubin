---
id: TASK-19
title: Release the CLI and images from CI
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-17T10:11:00Z'
labels:
  - deploy
milestone: m-3
dependencies:
  - TASK-8
priority: high
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The README tells people to `go install`, which needs a Go toolchain. A tag should produce binaries and images.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Pushing a v* tag publishes CLI binaries for linux, macOS and windows on amd64 and arm64 to GitHub releases
- [ ] #2 The kubin image is pushed to ghcr.io/3nd3r1/kubin with the tag and latest
- [ ] #3 `kubin version` prints the version and commit
- [ ] #4 The README install section shows a one-line install
<!-- AC:END -->
