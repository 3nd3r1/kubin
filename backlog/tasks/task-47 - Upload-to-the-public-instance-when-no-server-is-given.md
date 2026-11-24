---
id: TASK-47
title: Upload to the public instance when no server is given
status: To Do
assignee: []
created_date: '2026-11-24T09:36:00Z'
labels:
  - cli
  - tier:3
milestone: m-3
dependencies:
  - TASK-23
priority: medium
type: feature
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
doc-1's goal is "run `kubin create`, get a link". Today `kubin create` fails without `--server` or `KUBIN_SERVER`. decision-11 (proposed) recommends uploading to the public instance (TASK-23, decision-10) when neither is set. Do this task only once decision-11 is accepted; if it is still proposed, set the task Blocked with a comment pointing at it. If decision-11 is rejected, the orchestrator archives this task.

## Where

- `cli/cmd/create.go`: the `--server` flag default (`os.Getenv("KUBIN_SERVER")`) and the "no kubin server" error in `RunE`. Add a `defaultServer` const with the public URL from decision-10.
- `cli/cmd/create_test.go`, `README.md` ("What works"), the `--help` text of `create`.

## Spec

- No `--server` and no `KUBIN_SERVER`: upload to `defaultServer`.
- In that case, before collecting, stderr (`log.Info`) says the snapshot goes to a public kubin instance, that anyone with the link can open it until it expires, and that `--server` or `KUBIN_SERVER` picks your own server.
- Stdout stays only the link. `--server`, `KUBIN_SERVER` and `--output` behave as today.

## Out of scope

Server changes, redaction (decision-5 stands), the expiry flag (TASK-45).

## Traps

- The flag's default value shows in `--help`. Keep the help clear about which server is used when.
- Tests must never reach the real public instance. Make the default overridable in tests and point it at an `httptest` server.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 With no `--server` and no `KUBIN_SERVER`, `kubin create` uploads to the public instance URL from decision-10
- [ ] #2 In that case stderr says, before collection starts, that the snapshot goes to a public instance and anyone with the link can open it until it expires; stdout is only the link
- [ ] #3 `--server`, `KUBIN_SERVER` and `--output` behave as before, covered by tests
- [ ] #4 README and `kubin create --help` describe the default
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd cli && go test ./... && golangci-lint run ./... passes
- [ ] #2 No test reaches the real public instance: go test with networking blocked (for example unshare -rn go test ./cmd/) still passes
<!-- DOD:END -->
