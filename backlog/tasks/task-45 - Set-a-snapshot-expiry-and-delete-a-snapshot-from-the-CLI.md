---
id: TASK-45
title: Set a snapshot expiry and delete a snapshot from the CLI
status: To Do
assignee: []
created_date: '2026-11-24T09:36:00Z'
labels:
  - cli
  - tier:3
milestone: m-3
dependencies:
  - TASK-21
priority: high
type: feature
ordinal: 7500
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
TASK-21 gives snapshots an expiry and a delete token on the server. Users need to give a sensitive dump a shorter life, and to remove a snapshot they shared by mistake. Read the API section of `docs/architecture.md` after TASK-21 for the exact parameter and endpoint; the task proposes `?expires_in=<duration>` on upload, `expires_at` and `delete_token` in the response, and `DELETE /api/v1/snapshots/{id}` with `Authorization: Bearer <token>`.

## Where

- `cli/cmd/create.go`: the `--expire` flag and the stderr lines after upload.
- `cli/pkg/upload/upload.go`: send the expiry, read `expires_at` and `delete_token` into `Result`, add a `Delete` call. Tests in `upload_test.go` use `httptest` servers; follow them.
- New `cli/cmd/delete.go` and a small package for the token file, like `cli/pkg/tokens`.
- `README.md`, "What works".

## Spec

- `kubin create --expire 24h` sends that expiry. Without the flag nothing is sent and the server default applies. The server's 400 above its maximum is shown as the server words it.
- After upload, stderr (`log.Info`) says when it expires and how to delete it: `Expires 2026-12-01 10:00 UTC. Delete it with: kubin delete <url>`. Stdout stays only the link (the command's Long help says so).
- The token goes to `<os.UserConfigDir()>/kubin/snapshots.json`, keyed by snapshot URL: file mode 0600, directory 0700, written to a temp file and renamed.
- `kubin delete <url>`: split the link at the last `/s/` into server base and ID, take the token from the file or `--token`, send the DELETE. On 204, or a 404 for a snapshot that's already gone, remove the stored entry. A 403 (wrong token) and a missing token give clear errors and a non-zero exit.
- `--output` writes a file: no expiry, no token.

## Out of scope

Server and UI changes, listing your snapshots.

## Traps

- The delete token is a secret. Never print it to stdout or in logs.
- Tests set `XDG_CONFIG_HOME` and `HOME` with `t.Setenv` to a temp dir, so they never touch the real config.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `kubin create --expire 24h` uploads with that expiry, and stderr shows when the snapshot expires and the `kubin delete` command; stdout is only the link
- [ ] #2 The delete token is stored in a 0600 file under the user config dir and is never printed
- [ ] #3 `kubin delete <url>` deletes the snapshot with the stored token or `--token`, and removes the stored entry
- [ ] #4 A wrong token, a missing token and an unknown snapshot each give a clear error; the first two exit non-zero
- [ ] #5 Tests against an httptest server cover create with --expire, delete success and each error case
- [ ] #6 README "What works" mentions `--expire` and `kubin delete`
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd cli && go test ./... && golangci-lint run ./... passes
- [ ] #2 Manual: cd server && make run, then cd cli && go run . create --server http://localhost:8080 --expire 1h -n kube-system against a kind cluster, then go run . delete <url>; the link then answers 404
<!-- DOD:END -->
