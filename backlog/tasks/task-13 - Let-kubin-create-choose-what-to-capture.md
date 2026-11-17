---
id: TASK-13
title: Let kubin create choose what to capture
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
labels:
  - cli
milestone: m-1
dependencies:
  - TASK-10
  - TASK-11
priority: medium
ordinal: 13000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Whole-cluster capture is slow and noisy on big clusters, and people often care about one namespace. The old cli README promised `--namespace` but it was never built.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `--namespace` (repeatable) and `--exclude-namespace` limit what is collected
- [ ] #2 `--no-logs` skips logs
- [ ] #3 `--context` picks a kubeconfig context
- [ ] #4 kubin create refuses snapshots over a configurable size before uploading
<!-- AC:END -->
