---
id: TASK-11
title: Collect pod logs with a size cap
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-18T08:34:00Z'
labels:
  - cli
milestone: m-0
dependencies:
  - TASK-2
priority: high
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Logs are usually the reason someone shares a cluster. `kube.Client.GetPodLogs` exists but nothing calls it, and it reads the full log of one container only. A large cluster can produce gigabytes of logs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Current logs for every container, including init containers, land in logs/<ns>/<pod>/<container>.log
- [x] #2 Previous logs are collected for containers that restarted
- [x] #3 Each container log is capped by a tail limit with a flag to change it, and truncation is recorded in the manifest
- [x] #4 Logs are fetched in parallel with a bounded worker count
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. kube.Client.GetPodLogs takes corev1.PodLogOptions and returns a stream.
2. collector.CollectLogs: for each pod container and init container with a started state, fetch the current log; for restartCount > 0 also the previous log. Fixed worker pool (8). Requests tail+1 lines and keeps the last --log-tail lines within a 1 MiB cap per file, cutting at a line start; marks truncated. Forbidden warns once and skips; other per-container errors warn and skip. Results go to the writer one at a time.
3. Writer.AddLog writes logs/<ns>/<pod>/<container>.log and <container>.previous.log.
4. Format v2: manifest gets logs {tail_lines, max_bytes, files[{namespace,pod,container,previous,truncated}]}. Server accepts 1 and 2. Update docs/snapshot-format.md.
5. kubin create --log-tail (default 1000).
6. Tests: tail trimming, which containers get current/previous logs, layout and manifest; server accepts v2 with a previous log. Try it on kind with a crashing pod.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Format bumped to v2 (the rule: layout changes bump the version): <container>.previous.log and a manifest logs section {tail_lines, max_bytes, files[{namespace,pod,container,previous,truncated}]}. Server reads v1 and v2.
Found on kind: the kubelet answers 200 with "unable to retrieve container logs for containerd://..." when the container was garbage collected. The collector treats that body as no log. The kubelet keeps one dead container per name, so a crash-looping container usually has its crash in <container>.log and no previous log; a container that restarted and is running has both.
client-go defaulted to 5 QPS and throttled the log fetches; the CLI now uses QPS 50, burst 100.
Verified: unit tests with -race (tail trimming, truncation, which containers get logs, worker bound, forbidden warns once, emit errors stop the pool, layout and manifest); server tests accept v2 with a previous log and still read v1; lint clean in both. On kind: a crash-looping pod with an init container gave init.log and app.log ending in the panic, cut to --log-tail 500 and marked truncated; a pod that crashed once and recovered gave app.log and app.previous.log with the first run panic; upload to a server from this tree accepted the v2 snapshot.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
kubin create now collects pod logs. LogsCollector fetches the current log of every container and init container that has run, and the previous log of containers that restarted, with 8 workers. Each log keeps the last --log-tail lines (default 1000) within 1 MiB, cut at a line start, and truncation is recorded in the manifest. A log that cannot be fetched is skipped with a warning; forbidden warns once. Snapshot format is v2 (previous logs and the manifest logs section); the server reads v1 and v2, and docs/snapshot-format.md and architecture.md are updated. Verified with unit tests and on kind. Follow-ups: TASK-29 should read manifest logs.files to choose which log to show (note added there); logs are not redacted, which TASK-12/TASK-13 should keep in mind.
<!-- SECTION:FINAL_SUMMARY:END -->
