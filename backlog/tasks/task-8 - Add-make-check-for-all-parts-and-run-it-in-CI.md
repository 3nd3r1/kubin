---
id: TASK-8
title: Add make check for all parts and run it in CI
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
labels:
  - deploy
milestone: m-0
dependencies: []
priority: medium
type: chore
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
There is no single command that says the repo is healthy and no CI runs tests. The root Makefile only has a help target. Agents and the user need one check to run before every commit.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `make check` at the root runs golangci-lint and go test for cli and server, and lint, typecheck and build for ui
- [ ] #2 A GitHub Actions workflow runs `make check` on pushes and pull requests
- [ ] #3 The repo passes it
<!-- AC:END -->
