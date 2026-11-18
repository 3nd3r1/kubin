---
id: TASK-12
title: Keep secrets out of snapshots
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:05:00Z'
updated_date: '2026-11-18T08:50:00Z'
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
- [x] #1 Secret objects are stored with keys but without data or stringData values
- [x] #2 Env values are handled per the recorded decision, with a flag to opt out
- [x] #3 The manifest records which redaction rules ran
- [x] #4 Tests show a Secret value and a matching env value never reach the tarball
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. cli/pkg/redact: one Redactor applied by the Manager to every collected object, so collectors added later (Secrets come with TASK-10) are covered without opting in.
2. Rules: secret-data (core Secret data and stringData keep keys, values emptied); env-values (decision-5: drop value from env entries of any containers, initContainers or ephemeralContainers list anywhere in the object, so pods, workload templates and CRDs are covered; names and valueFrom stay); last-applied-configuration (parse the kubectl annotation, apply the same rules, write it back; drop it if it does not parse).
3. kubin create --keep-env-values turns off env-values.
4. Manifest redactions lists the rules that ran. Format v3, so a reader can tell a snapshot without the field from one where nothing was redacted. Server reads 1 to 3.
5. Tests: unit tests per rule, and a Manager test with a Secret and a Deployment whose secret strings (plain and base64) appear nowhere in the tarball. Docs: snapshot-format, architecture security section.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Found while building TASK-6: objects applied with kubectl apply carry the kubectl.kubernetes.io/last-applied-configuration annotation, which for a Secret holds the full data in plain JSON. Removing data and stringData is not enough; redaction must also drop or scrub that annotation (on Secrets at least, and it can echo env values on any kind).

Snapshot format v3: manifest redactions lists the rules that ran, so a reader can tell a v2 snapshot (rules unknown) from one where they ran. Server reads 1 to 3.
Rules live in cli/pkg/redact and run in snapshot.Manager on every object from every collector. Secrets are not collected from clusters yet (TASK-10), so AC1 is shown by unit tests and the Manager test feeding a Secret through a static collector.
Verified: unit tests per rule (Secret keys kept, env value dropped in containers, init and ephemeral containers under any path, valueFrom kept, last-applied scrubbed with numbers unchanged, unreadable last-applied dropped); TestCreateKeepsSecretsOut finds no Secret value (plain or base64) or env value anywhere in the tarball, and fails when the redactor call is removed; --keep-env-values keeps env values only. On kind, a pod applied with kubectl and an env value: default snapshot has the value in no file, including last-applied; --keep-env-values has it. Lint clean, server tests pass.
Not covered, documented: container command/args, ConfigMaps, other annotations, logs.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Snapshots now keep credentials out. A Redactor in the CLI runs on every collected object: Secret data/stringData keep keys with empty values, env var values are dropped everywhere a container list appears (decision-5; --keep-env-values opts out), and the last-applied-configuration annotation is scrubbed with the same rules or dropped if unreadable. The manifest records the rules that ran (format v3; the server reads 1 to 3). Docs: snapshot-format.md has a Redaction section with the gaps (args, ConfigMaps, other annotations, logs), architecture.md security section updated. Verified with unit and Manager tests that fail without redaction, and on kind. Notes added to TASK-10 (Secrets and ConfigMaps) and TASK-16 (show redacted env).
<!-- SECTION:FINAL_SUMMARY:END -->
