---
id: TASK-16
title: Resource detail page with YAML and events for every kind
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-19T12:17:00Z'
labels:
  - ui
  - tier:3
milestone: m-2
dependencies:
  - TASK-15
documentation:
  - docs/architecture.md
  - docs/snapshot-format.md
priority: high
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
## Why

Debugging means opening any resource and reading its YAML and the events about it. m-0 built this for pods only (`ui/src/pages/pod.tsx`). This task gives every kind a detail page. Links between related resources (owners, owned, node, services) are TASK-33.

## Where

- The route `/s/{id}/{resource}/{namespace}/{name}` from TASK-15, with `_cluster` as namespace for cluster-scoped kinds. `getResource` in `ui/src/lib/api.ts` already maps an empty namespace to `_cluster`.
- `ui/src/pages/resource.tsx` (new): kind, name, namespace and age at the top, then labels and annotations, events and YAML. Pods keep their own page (`ui/src/pages/pod.tsx`, with containers and logs) and switch to the shared events and YAML components.
- Rows in the kind tables (`ui/src/pages/kind.tsx`, TASK-15) link to the detail.
- Events: generalize `podEvents` in `ui/src/lib/events.ts` to any object, matching `involvedObject` kind, namespace and name, and UID when both sides have one. Move `ui/src/components/pod/events.tsx` to a shared component and keep its table and empty text.
- YAML: `stringify` from the `yaml` package with `lineWidth: 0`, as `pod.tsx` does, and a Copy button using `navigator.clipboard.writeText`.
- Tests in `ui/src/lib/events.test.ts` and a new test file for the redaction marking.

## Out of scope

- Owner, owned, node and service links (TASK-33). The log viewer (TASK-17, TASK-34).
- Server, API, CLI and the snapshot format. Redaction itself happens in the CLI (`cli/pkg/redact/`); this task only shows what was redacted.
- Any editing. The viewer is read-only.

## Traps

- Redaction (from TASK-12, also in the implementation notes): when `manifest.redactions` contains `env-values`, env entries that have neither `value` nor `valueFrom` had their value removed. Show them as redacted, for example with a YAML comment, not as an entry that silently has no value. Never insert a made-up value: the Copy button copies the YAML as shown, comments included. Rules are listed in `docs/snapshot-format.md#redaction`.
- Events about cluster-scoped objects (Nodes, PersistentVolumes) are stored in some namespace, usually `default`. For a cluster-scoped object list Events without a namespace filter and match by kind and name.
- Event `involvedObject.namespace` is empty for cluster-scoped objects.
- `getResource` sends no `group`. That is fine for the collected kinds, which have unique names; don't add group handling.
- The existing `Owner` field on the pod page stays as text; TASK-33 turns it into a link.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every row of a kind table links to `/s/{id}/{resource}/{namespace}/{name}`, which shows the object as YAML with a Copy button that copies exactly the YAML shown
- [x] #2 The detail lists the events whose `involvedObject` matches the resource (kind, namespace, name, and UID when both have one) with type, reason, message, count and last seen, newest first; `ui/src/lib/events.test.ts` covers a namespaced kind and a Node
- [x] #3 When `manifest.redactions` contains `env-values`, env entries with neither `value` nor `valueFrom` are marked as redacted in the YAML, with a test
- [x] #4 A resource missing from the snapshot shows the not-found page
- [x] #5 The pod page keeps its containers and logs and uses the shared events and YAML components
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 `nix-shell --run 'cd ui && make lint && make typecheck && make test && make build'` passes
- [x] #2 The task's commits change nothing under `server/`, `cli/` or `docs/snapshot-format.md` (check with `git log --stat`)
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
From TASK-12: when the manifest redactions list has env-values, show env entries without value or valueFrom as redacted, not as empty. Secret data values are empty strings under secret-data.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a shared ResourcePage at /s/{id}/{resource}/{ns}/{name} with labels, annotations, events and YAML (Copy copies the exact string shown). Generic kind table rows now link to it. events are matched by kind, namespace, name and UID via resourceEvents (tests for a namespaced kind and a Node). toYaml adds a comment on env entries without value/valueFrom when env-values was redacted (tests). Missing resources show the not-found page. Pod page uses the shared ResourceEvents and YamlView. Checks: lint, typecheck, test, build pass. Not exercised in a browser; Copy and routing are covered only by typecheck/build.
<!-- SECTION:FINAL_SUMMARY:END -->
