---
id: TASK-12
title: Keep secrets out of snapshots
status: To Do
assignee: []
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-17T14:08:00Z'
labels:
  - cli
milestone: m-0
dependencies:
  - TASK-2
priority: high
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A snapshot is a public link to a cluster dump, so a leaked credential is the worst bug kubin can have. Secret objects carry data directly. Pod and workload specs can carry credentials in env values, and ConfigMaps sometimes do too. How aggressive the default redaction is needs one call from the user: drop all env values, or keep them and match names like PASSWORD and TOKEN. Record the answer as a decision.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Secret objects are stored with keys but without data or stringData values
- [ ] #2 Env values are handled per the recorded decision, with a flag to opt out
- [ ] #3 The manifest records which redaction rules ran
- [ ] #4 Tests show a Secret value and a matching env value never reach the tarball
<!-- AC:END -->
