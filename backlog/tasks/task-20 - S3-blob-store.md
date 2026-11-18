---
id: TASK-20
title: S3 blob store
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-18T11:02:00Z'
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
- [ ] #2 Works against AWS S3 and MinIO
- [ ] #3 Integration tests run against a MinIO container started by the test
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-14: blob.Store has OpenRange(key, offset, length) for serving logs; map it to a ranged GetObject.
<!-- SECTION:NOTES:END -->
