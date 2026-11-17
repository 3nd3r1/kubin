---
id: TASK-5
title: Upload from kubin create and print the URL
status: To Do
assignee: []
created_date: '2026-11-17T10:04:00Z'
labels:
  - cli
milestone: m-0
dependencies:
  - TASK-2
  - TASK-3
priority: high
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
This closes the loop that is the whole product: one command, one link. Today `kubin create` drops a tarball in the current directory and the user has to do something with it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `kubin create` uploads to the server from `--server` or `KUBIN_SERVER` and prints only the URL on stdout
- [ ] #2 `kubin create --output file.tar.gz` writes the tarball and skips the upload
- [ ] #3 Progress and errors go to stderr, and an unreachable server fails with a clear message
- [ ] #4 The temporary snapshot is removed after upload
<!-- AC:END -->
