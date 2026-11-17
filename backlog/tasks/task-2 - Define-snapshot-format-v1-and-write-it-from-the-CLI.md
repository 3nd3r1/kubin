---
id: TASK-2
title: Define snapshot format v1 and write it from the CLI
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
labels:
  - cli
milestone: m-0
dependencies: []
priority: high
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The tarball is the contract between the CLI and the server, and it has no spec. Today the CLI writes `<kind>/<name>.json`, so two pods with the same name in different namespaces overwrite each other. Errors from Persist are ignored in snapshot.Manager, and errors from closing the tar and gzip writers are dropped, so a truncated archive can be reported as success. There is no manifest, so the server cannot tell what cluster or CLI version produced a snapshot. The target layout is sketched in docs/architecture.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `docs/snapshot-format.md` specifies v1: manifest.json fields and the resources/ and logs/ layout
- [ ] #2 The CLI writes v1, including manifest.json with format version, cluster name, server version, kube context, CLI version and per-kind counts
- [ ] #3 Same-named resources in different namespaces are both kept
- [ ] #4 `managedFields` is stripped from every object
- [ ] #5 Any write or close error fails `kubin create` with a non-zero exit
- [ ] #6 A test builds a snapshot from fake objects and checks the layout and manifest
<!-- AC:END -->
