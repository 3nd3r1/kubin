---
id: decision-12
title: >-
  m-3 runs on one node: Postgres and S3 wait for a real need, the public
  instance runs on Coolify
date: '2026-11-24T10:29:00Z'
status: accepted
---
## Context

decision-2 made SQLite and a local directory the defaults, with Postgres and S3 as options for running several replicas. m-3 as planned on 2026-11-24 built both options: Postgres (TASK-26), S3 (TASK-20), a compose file with Postgres and MinIO (TASK-27), a two-mode Helm chart with dev Postgres and MinIO (TASK-7), and a public instance on k3s with Postgres and R2 (TASK-23, decision-10). Every later migration would then be written twice, CI would need containers for the store tests, and MinIO no longer publishes community images.

Nobody had measured whether one container is enough. On 2026-11-24 it was measured ([docs/capacity.md](../../docs/capacity.md)): about 3.5 uploads a second of a 600-resource snapshot, 0.27 a second of a 52 MB worst case, and 1,500 to 2,200 reads a second with p95 under 45 ms. That is orders of magnitude over the expected load. The one failure was SQLITE_BUSY on concurrent uploads, a bug in how the server uses SQLite (TASK-48), not a limit of SQLite.

The user runs a Coolify server, which deploys Docker images with Traefik and Let's Encrypt.

## Decision

The user agreed to this on 2026-11-24.

- m-3 ships one node only: SQLite and a local directory. `docker run`, the compose file and the Helm chart all run one replica with a volume.
- TASK-26 (Postgres) and TASK-20 (S3) leave m-3 and stay on the board without a milestone. Keep SQL portable and keep the `store.Store` and `blob.Store` interfaces, so adding them later is one task each.
- The public instance runs as one container on the user's Coolify server: the released image at a pinned tag, a named volume at `/data`, TLS from Coolify. This replaces decision-10. The domain for the first version is `kubin.ender.fi` (the user, 2026-11-24).
- Postgres and S3 come back when one of these happens: uploads queue up behind each other in normal use, the disk can't hold a week of snapshots, or someone needs an instance that survives a node going down.

## Consequences

m-3 loses about three tasks, and TASK-7 and TASK-23 get smaller. TASK-48 has to land before the public instance. The public instance has no failover: when the Coolify server is down, so are the links, which is acceptable for a free pastebin. Backups of the volume are the user's; Litestream to S3-compatible storage is an option that doesn't change the architecture. doc-1's goal still says bigger setups add Postgres, S3 or Kubernetes, which now describes the future rather than m-3.
