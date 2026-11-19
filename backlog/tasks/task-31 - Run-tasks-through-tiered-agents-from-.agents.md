---
id: TASK-31
title: Run tasks as subagents from one Claude session
status: Done
assignee:
  - '@claude'
created_date: '2026-11-19T10:26:00Z'
updated_date: '2026-11-19T11:51:00Z'
labels:
  - chore
dependencies: []
priority: high
ordinal: 31000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
decision-6's .agents/run script, make targets and Run column were too complicated, and runs gave no view of what was happening (user, 2026-11-19). decision-7 replaces it: one Claude session orchestrates, and each task runs in a fresh subagent whose model and effort its tier label picks.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 .claude/agents has worker-tier1 to worker-tier4 and planner, each with a fixed model and effort
- [x] #2 .agents/, the agent make targets and the Run status are gone
- [x] #3 CLAUDE.md tells the orchestrator how to continue, check, retry and stop, and decision-7 is accepted with decision-6 superseded
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

Claude, 2026-11-19: replaced the script with subagents (decision-7). The m-2 plan from the script's planner run (commit 139d6d5) stays. The reviewer is dropped. Tier 4 plan approval by the user takes its place, and the planner tiers changes to collectors or redaction as tier 4.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Replaced .agents/run with subagents run from one Claude session (decision-7, superseding decision-6). .claude/agents has worker-tier1 to worker-tier4 (haiku low, sonnet low, sonnet medium, opus medium) and planner (opus high). Tier 4 workers plan first and wait for the user's approval. CLAUDE.md says how the orchestrator continues, checks, retries and stops. Removed .agents/, the board/next/task/plan/usage make targets and the Run status. Verified that a headless session lists all five agents. Not yet tried on a real task.
<!-- SECTION:FINAL_SUMMARY:END -->
