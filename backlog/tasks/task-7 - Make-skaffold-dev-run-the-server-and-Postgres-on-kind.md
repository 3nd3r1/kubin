---
id: TASK-7
title: Make skaffold dev run the server and Postgres on kind
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
labels:
  - deploy
  - bug
milestone: m-0
dependencies:
  - TASK-1
priority: high
type: bug
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The chart has never deployed a working pod. `helm template` renders the image as `kubin-api-gateway/kubin-api-gateway:dev` because the deployment template uses image.repository twice. The liveness and readiness probes exec `/api-gateway healthcheck`, a subcommand that does not exist, so running it starts a second server, fails to bind and the pod restarts. skaffold.yaml port-forwards Postgres and Redis, which the chart never installs. `charts/kubin/charts/` has both the unpacked subchart and a stale .tgz of it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 On a fresh kind cluster `skaffold dev` brings up the server and Postgres and the server becomes Ready
- [ ] #2 Probes use HTTP /healthz and /readyz, and /readyz checks the database
- [ ] #3 The rendered image name matches what skaffold builds
- [ ] #4 Postgres is a chart dependency with dev credentials in values-dev.yaml
- [ ] #5 The UI is added to the chart or documented as run with npm run dev against the forwarded server
- [ ] #6 docs/development.md describes the loop and no longer warns that the chart is broken
<!-- AC:END -->
