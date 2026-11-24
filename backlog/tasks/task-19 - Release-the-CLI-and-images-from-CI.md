---
id: TASK-19
title: Release the CLI and images from CI
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-24T11:48:00Z'
labels:
  - deploy
  - tier:3
milestone: m-3
dependencies:
  - TASK-8
  - TASK-44
priority: high
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
doc-1 m-3 asks for a released CLI and image. Today self-hosters `docker build` from a clone (docs/self-hosting.md) and the README sends people to the demo script. A `v*` tag should publish CLI binaries and a multi-arch image, and installing the CLI should be one line.

## Where

- New `.github/workflows/release.yml`, on push of tags `v*`. Needs `permissions: contents: write, packages: write`.
- CLI binaries: GoReleaser, config `.goreleaser.yaml` at the repo root. Build from `dir: cli`, binary `kubin`, `CGO_ENABLED=0`, linux, darwin and windows on amd64 and arm64. ldflags `-s -w -X github.com/3nd3r1/kubin/cli/cmd.version=v{{.Version}} -X github.com/3nd3r1/kubin/cli/cmd.commit={{.ShortCommit}}` (the names come from TASK-44). tar.gz archives, zip for windows, `checksums.txt`.
- Image: `docker/login-action` to ghcr.io with `GITHUB_TOKEN`, then `docker/build-push-action` with `platforms: linux/amd64,linux/arm64` and tags `ghcr.io/3nd3r1/kubin:<tag>` and `ghcr.io/3nd3r1/kubin:latest`. The image is not built by GoReleaser, because the Dockerfile builds the UI too.
- `Dockerfile`: run the `ui` and `server` stages on `--platform=$BUILDPLATFORM` and cross-compile with `GOOS=$TARGETOS GOARCH=$TARGETARCH`, so arm64 doesn't build node under QEMU.
- New `install.sh` at the repo root: detects OS and arch, downloads the archive of the latest release (or `KUBIN_VERSION`) from GitHub releases, checks it against `checksums.txt`, installs `kubin` into `/usr/local/bin` or `$KUBIN_INSTALL_DIR`. A `KUBIN_DOWNLOAD_URL` override for the base URL lets it be tried against a local `dist/`.
- `README.md`: a one-line install, `curl -fsSL https://raw.githubusercontent.com/3nd3r1/kubin/main/install.sh | sh`, and remove "the repository does not publish a release image yet". `docs/self-hosting.md`: `docker run ... ghcr.io/3nd3r1/kubin` instead of building.
- `.gitignore`: GoReleaser writes `dist/` at the root.
- Add `goreleaser` to `shell.nix`.

## Out of scope

- Pushing a tag or creating a release. The user does that and picks the first version number. Say so in the final summary.
- Publishing the Helm chart, Homebrew, signing.
- The image's runtime contract stays: scratch, uid 1001, `/data` volume, port 8080, `ENTRYPOINT ["/kubin-server"]`. `docker run -v kubin-data:/data` must keep working (CLAUDE.md).

## Traps

- The CLI is a nested module (`cli/go.mod`). A `v1.2.3` tag is not a version of `github.com/3nd3r1/kubin/cli` (that would be `cli/v1.2.3`), so `go install ...@v1.2.3` won't work. Don't promise it in the README.
- GoReleaser's `{{.Version}}` drops the leading `v`.
- In a multi-arch Dockerfile, `ARG TARGETOS TARGETARCH` has to come after the `FROM` that uses them.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `.github/workflows/release.yml` runs on `v*` tags, uploads the GoReleaser archives and checksums to the GitHub release, and pushes `ghcr.io/3nd3r1/kubin:<tag>` and `:latest` for linux/amd64 and linux/arm64
- [x] #2 `goreleaser release --snapshot --clean` builds kubin for linux, darwin and windows on amd64 and arm64, with archives and checksums.txt in dist/
- [x] #3 A binary from that build prints its version and commit with `kubin version`
- [x] #4 The image builds for linux/amd64 and linux/arm64, and `make image` plus the docker run in docs/self-hosting.md still work
- [x] #5 `install.sh` installs the archive for the current OS and arch after checking it against checksums.txt, tried against the snapshot dist/ through `KUBIN_DOWNLOAD_URL`
- [x] #6 README shows the one-line install, and docs/self-hosting.md runs the ghcr.io image instead of building one
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 nix-shell -p goreleaser --run "goreleaser check && goreleaser release --snapshot --clean" passes, and the linux/amd64 binary in dist/ prints a version and commit
- [x] #2 docker buildx build --platform linux/amd64,linux/arm64 . passes
- [x] #3 make image && docker run --rm -d --name kubin-check -p 18080:8080 kubin:dev && curl -fsS localhost:18080/readyz; docker rm -f kubin-check prints ok
- [x] #4 nix-shell -p shellcheck actionlint --run "shellcheck install.sh && actionlint" passes
- [x] #5 make check passes and git status is clean
<!-- DOD:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added .goreleaser.yaml, release.yml (v* tags: GoReleaser CLI job, multi-arch ghcr.io image job), install.sh with checksum check and KUBIN_DOWNLOAD_URL override, cross-compiling Dockerfile (ui/server stages on BUILDPLATFORM), README install line, self-hosting docs using the ghcr.io image, goreleaser in shell.nix, /dist/ in .gitignore. Verified: goreleaser check and snapshot release (6 targets, archives, checksums.txt), linux/amd64 binary prints version and commit, install.sh against dist/ installs a working kubin, shellcheck and actionlint pass, buildx amd64+arm64 passed, make image plus docker run /readyz returns ok, make check passes. Not done by design: no tag was pushed and no release created. The user pushes the first v* tag and picks the version. Note go install ...@vX is not promised since the CLI is a nested module.
<!-- SECTION:FINAL_SUMMARY:END -->
