---
id: TASK-23
title: Deploy a hosted instance
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-24T09:37:00Z'
labels:
  - deploy
  - tier:4
milestone: m-3
dependencies:
  - TASK-7
  - TASK-19
  - TASK-21
  - TASK-22
priority: medium
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
doc-1's goal is a link anyone can open, and m-3 ends with "a public instance runs". Where it runs, the domain and how deploys happen are decision-10, which is proposed. Don't plan this task until decision-10 is accepted; if it is still proposed, set this task Blocked with a comment pointing at it.

## Where

- The chart from TASK-7, in its Postgres and S3 mode unless decision-10 says otherwise.
- A values file for the instance, like `deploy/hosted/values.yaml`, with no secrets in it.
- A deploy workflow, like `.github/workflows/deploy.yml`, run on release tags (TASK-19), by the method decision-10 names.
- A short `deploy/hosted/README.md`: what runs where, which secrets exist and where they live, how to deploy by hand.

## The plan has to cover

- The steps only the user can do: the account and VM or cluster, the DNS record, cloud credentials, GitHub secrets. List them exactly; the worker sets the task Blocked with that list while waiting for them.
- The public config: `SERVER_PUBLIC_URL`, upload and rate limits (TASK-22), default and maximum expiry (TASK-21), `SERVER_UPLOAD_TIMEOUT` for slow uploads.
- A privacy note and abuse contact on the site (decision-10, consequences).

## Out of scope

- The CLI default server. That is TASK-47, which waits for this one and decision-11.
- Auth and private snapshots (doc-1, Parked).

## Traps

- No secrets in git. The kubeconfig and credentials go into GitHub secrets or the cluster.
- Deploy a pinned release tag, not `latest`, so a deploy is reproducible.
- The ingress has to allow 64 MiB request bodies.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The instance runs from the chart with TLS on the domain in decision-10: `curl -fsS https://<domain>/readyz` prints ok
- [ ] #2 `kubin create --server https://<domain>` against a kind cluster prints a link that opens
- [ ] #3 Deploys run from CI on release tags by the method in decision-10, with no secrets in the repo
- [ ] #4 deploy/hosted/README.md says what runs where, which secrets exist and where they live, and how to deploy by hand
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 nix-shell -p actionlint --run actionlint passes
- [ ] #2 git grep -nE "(password|secret|token)[\"']?\s*[:=]\s*[\"']?[A-Za-z0-9/+]{12,}" deploy .github prints nothing
- [ ] #3 make check passes
<!-- DOD:END -->
