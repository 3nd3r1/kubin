---
id: TASK-22
title: Upload limits and rate limiting
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
labels:
  - server
milestone: m-3
dependencies:
  - TASK-3
priority: medium
ordinal: 22000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A public upload endpoint will be abused without limits.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Max upload size and max uncompressed size are enforced (zip bomb protection)
- [ ] #2 Uploads are rate limited per client IP
- [ ] #3 Limits are configurable and return 413 or 429 with a message
<!-- AC:END -->
