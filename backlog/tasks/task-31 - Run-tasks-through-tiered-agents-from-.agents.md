---
id: TASK-31
title: Run tasks through tiered agents from .agents/
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-19T10:26:00Z'
updated_date: '2026-11-19T11:19:00Z'
labels:
  - chore
dependencies: []
priority: high
ordinal: 31000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Every task runs on Opus 5.5, which wastes tokens on small work and carries old context between tasks (user, 2026-11-19). The user approved a workflow where Opus plans and tiers tasks, a script runs each task in a fresh headless Claude session on the model its tier locks, and Opus reviews only risky changes. Design: https://claude.ai/artifact/5UNZGZ5dQvafSGWKtkjrV2. Keep it generic in .agents/ so it can be copied to other repos; only .agents/config and the CLAUDE.md section are kubin's.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 .agents/config maps tier:1-4, planner and reviewer to full model IDs and effort, and lists the paths that always get a review
- [x] #2 .agents/run runs one task (TASK-N or --next) headless on its tier, checks Done, ticked AC and the repo checks, retries once a tier up, and stops
- [ ] #3 .agents/run --plan <milestone> runs the planner, which turns drafts into tasks and gives every task in the milestone a tier label
- [x] #4 Changes in review paths or tier 4 go through an Opus review whose findings go back to the worker once
- [x] #5 Backlog has a Blocked status, CLAUDE.md points at .agents/README.md, and the workflow is an accepted decision
- [x] #6 --dry-run shows the task, model, effort and review choice without launching anything
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. .agents/config (bash, sourced): tier and role model+effort, review paths, permission mode, budget, and a run_checks function given changed files.
2. .agents/run: TASK-N | --next | --plan <m> | --dry-run. Worker = claude -p --model --effort with prompts/worker.md; verify status Done, AC ticked, run_checks; one retry a tier up; review via prompts/reviewer.md when tier 4 or review paths touched, findings back to worker once; logs in .agents/logs (ignored).
3. prompts/planner.md carries the tier rubric and task-writing rules; drafts are the intake.
4. README, CLAUDE.md workflow section, decision-6. Test with --dry-run and stubbed claude.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Claude, 2026-11-19: added .agents/ (config, run, prompts/worker|reviewer|planner.md, README), a Blocked status and tier:1-4 labels in backlog/config.yml, the CLAUDE.md workflow lines, and decision-6. Prompts go to claude on stdin because --allowedTools takes a list. The reviewer runs with --permission-mode dontAsk and only Read,Grep,Glob, and gets the task text and diff in its prompt. --next stops with 'run --plan' when the first ready task has no tier.
Tested in a throwaway clone with a stub claude (CLAUDE_BIN): tier 2 stopping short then passing on tier 3; a review path change reviewed, findings sent back once, passed on re-review; a Blocked task exits 3 with the worker's comment; checks failing on tier 1 and tier 2 exits 1. shellcheck passes. A real 'claude -p --model claude-haiku-5-5 --effort low --permission-mode dontAsk --allowedTools Read,Grep,Glob' with a stdin prompt answered as expected. Not yet run for real: a worker or planner session. AC #3 waits on the first real --plan run.

Claude, 2026-11-19: added make board|next|task|plan|usage. make board runs backlog browser plus .agents/run --watch: a card dragged to the new Run status gets done, one at a time; a card the script refuses goes to Blocked with the reason as a comment. Sessions use --output-format json; each prints tokens, turns, seconds and API-equivalent dollars and appends to .agents/logs/usage.tsv. Changes under backlog/ no longer count as a dirty tree, since moving a card edits its file. Tested watch in the clone with the JSON stub: an untiered card went to Blocked, a tier 2 card ran, was reviewed and committed, make usage showed both sessions.
<!-- SECTION:NOTES:END -->
