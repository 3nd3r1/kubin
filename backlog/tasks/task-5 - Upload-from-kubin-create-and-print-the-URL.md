---
id: TASK-5
title: Upload from kubin create and print the URL
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T14:23:00Z'
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
- [x] #1 `kubin create` uploads to the server from `--server` or `KUBIN_SERVER` and prints only the URL on stdout
- [x] #2 `kubin create --output file.tar.gz` writes the tarball and skips the upload
- [x] #3 Progress and errors go to stderr, and an unreachable server fails with a clear message
- [x] #4 The temporary snapshot is removed after upload
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. New cli/pkg/upload: POST the tarball to <server>/api/v1/snapshots, return {id,url}; unreachable server and non-2xx answers become clear errors carrying the server message.
2. kubin create gets --server (default from KUBIN_SERVER) and --output/-o. With --output it writes the file and does not upload. Otherwise it writes to a temp file, uploads it, removes it, and prints only the URL on stdout. With neither, it fails and says how to set a server.
3. Move the command body into a function that takes the kube client and writers so it can be tested with the fake clientset and an httptest server.
4. Update docs/development.md for the new flags.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verified on a kind cluster (v1.37) against the built server: stdout is only the URL, the snapshot reads back through the read API with all 9 pods, no temp files left. Without --server/KUBIN_SERVER and without --output the command fails and says how to fix it (the old default of writing kubin-snapshot-<time>.tar.gz to the working directory is gone, along with snapshot.DefaultFileName). --output stages next to the target and renames, so a failure never leaves a partial file. Progress on stderr is still zap JSON, which will look bad in the demo recording; noted on TASK-30.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
kubin create now uploads to --server or KUBIN_SERVER and prints only the link on stdout. --output/-o writes the tarball instead and skips the upload. New cli/pkg/upload posts the tarball and turns an unreachable server, a server error message, or a non-kubin answer into clear errors. The temporary snapshot is removed after upload, also on failure. The command body is runCreate, tested with the fake clientset and an httptest server. docs/development.md shows the new usage.
<!-- SECTION:FINAL_SUMMARY:END -->
