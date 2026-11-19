---
id: TASK-37
title: Add k9s to the nix shell
status: To Do
assignee: []
created_date: '2026-11-19T14:45:00Z'
labels:
  - chore
  - tier:1
milestone: m-0
dependencies: []
references:
  - shell.nix
priority: high
ordinal: 37000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

The new m-0 demo (user, 2026-11-19) opens on k9s showing the broken checkout app before `kubin create`. Everyone using the repo's shell, the user recording the demo included, should get the same k9s from `nix-shell`. CLAUDE.md: tools come from `nix-shell`.

## Where

- `shell.nix`: add `k9s` to `buildInputs` under the "Local cluster and deploys" comment.
- `docs/development.md`: if it lists the shell's tools, add k9s there. Otherwise leave it.

## Out of scope

- The `shellHook`. The setup-script task adds `demo/bin` to PATH there; keep whatever is in it.
- Any other package.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `nix-shell --run "k9s version"` prints a k9s version
- [ ] #2 The diff touches only `shell.nix` and, if it lists the tools, `docs/development.md`
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 `nix-shell --run "k9s version && kind version && kubectl version --client"` succeeds
<!-- DOD:END -->
