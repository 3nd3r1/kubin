---
id: decision-7
title: Tasks run as subagents from one Claude session
date: '2026-11-19T11:51:00Z'
status: accepted
---
## Context

decision-6 ran each task through `.agents/run`, a bash script that started a headless `claude -p` per task. The user found the make targets too complicated and couldn't see what a run was doing. The board's Run column needed more design than it had. Usage numbers weren't needed. Backlog.md's own guide asks for one task per context window, with the task file holding the spec. Subagents give that inside one session.

## Decision

The user chose this on 2026-11-19:

- You talk to one Claude session, the orchestrator. It reads the board, launches agents and checks their work. It doesn't write code, so Sonnet is enough to run it.
- Each task runs in a fresh subagent that sees only CLAUDE.md, its agent file and the task ID. The task's tier label picks the agent:

  | Label | Agent | Model, effort | Fits |
  | --- | --- | --- | --- |
  | `tier:1` | `worker-tier1` | Haiku, low | Mechanical edits: docs, config, renames, bookkeeping |
  | `tier:2` | `worker-tier2` | Sonnet, low | One area, the change spelled out, a few files |
  | `tier:3` | `worker-tier3` | Sonnet, medium | One feature in one part of the repo |
  | `tier:4` | `worker-tier4` | Opus, medium | Crosses parts, changes a contract or what leaves the cluster, or needs design calls |

- The agents live in `.claude/agents/`. Their model and effort are fixed there, and workers never change a tier.
- "continue" runs the next ready task in the current milestone in the background. The orchestrator then checks it: the task is Done, every criterion is ticked, the tree is clean, and the tests and lint pass for the parts it touched. If any of that fails, it retries once a tier up. Then it reports and goes on to the next task, one task at a time.
- Tier 4 workers write a plan into the task first. The user approves it before any code is written. This replaces decision-6's Opus review.
- The orchestrator stops and asks for a blocked task, a tier 4 plan, a proposed decision, or planning being due. On the user's go it launches the `planner` agent (Opus, high), which turns drafts into tasks, writes them out and tiers them. Only the planner creates tasks. Workers write drafts.
- All state lives in Backlog, so clearing the orchestrator loses nothing. Clear after a milestone or when the session is long, then say "continue".

## Consequences

`.agents/`, the make targets for agent runs, the Run status and usage logging are gone. Model choice uses aliases (`haiku`, `sonnet`, `opus`), so a tier moves when an alias moves. The session has to stay open while work runs. Parallel tasks would need worktrees and are left out for now.
