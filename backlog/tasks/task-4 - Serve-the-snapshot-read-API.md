---
id: TASK-4
title: Serve the snapshot read API
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T14:02:00Z'
labels:
  - server
milestone: m-0
dependencies:
  - TASK-3
priority: high
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The UI needs to read a snapshot back. The endpoints are listed in docs/architecture.md under API.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `GET /api/v1/snapshots/{id}` returns the manifest and counts
- [x] #2 `GET /api/v1/snapshots/{id}/resources` lists resources, filterable by kind and namespace
- [x] #3 `GET /api/v1/snapshots/{id}/resources/{kind}/{ns}/{name}` returns one object
- [x] #4 Unknown IDs return 404
- [x] #5 Handlers are tested against the fixture from TASK-3
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. store: GetSnapshot, CountResources (GROUP BY over the index, since manifest counts are only a hint), ListResources with optional group/kind/namespace filter, GetResource; ErrNotFound for misses.
2. api: GET /snapshots/{id} returns id, uploaded_at, the manifest as uploaded and actual counts per kind. GET /resources?kind=&namespace=&group= returns items with group, version, kind, namespace, name and object; namespace=_cluster selects cluster-scoped. GET /resources/{kind}/{ns}/{name} returns the object itself, _cluster for cluster-scoped, optional group= to pick between kinds that share a name.
3. 404 JSON for unknown snapshot IDs (also on the list and object routes) and unknown objects.
4. Tests upload snapshottest.Fixture through the router and read it back.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Counts come from the index (GROUP BY), not the manifest, which the format spec calls a hint. List returns full objects with no paging; fine for m-0 snapshot sizes, revisit in TASK-15 if tables get large. Optional group= on list and get picks between kinds that share a name (core Event vs events.k8s.io Event); without it get takes the first by group, version. Verified: go test -race ./..., golangci-lint clean.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added the read API: GET /api/v1/snapshots/{id} (id, uploaded_at, manifest as uploaded, indexed counts per kind), GET .../resources with kind, namespace (_cluster for cluster-scoped) and group filters, GET .../resources/{kind}/{ns}/{name} returning the object itself. Unknown snapshots and resources return JSON 404s on every route. Store gained GetSnapshot, CountResources, ListResources and GetResource. Handlers are tested by uploading snapshottest.Fixture through the router onto a temp SQLite file and reading it back.
<!-- SECTION:FINAL_SUMMARY:END -->
