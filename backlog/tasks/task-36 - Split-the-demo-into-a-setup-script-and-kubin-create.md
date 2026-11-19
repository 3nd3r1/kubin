---
id: TASK-36
title: Split the demo into a setup script and kubin create
status: Done
assignee:
  - '@claude'
created_date: '2026-11-19T14:45:00Z'
updated_date: '2026-11-19T15:06:00Z'
labels:
  - demo
  - tier:3
milestone: m-0
dependencies:
  - TASK-35
references:
  - demo/run.sh
  - shell.nix
priority: high
ordinal: 36000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

In the new demo story (user, 2026-11-19) the user has k9s open on the cluster and runs `kubin create` on camera. So the cluster, the kubin server and a `kubin` binary on PATH have to be ready before recording, without a snapshot taken yet. Today `demo/run.sh` does everything in one go, including `kubin create`, and keeps the kubeconfig in a temp dir it deletes on exit, so k9s can't see the cluster afterwards. The README's one-command reproduction (`./demo/run.sh`) must keep going from nothing to a link.

## Where

- `demo/setup.sh` (new): everything `demo/run.sh` does today up to, but not including, `kubin create`. Keep its structure: tool checks, refusing an existing cluster, container or busy port, the cleanup trap on failure, `kind create cluster --wait 2m`, `docker run` with the `kubin-demo-data` volume and `SERVER_PUBLIC_URL`, the `/readyz` retry and the apply retry loop.
  - Apply `demo/checkout.yaml` (from the manifests task) instead of `demo/crashing-pod.yaml`.
  - Wait, with a timeout and an error that says which `kubectl` command to run, until the problems are visible: api pod restarted at least once, `postgres-0` has a FailedScheduling event, Job `worker` has condition Failed, and Deployment `web` is 2/2 ready.
  - Build the CLI into `demo/bin/kubin` (`cd cli && go build -o ../demo/bin/kubin .`).
  - Leave the kind context as the current context in the user's kubeconfig (what plain `kind create cluster` does: it writes `kind-kubin-demo` into `$KUBECONFIG` or `~/.kube/config` and switches to it), so k9s and `kubin create` find the cluster with no flags. Say in the output that the context was switched.
  - End by printing the next steps: an `export PATH="<repo>/demo/bin:$PATH" KUBIN_SERVER=http://localhost:<port>` line, `kubin create`, `k9s -n checkout`, and the two teardown commands.
- `demo/run.sh`: runs `setup.sh`, then `demo/bin/kubin create --server "$server"`, then prints the teardown commands as it does today. If `kubin create` fails, leave the cluster and server running and print the teardown commands.
- `demo/crashing-pod.yaml`: delete.
- `.gitignore`: add `/demo/bin/`.
- `shell.nix` shellHook: add `$PWD/demo/bin` to PATH next to `$PWD/tools/bin`, so inside `nix-shell` the demo's `kubin` is found once setup has built it.

## Out of scope

- Don't set `KUBIN_SERVER` in `shell.nix`: the CLI uses it as the default of `--server`, which would change `kubin create` for every dev shell.
- README text and the GIF (a later task), `demo/video-script.md`, the k9s package (its own task), CLI, server and UI code.
- TASK-30: it is the user's and stays as it is.

## Traps

- A `kubin-demo` kind cluster and container from the old crashing-pod demo may already be running on this machine. They are leftovers of the demo this task replaces; delete them (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`) to test from a clean state, and say so in the notes.
- `setup.sh` must set its "finished" flag before exiting 0, or the cleanup trap deletes the cluster it just built.
- `kind create cluster` merges into `$KUBECONFIG` when it is set. Respect that instead of hard-coding `~/.kube/config`.
- CrashLoopBackOff back-off grows, but the first restart comes within about 20 seconds. Give the waits about 2 minutes.
- The `kubin-demo-data` volume outlives the container, so snapshots from earlier runs stay on the server. That's fine.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 From a clean state (no `kubin-demo` cluster or container), `./demo/setup.sh` exits 0 and leaves the kind cluster and the kubin container running, the checkout app applied with its four problems visible, and `demo/bin/kubin` built
- [x] #2 `demo/setup.sh` never runs `kubin create`: `grep -n "kubin create" demo/setup.sh` shows only lines inside printed instructions (echo or heredoc), and setup prints no snapshot link
- [x] #3 After setup, in a new shell, `kubectl config current-context` prints `kind-kubin-demo`, and `PATH="$PWD/demo/bin:$PATH" KUBIN_SERVER=http://localhost:8080 kubin create` prints a `http://localhost:8080/s/<id>` link
- [x] #4 From a clean state, `./demo/run.sh` sets up and runs `kubin create` and prints the link, so the README's three commands still work unchanged
- [x] #5 Setup ends by printing the export line for PATH and KUBIN_SERVER, `kubin create`, `k9s -n checkout` and the teardown commands
- [x] #6 `demo/crashing-pod.yaml` is deleted and `git grep -n "crashing-pod\|payment-api" -- demo/*.sh` finds nothing
- [x] #7 `/demo/bin/` is gitignored and, after setup, `nix-shell --run "command -v kubin"` prints the path of `demo/bin/kubin`
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 `bash -n demo/setup.sh demo/run.sh` and `nix-shell -p shellcheck --run "shellcheck demo/setup.sh demo/run.sh"` pass
- [x] #2 Ran `./demo/setup.sh` and then `./demo/run.sh` from a clean state on this machine, output summarised in the implementation notes, and tore both down afterwards (`docker rm -f kubin-demo; kind delete cluster --name kubin-demo`)
- [x] #3 The task's commits change only `demo/`, `shell.nix` and `.gitignore` (check with `git log --stat`)
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Removed leftover kubin-demo cluster and container from the old demo before testing. Ran setup.sh (rc 0, context kind-kubin-demo, kubin create printed a link), tore down, ran run.sh (rc 0, link printed), tore down. nix-shell finds demo/bin/kubin. Shellcheck and bash -n pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added demo/setup.sh (cluster, server, checkout app, waits for the four problems, builds demo/bin/kubin, prints next steps). run.sh now calls it then runs kubin create and always prints teardown. Deleted crashing-pod.yaml, gitignored /demo/bin/, added demo/bin to the shell.nix PATH.
<!-- SECTION:FINAL_SUMMARY:END -->
