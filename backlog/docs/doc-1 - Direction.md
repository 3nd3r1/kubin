---
id: doc-1
title: Direction
type: other
created_date: '2026-11-17T10:02:00Z'
updated_date: '2026-11-17T10:12:00Z'
---
Owned by the user. Agents read it and don't edit it.

## Goal

Run `kubin create`, get a link, and anyone with the link can browse the cluster as it was: resources, events and logs, in a Lens-like UI. It replaces cluster dumps attached to tickets.

Anyone can run their own kubin easily. One container and a volume is enough. Bigger setups add Postgres, S3 or Kubernetes.

## Where we are (2026-11-17)

Work stopped in September 2025. The CLI collects namespaces and pods into a local tarball, with unit tests. The server is a chi skeleton where every route returns 501. The UI is a placeholder page. The Helm chart and skaffold config exist but the chart doesn't deploy a working pod. Nothing uploads, so there are no links yet.

The old plan was seven microservices with Kafka, ClickHouse and Redis. On 2026-11-17 it was replaced by one server (decision-1) that ships as one container with the UI built in and SQLite by default (decision-2).

## Milestones, in order

1. **Snapshot to link (m-0).** `kubin create` against a kind cluster prints a URL from a kubin started with `docker run`. Opening it shows the namespaces and pods, and a pod's details.
2. **Worth sharing (m-1).** The snapshot has what a debugging session needs: workloads, services, events, nodes and pod logs. Secrets don't leak.
3. **Lens-like viewer (m-2).** Browsing a snapshot feels like Lens: kinds in a sidebar, YAML, events and logs per resource, unhealthy things first.
4. **Deployable anywhere (m-3).** Released CLI and image. Self-hosting works with docker run, docker compose and the Helm chart. Snapshots expire. A public instance runs.

## Parked

- Auth and private snapshots: when there is a hosted instance with real users.
- Diffing two snapshots.
- Log analytics beyond search in one file (ClickHouse): when log search gets slow on real snapshots.
- A job queue: when indexing gets slow enough to hold up uploads. Start with a jobs table in the database before anything like Kafka.
- Running kubin in-cluster on a schedule.
