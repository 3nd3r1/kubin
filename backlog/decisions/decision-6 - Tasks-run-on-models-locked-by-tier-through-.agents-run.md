---
id: decision-6
title: Tasks run on models locked by tier, through .agents/run
date: '2026-11-19T10:30:00Z'
status: superseded by decision-7
---
## Context

Every task ran on Opus 5.5 in long sessions, which spent tokens on routine work and carried old context from task to task. Choosing a model per task by hand didn't happen. Tasks were also meant to be created as work went on, but none were: CLAUDE.md let agents create tasks while Backlog's own guide told them to ask first, and they followed the stricter rule. Design: https://claude.ai/artifact/5UNZGZ5dQvafSGWKtkjrV2

## Decision

The user chose this on 2026-11-19:

- Opus 5.5 plans. It writes the tasks and gives each a tier label, `tier:1` to `tier:4`.
- Each tier is locked to one model and effort in `.agents/config`. Workers never choose.
- `.agents/run` runs one task in a fresh headless session on its tier, checks the result, retries once a tier up, and stops.
- Opus reviews tier 4 work and any change to the paths in `REVIEW_PATHS`. Other work gets the checks only.
- Only the planner creates tasks. Workers write drafts for work they find, and the planner turns them into tasks at the start and end of each milestone.
- `.agents/` stays generic so it can be copied to other repos.

## Consequences

Backlog has a Blocked status and the tier labels. Tasks without a tier can't be run until `.agents/run --plan` labels them. The tiers are a first guess and get adjusted after a few tasks. Token use on m-2 is the first comparison against earlier milestones.
