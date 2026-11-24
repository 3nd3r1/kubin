---
id: TASK-9
title: Bring the GitHub issue setup in line with the backlog
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:04:00Z'
updated_date: '2026-11-24T11:08:00Z'
labels:
  - chore
  - tier:1
milestone: m-3
dependencies: []
priority: low
type: chore
ordinal: 1500
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The issue templates and the apply-component-label workflow list the seven services from the old design (decision-1), and the templates point at GitHub project 3nd3r1/6, which Backlog replaced. Whether outside people can open issues is decision-9. If decision-9 is accepted or still proposed, do its recommended option: keep the bug and feature templates and remove the others. If it is rejected, do what the user wrote in it.

## Where

- `.github/ISSUE_TEMPLATE/`: `bug.yml`, `chore.yml`, `docs.yml`, `feature.yml`, `subtask.yml`.
- `.github/config.yml`: `blank_issues_enabled: false` plus contact links that only repeat the templates.
- `.github/workflows/apply-component-label.yml`: `componentMap` of old services, with mixed `component:` and `components:` prefixes.

## What to change (recommended option)

- Delete `chore.yml`, `docs.yml` and `subtask.yml`.
- In `bug.yml` and `feature.yml`: drop `projects: [3nd3r1/6]`, make the Component dropdown options exactly `CLI`, `Server`, `UI`, `Deploy`, and replace placeholders that mention commands that don't exist (like `kubin upload ./broken.json`) with `kubin create`.
- In the workflow: `componentMap` maps `CLI`, `Server`, `UI`, `Deploy` to `component:cli`, `component:server`, `component:ui`, `component:deploy`. Bump `actions/github-script@v6` to `@v7`.
- In `config.yml`: keep `blank_issues_enabled: false`, drop the contact links that point back at the templates.

## Out of scope

- `.github/workflows/ci.yml` (TASK-8).
- Creating labels on GitHub. Say in the final summary that `component:server` and `component:deploy` may need creating in the repo settings.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `grep -rnE "API Gateway|Upload Orchestrator|Metadata Service|Storage Service|Log Processor|Query Service|Analytics Service|3nd3r1/6" .github` prints nothing
- [x] #2 Every remaining template Component dropdown lists exactly CLI, Server, UI and Deploy, and the workflow maps each to a `component:` label
- [x] #3 Only the templates decision-9 keeps remain in .github/ISSUE_TEMPLATE: bug.yml and feature.yml under its recommended option
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 nix-shell -p yq-go --run 'for f in .github/ISSUE_TEMPLATE/*.yml .github/config.yml; do yq e . "$f" >/dev/null || exit 1; done' passes
- [x] #2 nix-shell -p actionlint --run "actionlint .github/workflows/apply-component-label.yml" passes
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Delete chore, docs and subtask issue templates.
2. Drop projects field and old component list from bug and feature templates; Component options CLI, Server, UI, Deploy; placeholder commands use kubin create.
3. Config keeps blank_issues_enabled false, contact links removed.
4. Workflow componentMap maps the four components to component: labels, github-script v7.
5. Verify with yq, actionlint and the grep in AC 1.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
yq parse of all templates and config passed; actionlint passed on the workflow; grep for old service names and project board returns nothing. Replaced kubin deploy --dry-run in feature template with kubin create --output since kubin deploy does not exist.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Removed the chore, docs and subtask issue templates (decision-9, accepted). Bug and feature templates now list CLI, Server, UI and Deploy and no longer point at GitHub project 3nd3r1/6. Placeholder commands that did not exist now use kubin create. config.yml keeps blank_issues_enabled false and drops the contact links. The apply-component-label workflow maps the four components to component: labels and uses github-script v7. Checks: yq parse of templates and config, actionlint on the workflow, and the grep for old service names all pass. Follow-up for the user: component:server and component:deploy labels may need creating in the repo settings.
<!-- SECTION:FINAL_SUMMARY:END -->
