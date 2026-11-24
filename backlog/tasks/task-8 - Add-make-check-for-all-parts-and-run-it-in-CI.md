---
id: TASK-8
title: Add make check for all parts and run it in CI
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-24T09:36:00Z'
labels:
  - deploy
  - tier:2
milestone: m-3
dependencies: []
priority: medium
type: chore
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
No single command says the repo is healthy, and no CI runs the tests. The orchestrator runs each part's checks by hand after every task (CLAUDE.md, decision-7), and the release workflow (TASK-19) builds on a green check.

## Where

- Root `Makefile`: only `help` and `image` today. Add `check`, with a `## ` comment so `make help` lists it. Call the parts with `$(MAKE) -C <dir> <target>`.
- `cli/Makefile` has no `lint` target. Add `lint: golangci-lint run ./...` like `server/Makefile`.
- `server/Makefile` already has `test` and `lint`. `ui/Makefile` has `install`, `lint`, `typecheck`, `test` and `build`.
- New `.github/workflows/ci.yml`.
- `docs/development.md`, "What works today".

## Out of scope

- The Dockerfile, the release workflow (TASK-19) and `apply-component-label.yml` (TASK-9).
- Fixing real bugs. If a check fails for a reason that isn't a one-line lint fix, write a draft (`backlog draft create`) and set this task Blocked with a comment.

## Traps

- `ui` needs `npm ci` before lint, typecheck, test or build. `make check` should not reinstall every run; CI runs `make -C ui install` as its own step first.
- `make -C ui build` writes `server/internal/web/dist`, which is gitignored except the tracked `.gitkeep`. After `make check`, `git status` must be clean.
- `cli` and `server` are separate Go modules with different `go` lines (cli 1.24, server 1.26). In CI use `actions/setup-go` with one Go that satisfies both, or `go-version-file` per module.
- Pin golangci-lint in CI to the version nix-shell has (`golangci-lint --version` inside nix-shell), so CI and local runs agree.
- The UI unit tests (`make -C ui test`, vitest) belong in the check: the orchestrator already runs them.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `make check` at the root runs golangci-lint and go test for cli and server, and lint, typecheck, test and build for ui, and exits non-zero when any of them fails
- [ ] #2 `.github/workflows/ci.yml` installs the UI dependencies and runs `make check` on pushes to main and on pull requests
- [ ] #3 The repo passes it: `make check` exits 0 and leaves `git status` clean
- [ ] #4 `docs/development.md` names `make check` as the command to run before committing
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 make check exits 0, and git status is clean afterwards
- [ ] #2 nix-shell -p actionlint --run "actionlint .github/workflows/ci.yml" passes
<!-- DOD:END -->
