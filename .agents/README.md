# .agents

Runs Backlog tasks in fresh headless Claude sessions, each on the model its tier locks. Opus plans and reviews, cheaper models do the routine work. The design is in the [tiered-model workflow doc](https://claude.ai/artifact/5UNZGZ5dQvafSGWKtkjrV2).

## Use

```sh
make plan M=m-2   # planner: turns drafts into tasks, writes them out, labels each tier:1-4
make next         # does the next ready To Do task, then stops
make task T=15    # does TASK-15, then stops
make board        # opens the Backlog board; drag a card to Run and it gets done
make usage        # tokens, turns and time per session
```

`make board` starts `backlog browser` and a watcher. The watcher does nothing until you drag a card into the Run column, then does that one task and goes back to waiting. A card it can't do moves to Blocked with the reason as a comment. Ctrl+C stops both.

Each run prints the tokens it used and adds a line to `.agents/logs/usage.tsv`. The dollar figure is what the run would cost at API prices. On a subscription it only helps compare runs.

A run exits 0 when the task is Done, committed and checked, 1 when it failed, 2 on bad usage or a dirty tree, and 3 when the task needs you. `.agents/run --help` lists the flags behind the make targets, including `--dry-run`.

## What a run does

1. Launches `claude -p` with the tier's model and effort from `config` and `prompts/worker.md`.
2. Checks that the task is Done, every acceptance criterion is ticked, nothing is left uncommitted, and `run_checks` from `config` passes for the files that changed.
3. If that fails, retries once a tier up. If it fails again, or the worker set the task to Blocked, it stops and hands over to you.
4. Sends the change to the reviewer when the tier is in `REVIEW_TIERS` or a changed file is under `REVIEW_PATHS`. The reviewer only reads. Its findings go back to the worker once, then the reviewer looks again. If it still asks for changes, the run stops.

## Tiers

| Label | Fits |
| --- | --- |
| `tier:1` | Mechanical edits with no design: docs, config, renames, bookkeeping |
| `tier:2` | One area, the change spelled out, a few files |
| `tier:3` | One feature in one part of the repo, clear acceptance criteria |
| `tier:4` | Crosses parts, changes a contract or file format, or needs design calls |

The models behind each tier are in `config`. The planner sets the tier. Workers never change it.

## Task intake

Only the planner creates tasks. A worker that finds other work writes `backlog draft create` and carries on. `--plan` turns drafts into tasks or archives them. Run it when a milestone starts and when it ends.

## Another repo

Copy `.agents/`, then change `config`: models, review paths, permission mode and `run_checks`. Add `Run` and `Blocked` statuses and the `tier:1` to `tier:4` labels to the Backlog config, and a Workflow line in CLAUDE.md pointing here. It needs `git`, `jq`, `backlog` and `claude`.

`CLAUDE_BIN` swaps the `claude` binary, which is how the script was tested with a stub.
