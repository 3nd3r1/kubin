---
id: TASK-2
title: Define snapshot format v1 and write it from the CLI
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-17T10:43:00Z'
labels:
  - cli
milestone: m-0
dependencies: []
priority: high
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The tarball is the contract between the CLI and the server, and it has no spec. Today the CLI writes `<kind>/<name>.json`, so two pods with the same name in different namespaces overwrite each other. Errors from Persist are ignored in snapshot.Manager, and errors from closing the tar and gzip writers are dropped, so a truncated archive can be reported as success. There is no manifest, so the server cannot tell what cluster or CLI version produced a snapshot. The target layout is sketched in docs/architecture.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `docs/snapshot-format.md` specifies v1: manifest.json fields and the resources/ and logs/ layout
- [x] #2 The CLI writes v1, including manifest.json with format version, cluster name, server version, kube context, CLI version and per-kind counts
- [x] #3 Same-named resources in different namespaces are both kept
- [x] #4 `managedFields` is stripped from every object
- [x] #5 Any write or close error fails `kubin create` with a non-zero exit
- [x] #6 A test builds a snapshot from fake objects and checks the layout and manifest
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Write docs/snapshot-format.md: manifest.json first, resources/<group>/<version>/<Kind>/<ns|_cluster>/<name>.json with apiVersion and kind set, logs/ reserved for TASK-11.
2. collector.Resource carries group, version, kind, namespace, name and the object as an unstructured map. CoreCollector converts typed objects and sets apiVersion/kind (List drops them).
3. kube.Client takes kubernetes.Interface so client-go fake works, loads kubeconfig with default loading rules, and exposes cluster info (context, cluster name, server version).
4. Snapshot writer: stages files in a temp dir, strips metadata.managedFields, rejects unsafe names, writes manifest.json first then files in sorted order, and returns every write and close error.
5. Manager takes its client and output path from cmd, checks every error, builds the manifest with counts.
6. Tests: end to end with the fake clientset (layout, manifest, managedFields, same-named pods), writer error paths, CoreCollector.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Writer stages files in a temp dir and tars them at the end, so the manifest can come first with final counts; TASK-11 can stream logs to disk the same way. Resource objects are unstructured maps so the dynamic client in TASK-10 fits without changes. Pods are now listed once across all namespaces instead of per namespace. MockClient replaced by client-go fake clientset. The API server URL is left out of the manifest on purpose (can be an internal hostname). Deleted pkg/persister; its filename tests were replaced by DefaultFileName. pkg/config is still empty and unused, kept for TASK-5.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
kubin create now writes snapshot format v1, specified in docs/snapshot-format.md: manifest.json first (format version, created_at, CLI version, kube context, cluster name, server version, per-kind counts), then resources/<group>/<version>/<Kind>/<ns|_cluster>/<name>.json sorted by path. Fixes the old bugs: same-named pods in different namespaces overwrote each other, Persist errors were ignored, and tar/gzip close errors were dropped. managedFields is stripped and apiVersion/kind are set on every object. Unsafe or duplicate names are rejected. The CLI no longer overwrites an existing file, removes a partial file on failure, prints errors once, and has --version. Verified: go vet, golangci-lint 0 issues, go test -race (end-to-end layout and manifest test with the fake clientset, unsafe names, duplicates, write and close errors; the close-error test was checked to fail with the fix reverted). Ran against a throwaway kind cluster (v1.37.0): 16 resources plus manifest, correct cluster info, no managedFields. Against an unreachable server it exits 1 and leaves no file.
<!-- SECTION:FINAL_SUMMARY:END -->
