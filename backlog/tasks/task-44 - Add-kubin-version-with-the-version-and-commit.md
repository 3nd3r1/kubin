---
id: TASK-44
title: Add kubin version with the version and commit
status: To Do
assignee: []
created_date: '2026-11-24T09:36:00Z'
labels:
  - cli
  - tier:2
milestone: m-3
dependencies: []
priority: high
type: feature
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Released binaries (TASK-19) need to say which build they are, so a bug report can name the exact commit. `cli/cmd/root.go` has `var version = "dev"`, set by ldflags, and cobra's `--version` prints it. Nothing records the commit and there is no `kubin version` command.

## Where

- `cli/cmd/root.go`: the `version` var and `rootCmd.Version`.
- New `cli/cmd/version.go` and `cli/cmd/version_test.go`. Follow the command style of `cli/cmd/create.go`.
- `cli/Makefile`: the `build` target.

## Spec

- Package `cmd` gets `var commit = "none"` next to `version`. Both are set with `-ldflags "-X github.com/3nd3r1/kubin/cli/cmd.version=v1.2.3 -X github.com/3nd3r1/kubin/cli/cmd.commit=<sha>"`. TASK-19's release config uses exactly these two names.
- Without ldflags, fill them from `runtime/debug.ReadBuildInfo()`: the main module version when it isn't `(devel)`, and the `vcs.revision` setting cut to 12 characters. Put this in a small pure function that the test can call with a fake build info.
- `kubin version` prints one line to stdout: `kubin v1.2.3 (commit 0123456789ab)`. `kubin --version` prints the same line (`rootCmd.SetVersionTemplate`).
- `make -C cli build` stamps `git describe --tags --always --dirty` and `git rev-parse --short=12 HEAD`.

## Out of scope

The release workflow (TASK-19), the server's version, the upload User-Agent (it keeps using `version`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `kubin version` prints one line, `kubin <version> (commit <commit>)`, to stdout and exits 0
- [ ] #2 `-X` ldflags for `cmd.version` and `cmd.commit` set both values, and a test covers the output with both set
- [ ] #3 Without ldflags, the version and commit come from the Go build info when it has them, else `dev` and `none`; a test covers this with a fake build info
- [ ] #4 `kubin --version` prints the same line as `kubin version`
- [ ] #5 `make -C cli build` stamps the version from `git describe --tags --always --dirty` and the 12-character commit
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd cli && go test ./... && golangci-lint run ./... passes
- [ ] #2 make -C cli build && ./cli/kubin-cli version prints the commit that git rev-parse --short=12 HEAD prints
<!-- DOD:END -->
