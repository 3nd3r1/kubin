---
id: TASK-7
title: Make the Helm chart deploy kubin with Postgres and S3
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-24T09:37:00Z'
labels:
  - deploy
  - bug
  - tier:4
milestone: m-3
dependencies:
  - TASK-25
  - TASK-26
  - TASK-20
priority: medium
type: bug
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The chart is the Kubernetes way to self-host and the base for the public instance (TASK-23), and it has never deployed a working pod. `helm template` renders the image as `kubin-api-gateway/kubin-api-gateway:dev` because the deployment template uses `image.repository` twice. The probes exec `/api-gateway healthcheck`, a subcommand that doesn't exist. skaffold.yaml port-forwards Postgres and Redis, which the chart never installs. `charts/kubin/charts/` holds the unpacked `api-gateway` subchart and a stale .tgz of it, and `templates/app-of-apps.yaml` points ArgoCD at `services/<name>`. With decision-1 and decision-2 there is one image, so the chart needs one Deployment, not a subchart per service.

## Where

- `charts/kubin/`: `Chart.yaml`, `Chart.lock`, `values.yaml`, `values-dev.yaml`, `templates/`, `charts/`.
- `skaffold.yaml`: image name, values, port-forwards.
- `docs/self-hosting.md` (install the chart), `docs/development.md` ("Local cluster" still says the chart is broken), `docs/architecture.md` (Deploying table).
- The binary is `/kubin-server` in a scratch image: port 8080, runs as uid 1001, data in `/data`, configured only by env vars (docs/self-hosting.md). `/healthz` answers while the process runs, `/readyz` checks the database and, after TASK-20, the blob store.

## Design calls for the plan

- One Deployment, a Service, an optional Ingress with TLS, an optional PVC. Image `ghcr.io/3nd3r1/kubin` with the tag defaulting to the chart's `appVersion`.
- Two modes. SQLite with a PVC: one replica. External Postgres (`SERVER_DATABASE_URL`, TASK-26) and S3 (`SERVER_S3_*`, TASK-20): several replicas. Credentials come from an existing Secret named in values, never as plain values.
- Rendering more than one replica without both Postgres and S3 fails with a clear message (`fail` in the template).
- The kind dev setup needs Postgres and MinIO. Bitnami's charts and images went behind a paywall in 2025, so don't depend on them. Recommended: plain Postgres and MinIO manifests that only the dev setup deploys (like `deploy/kind/`), outside the chart.
- Ingress: class and annotations from values. Don't default to nginx annotations: ingress-nginx's upstream maintenance ends in 2026, and the public instance may use Traefik (decision-10).

## Out of scope

Server code (TASK-20 and TASK-26 add the config), publishing the chart to a registry, the public instance (TASK-23), ArgoCD.

## Traps

- The image runs as uid 1001: set `runAsUser`, `runAsGroup` and `fsGroup` to 1001, or the PVC isn't writable and the server exits with `mkdir /data/tmp: permission denied`.
- `/data` must be writable in every mode, because uploads spool to `/data/tmp`. Use an emptyDir when there is no PVC.
- A ReadWriteOnce PVC with a RollingUpdate leaves the new pod stuck waiting for the volume. Use `strategy: Recreate` in SQLite mode.
- Ingress controllers cap request bodies (nginx at 1 MiB by default). The docs must say to allow `SERVER_MAX_UPLOAD_BYTES` (64 MiB).
- Links: set `SERVER_PUBLIC_URL` from the ingress host, or rely on `X-Forwarded-Proto` and `Host` (docs/self-hosting.md, "Behind a proxy").
- skaffold's image name must match what the chart renders. Set `image.repository` and `image.tag` from skaffold's `setValueTemplates` (`{{.IMAGE_REPO_kubin}}`, `{{.IMAGE_TAG_kubin}}`).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 On a fresh kind cluster `skaffold run` brings up kubin with Postgres and MinIO, and kubin becomes Ready
- [ ] #2 Probes use HTTP /healthz for liveness and /readyz for readiness
- [ ] #3 The rendered image name matches what skaffold builds
- [ ] #4 Values allow SQLite with a PVC for one replica, or external Postgres and S3 for several; more than one replica without both fails to render with a clear message
- [ ] #5 An optional Ingress with TLS takes its class and annotations from values
- [ ] #6 The api-gateway subchart, its .tgz, Chart.lock and app-of-apps.yaml are gone: `grep -rn api-gateway charts skaffold.yaml docs` prints nothing
- [ ] #7 `kubin create --server` through a port-forward to the chart Service uploads, and the link opens
- [ ] #8 docs/self-hosting.md covers installing the chart in both modes, and the Local cluster section of docs/development.md is current
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 helm lint charts/kubin passes, and helm template renders in both modes
- [ ] #2 Manual: kind create cluster --name kubin, skaffold run, kubectl -n kubin-dev wait --for=condition=Ready pod -l app.kubernetes.io/name=kubin --timeout=180s, then an upload through kubectl port-forward
- [ ] #3 make check passes
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
After TASK-1 the binary is /kubin-server (scratch image, port 8080, no healthcheck subcommand). The subchart is still named api-gateway and app-of-apps.yaml still points ArgoCD at services/<name>.
<!-- SECTION:NOTES:END -->
