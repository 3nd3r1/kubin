---
id: decision-10
title: Where the public kubin instance runs and under which domain
date: '2026-11-24T09:20:00Z'
status: >-
  rejected: the public instance runs as one container on the user's Coolify
  server, see decision-12
---
## Context

doc-1 ends m-3 with "a public instance runs": the goal is a link anyone can open. TASK-23 deploys it from the Helm chart (TASK-7). Where it runs, who pays for it, the domain and how deploys happen are the user's calls, and TASK-23 can't be planned without them. Everything before TASK-23 in m-3 can go ahead without this.

## Decision

Proposed, waiting for the user. Recommended:

- One small VM (about 5 EUR a month, for example Hetzner CX22) running k3s, in an account the user owns. k3s ships Traefik as ingress, which avoids ingress-nginx, whose upstream maintenance ends in 2026.
- The chart in its external Postgres and S3 mode: Postgres in the cluster or the provider's managed one, blobs in S3-compatible storage with free egress such as Cloudflare R2. That runs the same setup self-hosters with several replicas use.
- TLS from cert-manager and Let's Encrypt, on a subdomain of a domain the user owns, like `kubin.<domain>`. The user fills in the domain when accepting.
- Deploys by `helm upgrade --install` from GitHub Actions on release tags, with the kubeconfig and credentials in GitHub secrets. No ArgoCD: one less system to run for one instance.
- Snapshots on the public instance live 7 days by default and at most 30 (TASK-21), with the default upload and rate limits (TASK-22).

## Consequences

A running cost and an account, DNS and secrets the user holds; TASK-23's worker will list the steps only the user can do. A public instance stores other people's cluster data, so it needs a short privacy note and an abuse contact on the site. One VM means downtime during VM maintenance, which is acceptable for a free pastebin. decision-11 (the CLI default server) depends on this.

