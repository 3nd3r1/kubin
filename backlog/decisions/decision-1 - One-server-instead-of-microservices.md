---
id: decision-1
title: One server instead of microservices
date: '2026-11-17T10:03:00Z'
status: accepted
---
## Context

The 2025 plan split the backend into seven services (API gateway, upload orchestrator, metadata, storage, log processor, query, analytics) with Kafka, ClickHouse, Redis, Postgres and S3. After two months only the gateway skeleton existed, with every route returning 501, and the work stopped. No snapshot had ever been uploaded. The load the plan was sized for (100 concurrent uploads, sub-second log analytics) is a guess with no users behind it.

## Decision

Made by the agent on 2026-11-17 during the reassessment the user asked for. Reverse it if the project's point is to practice microservices.

- One Go server. The old services become packages inside it: api, snapshot, store, blob.
- Postgres for snapshot metadata and a resource index. Raw tarballs and logs in a blob store: filesystem first, S3 when hosted.
- Upload is synchronous. No Kafka, no Redis, no ClickHouse.
- The tarball layout is the CLI to server contract and is versioned.

## Consequences

The path to a first working link is about six tasks (m-0). `services/api-gateway` and `services/shared` merge into `server/`. The chart gets one server subchart instead of seven. The GitHub issue templates and label workflow list the old services and need updating. If a part gets slow under real use, it can be split out along its package boundary then.
