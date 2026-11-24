---
id: TASK-47
title: Upload to the public instance when no server is given
status: To Do
assignee: []
created_date: '2026-11-24T09:36:00Z'
updated_date: '2026-11-24T10:50:00Z'
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
doc-1's goal is "run `kubin create`, get a link". Today `kubin create` fails without `--server` or `KUBIN_SERVER`. decision-11 (accepted) makes the public instance `https://kubin.ender.fi` (TASK-23, decision-12) the default, with a confirmation the first time so nobody uploads a cluster to a public server without noticing. Read decision-11 before starting.

## Where

- `cli/cmd/create.go`: the `--server` flag default (`os.Getenv("KUBIN_SERVER")`), the "no kubin server" error in `RunE`, and the step between collecting (`Collecting the cluster...`) and uploading (`Uploading %d resources...`), where the summary and the question go. Add a `defaultServer` const, `https://kubin.ender.fi`, overridable in tests.
- A small package for the remembered answer, like `cli/pkg/userconfig`: `<os.UserConfigDir()>/kubin/config.json`. TASK-45 keeps delete tokens in `snapshots.json` in the same directory; keep the same file rules.
- `cli/cmd/create_test.go`, `README.md` ("What works"), the `--help` text of `create`.

## Spec

- No `--server` and no `KUBIN_SERVER`: the upload goes to `defaultServer`.
- In that case, after collecting and before uploading, stderr shows one summary line: context, API server host, namespaces, resources, log files, compressed size, and the public instance. The real snapshot from the demo cluster gives something like `Collected kind-kubin-demo (https://127.0.0.1:41231): 6 namespaces, 61 resources, 24 logs, 70 KB`.
- With no remembered yes, it then asks on stderr: upload to the public instance kubin.ender.fi, anyone with the link can open it until it expires, `--server` or `KUBIN_SERVER` for your own kubin, `[y/N]`. Only `y` or `yes` (any case) uploads; anything else, including an empty line or EOF, stops with a non-zero exit and uploads nothing. A yes is written to the config file, and later runs print the summary line but don't ask.
- When stdin is not a terminal and there is no remembered yes, it doesn't ask: without `--yes` it exits non-zero with a message naming `--yes`, `--server` and `KUBIN_SERVER`. `--yes` uploads without asking and doesn't write the config.
- `--server` or `KUBIN_SERVER` (whatever URL, the public one too) and `--output` behave as today: no summary, no prompt.
- Stdout stays only the link.

## Out of scope

Server changes, redaction (decision-5 stands), the expiry flag and `kubin delete` (TASK-45), a command to reset the answer (deleting the file is enough; the docs say where it is).

## Traps

- The flag's default value shows in `--help`. Keep the help clear about which server is used when, and keep the default out of the flag's default value so an explicit `--server` can be told apart.
- Tests must never reach the real public instance. Point `defaultServer` at an `httptest` server in tests.
- Tests set `XDG_CONFIG_HOME` and `HOME` with `t.Setenv` to a temp dir, and feed stdin from a buffer, with the terminal check injectable.
- Config file mode 0600, directory 0700, written to a temp file and renamed.
- The prompt reads from stdin while the snapshot sits in a temp file: clean it up when the answer is no.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 With no `--server` and no `KUBIN_SERVER`, `kubin create` uploads to `https://kubin.ender.fi` after a yes
- [ ] #2 The first time, after collecting and before uploading, stderr shows the summary line and the `[y/N]` question; anything but y or yes, including EOF, uploads nothing and exits non-zero
- [ ] #3 A yes is stored under the user config dir and later runs upload without asking, still printing the summary line
- [ ] #4 Without a terminal and without a stored yes, it refuses with a message naming `--yes`, `--server` and `KUBIN_SERVER`; `--yes` uploads without asking or storing anything
- [ ] #5 `--server`, `KUBIN_SERVER` and `--output` behave as before with no prompt, and stdout is only the link in every case
- [ ] #6 README and `kubin create --help` describe the default, the question, `--yes` and where the answer is stored
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd cli && go test ./... && golangci-lint run ./... passes
- [ ] #2 No test reaches the real public instance: go test with networking blocked (for example unshare -rn go test ./cmd/) still passes
<!-- DOD:END -->
