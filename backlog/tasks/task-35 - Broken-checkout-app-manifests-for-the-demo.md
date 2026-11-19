---
id: TASK-35
title: Broken checkout app manifests for the demo
status: To Do
assignee: []
created_date: '2026-11-19T14:44:00Z'
labels:
  - demo
  - tier:2
milestone: m-0
dependencies: []
references:
  - demo/crashing-pod.yaml
documentation:
  - docs/architecture.md
priority: high
ordinal: 35000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

The user changed the m-0 demo story on 2026-11-19: someone has k9s open on their own cluster, sees problems, and instead of copying output into a Linear or Jira ticket runs `kubin create` and pastes the link. The single crashing `payment-api` pod (`demo/crashing-pod.yaml`, TASK-30) is too simple for that story. The demo needs a small broken app with a few believable problems that k9s shows and a kubin link explains. Background: decision-3 (m-0 is a demo), decision-5 (env var values are dropped from snapshots, but ConfigMap data is kept, which is how the viewer sees `DB_HOST=db`).

## Where

New file `demo/checkout.yaml`, one multi-document manifest. Follow the style of `demo/crashing-pod.yaml`: small images with pinned tags (busybox:1.36 is already pulled by the old demo).

The app, all in namespace `checkout`:

- **web**: Deployment `web`, 2 replicas, pod label `app: web`, a container that serves HTTP and passes a readiness probe (for example busybox `httpd -f -p 8080`). Service `web` (port 80 to 8080) whose selector has a typo, for example `app: wbe`, so it selects no pods and has no endpoints.
- **api**: Deployment `api`, 1 replica, `envFrom` a ConfigMap `api-config` with `DB_HOST: db` and `DB_PORT: "5432"`. The container tries `$DB_HOST:$DB_PORT` once, writes `api: cannot connect to postgres at db:5432` to stderr (built from the env values, so the log follows the ConfigMap) and exits 1, which ends in CrashLoopBackOff.
- **postgres**: headless Service `postgres` (port 5432, selector `app: postgres`) and StatefulSet `postgres` (serviceName `postgres`, 1 replica) with a volumeClaimTemplate `data` that asks for storageClassName `fast-ssd`, which kind doesn't have. The image is never pulled, so `postgres:16-alpine` is fine for realism.
- **worker**: Job `worker`, `restartPolicy: Never`, `backoffLimit: 0`, that fails with one clear log line (it can use `api-config` too and fail on the same database name).

A comment at the top of the file lists the four planted problems in one line each.

Checked by the planner on 2026-11-19 on kind 0.33 with node image kindest/node:v1.37.0: a StatefulSet pod whose PVC asks for a missing StorageClass stays Pending with `Warning FailedScheduling  0/1 nodes are available: pod has unbound immediate PersistentVolumeClaims. not found`, and the PVC gets `Warning ProvisioningFailed  storageclass.storage.k8s.io "fast-ssd" not found`. So the postgres failure reproduces on kind.

## Out of scope

- `demo/run.sh`, `demo/crashing-pod.yaml`, README, `demo/video-script.md`: the next tasks switch them over. Leave `crashing-pod.yaml` in place so `run.sh` keeps working until then.
- CLI, server and UI code.
- Secrets or anything that looks like a real credential. The demo invites sharing links.

## Traps

- The names `checkout`, `web`, `api`, `api-config`, `postgres`, `worker` and the log line `cannot connect to postgres at db:5432` are used by the setup script, the UI tasks and the video script that follow. Don't rename them.
- kind's default StorageClass is `standard` (WaitForFirstConsumer). Don't omit `storageClassName` or use `standard`, or the claim binds and postgres starts.
- Applying a namespace and pods in one file can fail the first time because the namespace's default ServiceAccount doesn't exist yet (seen in TASK-30). Retry the apply for a few seconds.
- Don't set `ttlSecondsAfterFinished` on the Job: it would delete the failed run before `kubin create` sees it.
- Check that the command in the api and worker containers exists in the image (busybox `nc` flags vary by build). The log line has to appear whatever the probe does.
- Deployment pods have random suffixes. Refer to them as `deploy/api` in commands.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `demo/checkout.yaml` applies to a fresh kind cluster with `kubectl apply -f demo/checkout.yaml` (retried while the namespace's default ServiceAccount appears) and creates everything in namespace `checkout`
- [ ] #2 Within 2 minutes `kubectl -n checkout get deploy web` shows 2/2 ready and `kubectl -n checkout get endpointslices -l kubernetes.io/service-name=web -o jsonpath='{.items[*].endpoints}'` prints nothing
- [ ] #3 The api pod is in CrashLoopBackOff with restartCount 1 or more, and `kubectl -n checkout logs deploy/api --previous` contains `cannot connect to postgres at db:5432`; `kubectl -n checkout get configmap api-config -o yaml` shows `DB_HOST: db`
- [ ] #4 `postgres-0` stays Pending with a FailedScheduling event, and PVC `data-postgres-0` stays Pending with a ProvisioningFailed event naming `fast-ssd`
- [ ] #5 Job `worker` has condition Failed=True and `kubectl -n checkout logs job/worker` prints its error line
- [ ] #6 The file has no Secret objects or credentials, every image has a pinned tag, and a comment at the top lists the four planted problems
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 Checked on a fresh cluster (`kind create cluster --name kubin-checkout-test --wait 2m`), with the commands and their output in the implementation notes, and the cluster deleted afterwards (`kind delete cluster --name kubin-checkout-test`)
- [ ] #2 The task's commit changes only `demo/checkout.yaml` (check with `git show --stat HEAD`)
<!-- DOD:END -->
