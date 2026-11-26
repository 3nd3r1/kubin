---
id: TASK-7
title: Make the Helm chart deploy kubin
status: Done
assignee:
  - '@agy'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-26T09:35:00Z'
labels:
  - deploy
  - bug
  - tier:3
milestone: m-3
dependencies:
  - TASK-25
priority: medium
type: bug
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The chart is the Kubernetes way to self-host, and it has never deployed a working pod. `helm template` renders the image as `kubin-api-gateway/kubin-api-gateway:dev` because the deployment template uses `image.repository` twice. The probes exec `/api-gateway healthcheck`, a subcommand that does not exist. skaffold.yaml port-forwards Postgres and Redis, which the chart never installs. `charts/kubin/charts/` holds the unpacked `api-gateway` subchart and a stale .tgz of it, and `templates/app-of-apps.yaml` points ArgoCD at `services/<name>`. With decision-1 and decision-2 there is one image, so the chart needs one Deployment, not a subchart per service.

decision-12: m-3 runs one node, so the chart runs one replica on SQLite with a PVC. Postgres and S3 (TASK-26, TASK-20) left m-3, and the public instance runs on Coolify, not from this chart.

## Where

- `charts/kubin/`: `Chart.yaml`, `Chart.lock`, `values.yaml`, `values-dev.yaml`, `templates/`, `charts/`.
- `skaffold.yaml`: image name, values, port-forwards.
- `docs/self-hosting.md` (install the chart), `docs/development.md` ("Local cluster" still says the chart is broken), `docs/architecture.md` (Deploying table).
- The binary is `/kubin-server` in a scratch image: port 8080, runs as uid 1001, data in `/data`, configured only by env vars (docs/self-hosting.md). `/healthz` answers while the process runs, `/readyz` checks the database.

## Spec

- One Deployment with one replica, a Service, a PVC for `/data` (size and storage class from values), an optional Ingress with TLS. Image `ghcr.io/3nd3r1/kubin` with the tag defaulting to the chart's `appVersion`.
- `replicaCount` above 1 fails to render with a clear message: SQLite allows one replica (decision-12).
- Extra env vars from values, and secret ones from an existing Secret named in values, never as plain values.
- Ingress: class and annotations from values. Do not default to nginx annotations: ingress-nginx's upstream maintenance ends in 2026.
- skaffold runs the chart on kind with no other services.

## Out of scope

Server code, Postgres and S3 modes, publishing the chart to a registry, the public instance (TASK-23), ArgoCD.

## Traps

- The image runs as uid 1001: set `runAsUser`, `runAsGroup` and `fsGroup` to 1001, or the PVC is not writable and the server exits with `mkdir /data/tmp: permission denied`.
- A ReadWriteOnce PVC with a RollingUpdate leaves the new pod stuck waiting for the volume. Use `strategy: Recreate`.
- Ingress controllers cap request bodies (nginx at 1 MiB by default). The docs must say to allow `SERVER_MAX_UPLOAD_BYTES` (64 MiB).
- Links: set `SERVER_PUBLIC_URL` from the ingress host, or rely on `X-Forwarded-Proto` and `Host` (docs/self-hosting.md, "Behind a proxy").
- skaffold's image name must match what the chart renders. Set `image.repository` and `image.tag` from skaffold `setValueTemplates` (`{{.IMAGE_REPO_kubin}}`, `{{.IMAGE_TAG_kubin}}`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 On a fresh kind cluster `skaffold run` brings up kubin and it becomes Ready
- [x] #2 Probes use HTTP /healthz for liveness and /readyz for readiness
- [x] #3 The rendered image name matches what skaffold builds
- [x] #4 kubin runs one replica with SQLite on a PVC that survives a pod restart; replicaCount above 1 fails to render with a clear message
- [x] #5 An optional Ingress with TLS takes its class and annotations from values
- [x] #6 The api-gateway subchart, its .tgz, Chart.lock and app-of-apps.yaml are gone: `grep -rn api-gateway charts skaffold.yaml docs` prints nothing
- [x] #7 `kubin create --server` through a port-forward to the chart Service uploads, and the link opens
- [x] #8 docs/self-hosting.md covers installing the chart, and the Local cluster section of docs/development.md is current
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 Manual: kind create cluster --name kubin, skaffold run, kubectl -n kubin-dev wait --for=condition=Ready pod -l app.kubernetes.io/name=kubin --timeout=180s, then an upload through kubectl port-forward
- [x] #2 make check passes
- [x] #3 helm lint charts/kubin passes, helm template renders with defaults, and fails with replicaCount=2
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
After TASK-1 the binary is /kubin-server (scratch image, port 8080, no healthcheck subcommand). The subchart is still named api-gateway and app-of-apps.yaml still points ArgoCD at services/<name>.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replaced the broken api-gateway umbrella chart with a single-container kubin Helm chart (one replica, Recreate strategy, uid/gid/fsGroup 1001, HTTP /healthz and /readyz probes, PVC for /data, optional Ingress with TLS, env and existingSecret, and template failure when replicaCount > 1). Updated skaffold.yaml to build with BuildKit and deploy charts/kubin to kubin-dev. Updated docs/self-hosting.md, docs/development.md, and docs/architecture.md. Verified with helm lint, helm template (defaults, ingress, replicaCount=2 failure), kind + skaffold run, snapshot upload via port-forward, pod restart persistence, and make check.
<!-- SECTION:FINAL_SUMMARY:END -->
