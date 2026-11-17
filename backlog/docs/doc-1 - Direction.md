---
id: doc-1
title: Direction
type: other
created_date: '2026-11-17T10:02:00Z'
updated_date: '2026-11-17T10:02:00Z'
---
Owned by the user. Agents read it and don't edit it.

## Goal

Run `kubin create`, get a link, and anyone with the link can browse the cluster as it was: resources, events and logs, in a Lens-like UI. It replaces cluster dumps attached to tickets.

## Where we are (2026-11-17)

Work stopped in September 2025. The CLI collects namespaces and pods into a local tarball, with unit tests. The server is a chi skeleton where every route returns 501. The UI is a placeholder page. The Helm chart and skaffold config exist but the chart doesn't deploy a working pod. Nothing uploads, so there are no links yet.

The old plan was seven microservices with Kafka, ClickHouse and Redis. On 2026-11-17 it was replaced by one server, Postgres and a blob store (decision-1).

## Milestones, in order

1. **Snapshot to link (m-0).** `kubin create` against a kind cluster prints a URL. Opening it shows the namespaces and pods, and a pod's details. Everything runs locally.
2. **Worth sharing (m-1).** The snapshot has what a debugging session needs: workloads, services, events, nodes and pod logs. Secrets don't leak.
3. **Lens-like viewer (m-2).** Browsing a snapshot feels like Lens: kinds in a sidebar, YAML, events and logs per resource, unhealthy things first.
4. **Hosted (m-3).** CLI releases, images, CI, a deployed instance with S3 storage and expiry.

## Parked

- Auth and private snapshots: when there is a hosted instance with real users.
- Diffing two snapshots.
- Log analytics beyond search in one file (ClickHouse): when log search gets slow on real snapshots.
- Running kubin in-cluster on a schedule.
