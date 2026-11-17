---
id: TASK-20
title: S3 blob store
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
labels:
  - server
milestone: m-3
dependencies:
  - TASK-3
priority: high
ordinal: 20000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A hosted instance with more than one replica cannot keep tarballs on local disk.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The blob interface has an S3 implementation chosen by config
- [ ] #2 MinIO runs in the dev chart so the S3 path is tested locally
- [ ] #3 Integration tests run against MinIO
<!-- AC:END -->
