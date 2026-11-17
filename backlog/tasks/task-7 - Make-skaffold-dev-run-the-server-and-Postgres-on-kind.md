---
id: TASK-7
title: Make the Helm chart deploy kubin with Postgres and S3
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T10:11:00Z'
labels:
  - deploy
  - bug
milestone: m-3
dependencies:
  - TASK-25
  - TASK-26
  - TASK-20
priority: medium
type: bug
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The chart is the Kubernetes way to self-host and the base for a hosted instance. It has never deployed a working pod. `helm template` renders the image as `kubin-api-gateway/kubin-api-gateway:dev` because the deployment template uses image.repository twice. The liveness and readiness probes exec `/api-gateway healthcheck`, a subcommand that does not exist, so running it starts a second server, fails to bind and the pod restarts. skaffold.yaml port-forwards Postgres and Redis, which the chart never installs. `charts/kubin/charts/` has both the unpacked subchart and a stale .tgz of it. With decision-2 there is one image, so the chart needs one deployment, not a subchart per service.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 On a fresh kind cluster `skaffold dev` brings up kubin with Postgres and MinIO and kubin becomes Ready
- [ ] #2 Probes use HTTP /healthz and /readyz, and /readyz checks the database and blob store
- [ ] #3 The rendered image name matches what skaffold builds
- [ ] #4 Values allow SQLite with a PVC for one replica, or external Postgres and S3 for several
- [ ] #5 docs/self-hosting.md covers installing the chart
<!-- AC:END -->
