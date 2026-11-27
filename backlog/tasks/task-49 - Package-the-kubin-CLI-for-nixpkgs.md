---
id: TASK-49
title: Package the kubin CLI for nixpkgs
status: In Progress
assignee:
  - '@agy'
created_date: '2026-11-26T13:06:00Z'
updated_date: '2026-11-27T12:28:00Z'
labels:
  - cli
  - deploy
dependencies: []
references:
  - cli/go.mod
  - .goreleaser.yaml
  - README.md
priority: low
type: enhancement
ordinal: 44000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Nix and NixOS users currently have to install the kubin CLI via curl|sh or build from source. Adding kubin to nixpkgs lets users install it declaratively or run it via nix-shell -p kubin / nix run nixpkgs#kubin. The CLI lives in a nested Go module (cli/go.mod, pure Go with CGO_ENABLED=0), and the repo currently lacks a root LICENSE file (README.md declares MIT) and a tagged GitHub release, both of which nixpkgs expects.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A root LICENSE file (MIT) is added to the repository matching README.md
- [x] #2 A tagged GitHub release (v*) exists for nixpkgs to fetch from
- [ ] #3 A pkgs/by-name/ku/kubin/package.nix derivation builds the kubin binary from cli/ with tests enabled in the Nix sandbox and prints the release version via kubin version
- [ ] #4 A pull request adding kubin is submitted to NixOS/nixpkgs and passes local nix-build / nixpkgs-review checks
- [ ] #5 README.md documents installing kubin from nixpkgs once the PR is merged
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add root MIT LICENSE file matching README.md
2. Tag and push v0.1.0 to trigger the GitHub Release + GHCR workflow
3. Create and test the nixpkgs derivation (pkgs/by-name/ku/kubin/package.nix) and submit PR to NixOS/nixpkgs
4. Document nixpkgs installation in README.md once merged
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Added root MIT LICENSE (fad6290) and published v0.1.0 GitHub release + GHCR image (run 45998043680). Verified install.sh downloads kubin v0.1.0.
<!-- SECTION:NOTES:END -->
