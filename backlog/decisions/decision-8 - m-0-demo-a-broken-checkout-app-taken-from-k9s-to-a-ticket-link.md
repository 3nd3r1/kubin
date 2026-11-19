---
id: decision-8
title: 'm-0 demo: a broken checkout app taken from k9s to a ticket link'
date: '2026-11-19T14:46:00Z'
status: proposed
---
## Context

On 2026-11-19 the user changed the m-0 demo story, as relayed by the orchestrator to the planner. decision-3 set the story as one crashing pod (`payment-api`) debugged from a link. The user found that too simple. The m-0 milestone is still titled "Demo: crashing pod to a link", and decision-3 still describes the old story.

## Decision

Recommended, pending the user's confirmation: the m-0 demo is a small broken app, "checkout", in one namespace on kind. It has a `web` Deployment whose Service selector has a typo, an `api` Deployment crash-looping because its ConfigMap says `DB_HOST=db` while the Service is `postgres`, a `postgres` StatefulSet Pending on a missing StorageClass, and a failed `worker` Job. The story: k9s on the user's machine shows the problems. Instead of copying output into a Linear or Jira ticket, the user runs `kubin create`, clicks through the link (overview, pod to ReplicaSet to Deployment to ConfigMap, the Service with no pods) and pastes the link into the ticket. The narrated video is about 75 seconds and shows at most three problems.

This replaces the story part of decision-3. The rest of decision-3 still holds: m-0 is a demo, and secrets stay out of the snapshot.

## Consequences

TASK-35 to TASK-41 in m-0 build this: manifests, a setup script that stops before `kubin create`, k9s in `nix-shell`, ConfigMap links and "selects no pods" for Services in the UI, the video script, then the GIF and README. TASK-38 and TASK-39 are viewer features that would otherwise sit in m-2. They are planned in m-0 because the demo needs them. TASK-30 stays with the user for the narrated video. If accepted, the user may want to rename m-0 (for example "Demo: broken app to a ticket link") and mark decision-3 as partly superseded. Both are the user's calls. None of the tasks waits on this decision.
