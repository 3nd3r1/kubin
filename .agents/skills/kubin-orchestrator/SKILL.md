---
name: kubin-orchestrator
description: >-
  Run Kubin's tiered Backlog orchestrator loop in Claude Code or Antigravity (agy).
  Use when the user says "continue", asks to run or retry a Backlog task,
  approves a tier 4 plan, or asks to plan a milestone.
---

# Kubin Orchestrator

In orchestrator mode, you read the Backlog board, launch workers or the planner, check their work, and never write code yourself (decision-7). Start `claude` or `agy` inside `nix-shell`. A mid-tier model (Sonnet or Neon Medium) is enough for the orchestrator, since each worker sets its own model.

The worker and planner prompt source of truth lives in `.claude/agents/`.

## Orchestrator Loop

- Create, edit, tier, close and reassign single tasks yourself with the `backlog` CLI (read the matching guide first).
- Launch `planner` only when the user asks, or when it is obviously the job: planning a whole milestone, turning several drafts into tasks, or splitting a big task into many.
- Workers record outside work with `backlog draft create`, and set a task to Blocked with a comment when they need a call.
- On **"continue"**:
  1. Take the first To Do task in the current milestone that has a tier (`tier:1`..`tier:4`) and whose dependencies are Done.
  2. Launch its worker in the background with the task ID, one task at a time.
  3. After it finishes, verify:
     - the task is Done with every acceptance criterion ticked
     - `git status` is clean
     - checks pass for the parts it touched:
       - `cd cli && go test ./... && golangci-lint run ./...`
       - `cd server && make test && make lint`
       - `cd ui && make lint && make typecheck && make test`
  4. If verification fails, retry once one tier up. Report in a few lines and go on to the next task.
- **Tier 4 plan gate**: a tier 4 worker writes a plan into the task first and stops. Show the plan to the user, and once they approve it, launch the worker again with `"plan approved"`.
- **Stop and ask the user** when a task is Blocked, a tier 4 plan is waiting, a decision is proposed, or planning is due (tasks without a tier, drafts waiting, or the milestone done). On their go, launch `planner` with the milestone and show its report.
- All state lives in Backlog, so `/clear` loses nothing. Clear after a milestone or when the session gets long, then say `"continue"`.

## Launching Workers

### Claude Code

Launch `.claude/agents/worker-tier<N>` or `.claude/agents/planner` directly in the background with the task ID (plus `"plan approved"` for tier 4 execution, or the milestone ID like `m-2` for `planner`). Each file fixes its own model and effort (`haiku`/`sonnet`/`opus`, `low`/`medium`/`high`).

### Antigravity (`agy`)

`agy` uses `swarm` so each tier runs on an exact model and effort:

| Label / Role | Prompt file | `swarm` `model` | Notes |
|---|---|---|---|
| `tier:1` | `.claude/agents/worker-tier1.md` | `gemini-3.8-flash-low` | Mechanical edits |
| `tier:2` | `.claude/agents/worker-tier2.md` | `gemini-4-neon-low` | Scoped change, few files |
| `tier:3` | `.claude/agents/worker-tier3.md` | `gemini-4-neon-medium` | Single-part feature |
| `tier:4` | `.claude/agents/worker-tier4.md` | `gemini-4-neon-high` | Plans first; writes code only after `"plan approved"` |
| `planner` | `.claude/agents/planner.md` | `gemini-4-neon-high` | Plans one milestone; never writes code |

Steps for each task (or planner run):

1. Read the markdown body (everything after the YAML frontmatter) of the matching `.claude/agents/<name>.md` file.
2. Spawn one worker with `swarm`:
   - `action`: `"add"`
   - `name`: short lowercase name, e.g. `task-42` or `planner-m2`
   - `model`: from the table above
   - `workspace`: `"inherit"`
   - `brief`: the prompt body from step 1, followed by the assignment:
     - For a worker: `Task: <ID>. Use assignee @agy. Post your final report on the board when done.` (Add `Plan approved.` on the second run of a tier 4 task.)
     - For the planner: `Milestone: <m-X>. Post your final report on the board when done.`
3. Wait for the worker's report on the board (do not poll).
4. Run `swarm(action="kill", name="<name>")` once the worker is done so the next task starts fresh (commits are already in your working copy because `workspace="inherit"`).
5. Run the verification checks above. If retrying one tier up:
   - `tier:1` -> `worker-tier2.md` on `gemini-4-neon-low`
   - `tier:2` -> `worker-tier3.md` on `gemini-4-neon-medium`
   - `tier:3` -> `worker-tier4.md` on `gemini-4-neon-high` (plan-first; present the plan to the user before coding)
   - `tier:4` -> stop and report to the user
