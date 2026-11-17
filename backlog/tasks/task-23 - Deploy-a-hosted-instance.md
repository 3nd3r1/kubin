---
id: TASK-23
title: Deploy a hosted instance
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-17T10:11:00Z'
labels:
  - deploy
milestone: m-3
dependencies:
  - TASK-7
  - TASK-19
  - TASK-21
  - TASK-22
priority: medium
ordinal: 23000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The goal is a link anyone can open, which needs a public instance. Where it runs and under which domain is the user call, and should be recorded as a decision.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The instance runs from the chart with TLS on a real domain
- [ ] #2 The CLI default server points at it
- [ ] #3 Deploys happen from CI, by ArgoCD or helm upgrade, per the decision
<!-- AC:END -->
