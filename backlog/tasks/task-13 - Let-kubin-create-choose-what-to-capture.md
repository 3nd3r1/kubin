---
id: TASK-13
title: Let kubin create choose what to capture
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-19T09:13:00Z'
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
- [x] #1 `--namespace` (repeatable) and `--exclude-namespace` limit what is collected
- [x] #2 `--no-logs` skips logs
- [x] #3 `--context` picks a kubeconfig context
- [x] #4 kubin create refuses snapshots over a configurable size before uploading
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. kube: List takes a namespace; NewKubeClient takes a context name (--context).
2. collector.Scope{Namespaces, Exclude}: with --namespace, namespaced kinds are listed per namespace (works with namespace-only RBAC) and Namespace objects are filtered by name; cluster-scoped kinds like Nodes stay. --exclude-namespace filters after listing.
3. snapshot.Options.SkipLogs (--no-logs) skips the logs collector; the manifest keeps an empty logs.files.
4. --max-size (quantity, default 64Mi = the server default) refuses an upload over the limit, after writing the temp file and before sending it. --output is not limited.
5. Tests for each flag, docs in the README/cli help. No format change.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claude, 2026-11-19: --namespace/-n (repeatable, comma-separated) lists namespaced kinds per namespace, so a user with only namespace RBAC can snapshot; Namespace objects are filtered by name; cluster-scoped kinds (Nodes, PVs) stay. --exclude-namespace filters after listing and applies after --namespace. A forbidden list in one namespace warns and continues. --no-logs skips the logs collector; manifest logs.files is empty, no format change. --context sets the kubeconfig context. --max-size takes a Kubernetes quantity (64Mi default, matching SERVER_MAX_UPLOAD_BYTES, 0 = no limit) and refuses the upload after writing the temp file; --output files are not limited.
Verified on kind-kubin-demo: -n kubin-demo gave 16 resources (namespace, node, pod, events, configmap) and 2 logs; --exclude-namespace kube-system --no-logs gave 63 resources with no kube-system objects and 0 logs; --context nope fails with 'context "nope" does not exist'; --max-size 10Ki refused a 60 KiB snapshot before contacting the server. go test ./... and golangci-lint pass.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
kubin create can now narrow a snapshot: --namespace and --exclude-namespace limit collection (namespaced kinds are listed per chosen namespace, so namespace-only RBAC works), --no-logs leaves logs out, --context picks the kubeconfig context, and --max-size (default 64Mi, the server's default limit) refuses an oversized upload before sending it. Tests cover scope filtering, per-namespace RBAC skips, skipped logs, the size refusal and size parsing; every flag was checked against the kind cluster. README and self-hosting docs updated.
<!-- SECTION:FINAL_SUMMARY:END -->
