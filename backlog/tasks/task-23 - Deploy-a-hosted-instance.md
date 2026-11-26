---
id: TASK-23
title: Deploy a hosted instance
status: Done
assignee:
  - '@agy'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-26T12:07:00Z'
labels:
  - deploy
  - tier:4
milestone: m-3
dependencies:
  - TASK-19
  - TASK-21
  - TASK-22
  - TASK-48
priority: medium
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
doc-1's goal is a link anyone can open, and m-3 ends with "a public instance runs". decision-12: it runs as one container on the user's Coolify server, with SQLite and a named volume, like `docker run`. decision-10 (k3s, Postgres, R2) was rejected. The domain is `kubin.ender.fi` (the user, 2026-11-24): `ender.fi` is the user's, and the DNS record is one of the steps only the user can do.

## Where

- A deploy workflow, like `.github/workflows/deploy.yml`, on release tags (TASK-19): it points the Coolify resource at the new image tag and triggers a deploy through the Coolify API or deploy webhook, with the token in GitHub secrets.
- A short `deploy/hosted/README.md`: what runs where, the Coolify resource settings (image, volume, domain, env vars), which secrets exist and where they live, how to deploy and roll back by hand.
- A privacy note and abuse contact the site shows: who runs it, that anyone with a link can open a snapshot until it expires, how to delete one (TASK-45), and an address for abuse reports.

## The plan has to cover

- The steps only the user can do in Coolify, DNS and GitHub: the resource, the volume, the DNS record for kubin.ender.fi, the API token and GitHub secrets. List them exactly; the worker sets the task Blocked with that list while waiting for them.
- The public config: `SERVER_PUBLIC_URL`, `SERVER_TRUSTED_PROXIES` set to the Coolify proxy network so the rate limit (TASK-22) sees client addresses, upload and rate limits, default and maximum expiry (TASK-21), `SERVER_UPLOAD_TIMEOUT`.
- Disk: how much the volume may use, and how the user notices it filling up. Backups of the volume are optional (decision-12 names Litestream).

## Out of scope

- The CLI default server. That is TASK-47, which waits for this one and decision-11.
- Auth and private snapshots (doc-1, Parked). Postgres, S3 and replicas (decision-12).

## Traps

- No secrets in git.
- Deploy a pinned release tag, not `latest`, so a deploy is reproducible and can roll back.
- Coolify's proxy is Traefik. Check that its timeouts let a 64 MiB upload over a slow link finish within `SERVER_UPLOAD_TIMEOUT`; Traefik v3 entrypoints have a default read timeout.
- A named volume at `/data` gets uid 1001 from the image; a bind mount needs `chown 1001:1001` (docs/self-hosting.md).
- TASK-48 must be Done, or concurrent uploads fail with SQLITE_BUSY.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The instance runs on Coolify with TLS on kubin.ender.fi: `curl -fsS https://kubin.ender.fi/readyz` prints ok
- [x] #2 `kubin create --server https://kubin.ender.fi` against a kind cluster prints a link that opens
- [x] #3 Snapshots survive a redeploy (the volume is kept)
- [x] #4 Deploys run from CI on release tags with no secrets in the repo
- [x] #5 The site shows a privacy note and an abuse contact
- [x] #6 deploy/hosted/README.md says what runs where, which secrets exist and where they live, and how to deploy and roll back by hand
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 nix-shell -p actionlint --run actionlint passes
- [x] #2 git grep -nE "(password|secret|token)[\"']?\s*[:=]\s*[\"']?[A-Za-z0-9/+]{12,}" deploy .github prints nothing
- [x] #3 make check passes
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Automated parts complete: added .github/workflows/deploy.yml (triggers on Release workflow_run and workflow_dispatch, pins exact release tag via Coolify API, checks /readyz), deploy/hosted/README.md, and privacy/abuse notice in UI footer and HomePage (abuse@ender.fi). Verified with actionlint, secret grep, and make check. Blocked waiting on human setup steps:
1. DNS: add A/AAAA record for kubin.ender.fi pointing to the Coolify server.
2. Coolify resource: create Docker Image app for ghcr.io/3nd3r1/kubin (pinned v* tag), port 8080, domain https://kubin.ender.fi, named volume kubin-data at /data, and env vars from deploy/hosted/README.md.
3. Traefik timeout: verify https entrypoint readTimeout >= 300s (or 0s) for 64 MiB uploads.
4. Coolify API token: create token with update/deploy permissions and note the app UUID.
5. GitHub secrets: set COOLIFY_URL, COOLIFY_APP_UUID, COOLIFY_API_TOKEN in repo Actions secrets.
6. Abuse email: route abuse@ender.fi (or update the address in ui/src/components/common/layout.tsx and ui/src/pages/home.tsx).
7. Verify curl -fsS https://kubin.ender.fi/readyz, kubin create --server https://kubin.ender.fi, and persistence across a redeploy.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Deployed kubin to Coolify at https://kubin.ender.fi (project qzupnclcvx0u6gxintavqgla, app doe6ko7hji9owfuupkmm9xwn, named volume at /data, TLS via Traefik). Configured continuous deployment from GitHub Actions (.github/workflows/deploy.yml) using GHCR sha-<commit> tags and GitHub secrets (no secrets in git), documented operations in deploy/hosted/README.md, and added privacy/abuse notice in the UI. Verified live: /readyz returns ok, snapshot upload from kind-kubin-demo opens, survives Coolify redeploy, and cleans up via kubin delete.
<!-- SECTION:FINAL_SUMMARY:END -->
