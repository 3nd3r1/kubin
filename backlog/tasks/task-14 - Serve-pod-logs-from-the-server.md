---
id: TASK-14
title: Serve pod logs from the server
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-18T11:02:00Z'
labels:
  - server
milestone: m-0
dependencies:
  - TASK-4
  - TASK-11
priority: high
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Logs stay in the blob store per decision-1, so the UI needs an endpoint to stream them out of the stored tarball or an extracted copy.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `GET /api/v1/snapshots/{id}/logs/{ns}/{pod}/{container}` streams the log
- [x] #2 A `previous=true` parameter returns the previous container log when present
- [x] #3 A `q` parameter returns only matching lines
- [x] #4 Large logs stream without loading the file into memory
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Upload: copy each log entry from the tar stream into one spooled logs file, recording namespace, pod, container, previous, offset and size. Put it as snapshots/<id>.logs.txt next to the tarball, before commit; delete both if anything later fails.
2. Store: migration 0002 adds snapshots.logs_blob_key and a logs table (byte_offset, byte_size; portable SQL). SnapshotTx.AddLog, Store.GetLog.
3. Blob: OpenRange(key, offset, length). FS uses a section reader; S3 (TASK-20) maps it to a Range GET.
4. API: GET /api/v1/snapshots/{id}/logs/{ns}/{pod}/{container}, previous=true, q= case-insensitive substring filter, line by line with a 1 MiB line cap. text/plain with nosniff. 404 for unknown snapshot or log, 400 for a bad previous.
5. Tests on temp SQLite and FS blobs: upload indexes logs, store GetLog, FS OpenRange, API plain/previous/q/errors, and allocation under a 32 MiB log stays far below its size. Docs: architecture API and storage.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Logs are copied out of the tarball at upload into one blob per snapshot (snapshots/<id>.logs.txt) with a logs table recording offset and size (migration 0002, portable SQL: byte_offset/byte_size, previous as 0/1). Serving is a ranged blob read, so S3 later costs one Range GET per log and deletion is two known keys. Snapshots stored before this change have no logs rows and answer 404 for logs.
q is a case-insensitive (ASCII) substring match, line by line with reused buffers and a 1 MiB line cap. Responses are text/plain with nosniff.
Verified: store, blob, upload and API tests on temp SQLite and FS blobs (plain, previous, q, 404s, 400 for bad previous, bad log names rejected, rejected uploads leave no log rows or blobs); a 32 MiB log is served plain and with q while allocating under 1 MiB. On kind, through a server from this tree on an old-schema data dir: migration applied, kubin create uploaded, the crashing pod log streamed with Content-Length and nosniff, q=fatal returned the one matching line, previous=true 404 because the kubelet had already removed that container, the old snapshot 404. Lint clean.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
GET /api/v1/snapshots/{id}/logs/{ns}/{pod}/{container} streams a container log as text/plain, with previous=true for the run before a restart and q= for a case-insensitive line filter. At upload the server copies every log out of the tarball into one logs blob per snapshot and indexes offset and size in a new logs table, so a log is served by one ranged blob read without unpacking anything; blob.Store gained OpenRange. Verified with tests (including a 32 MiB log under 1 MiB allocated) and on kind. Notes added to TASK-20 (OpenRange on S3) and TASK-21 (delete both blobs). Snapshots uploaded before this change do not serve logs.
<!-- SECTION:FINAL_SUMMARY:END -->
