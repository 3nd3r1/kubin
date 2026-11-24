---
id: decision-9
title: GitHub issues stay open with a bug and a feature template
date: '2026-11-24T09:20:00Z'
status: proposed
---
## Context

The GitHub issue templates come from the seven-service design (decision-1): five templates (bug, feature, chore, docs, subtask), a component list of services that no longer exist, and a link to GitHub project 3nd3r1/6. The work itself now lives in Backlog, so chore, docs and subtask issues have no use. TASK-9 cleans this up, and whether outside people should still be able to open issues at all is the user's call.

## Decision

Proposed, waiting for the user. Recommended: keep issues open to everyone with two templates, bug and feature, whose Component dropdown lists CLI, Server, UI and Deploy. Remove the chore, docs and subtask templates and the project board link. The user turns issues worth doing into Backlog tasks or drafts.

The alternative is to turn issues off and point people at discussions or email.

## Consequences

People who try kubin can report bugs where they expect to, which matters once a release and a public instance exist. The user triages them by hand; nothing syncs GitHub issues and Backlog. If this is rejected, TASK-9 removes the templates instead and the user turns issues off in the repo settings.

