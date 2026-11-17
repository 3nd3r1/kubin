---
id: decision-2
title: 'Self-hosting is one container: embedded UI, SQLite by default'
date: '2026-11-17T10:11:00Z'
status: accepted
---
## Context

The user said on 2026-11-17 that easy deployment on your own server is a main goal. decision-1 already cut the backend to one server, but that still meant a server container, a UI container, Postgres and a blob store, which is four things for a self-hoster to run. The UI was still the create-next-app placeholder.

## Decision

The user agreed to this on 2026-11-17.

- The UI is a Vite React app (Tailwind, shadcn/ui) built into the server binary with go:embed. Next.js is dropped, because its static export handles `/s/{id}` routes badly.
- SQLite (pure Go, no CGO) and a local directory are the defaults. Postgres and S3 are optional, chosen by config.
- One image serves three setups: `docker run` with a volume, docker compose with Postgres and MinIO, and the Helm chart with replicas.

## Consequences

m-0 ends with `docker run` working, not skaffold. The Helm chart moves to m-3 (TASK-7). SQL has to stay portable between SQLite and Postgres, and store tests run against both once TASK-26 lands. One SQLite file limits a deployment to one replica, which is fine for a team; more replicas need Postgres and S3.
