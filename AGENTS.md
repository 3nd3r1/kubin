# Kubin

Pastebin for Kubernetes clusters: `kubin create` packs the cluster into a tarball, uploads it, and prints a link to a Lens-like web view. Self-hosting is one container. Read `docs/architecture.md` and the direction (`backlog doc view doc-1`) first.

## Layout

- `cli/`: the `kubin` command. Go, cobra, client-go.
- `server/`: the server. Go, chi, log/slog. `make test`, `make lint`, `make run`.
- `ui/`: Vite React app, Tailwind, shadcn/ui. `make build` writes it into `server/internal/web/dist`, which the server embeds. `make dev`, `make lint`.
- `charts/kubin/`: Helm chart. `skaffold.yaml` runs it on kind. Broken until TASK-7.
- `backlog/`: the board, managed with the `backlog` CLI.

## Rules

- One server, one image (decision-1, decision-2). Don't add services, queues or caches without a decision.
- `docker run` with a volume must keep working with no other services. SQLite and a local directory are the defaults, Postgres and S3 are options, so SQL stays portable.
- The snapshot tarball layout is the CLI to server contract. Changing it bumps the format version and updates `docs/snapshot-format.md`.
- Nothing secret leaves the cluster. A new collector says how it handles credentials.
- Tests use fakes for the Kubernetes API and real databases for the server: a temp SQLite file, Postgres once TASK-26 lands.

## Workflow

- Tools come from `nix-shell`, which also puts the pinned `backlog` CLI on the path.
- The user owns the direction (doc-1): the goal, the milestone order, pivots. Only the user edits doc-1. Bring them only direction calls, never implementation choices (decision-4):
  - If it blocks the work at hand, ask in chat as one question with a recommendation, and record the answer with `backlog decision create -s accepted`.
  - If it can wait, record it as a `proposed` decision (context, recommended choice, consequences) and keep working on something it doesn't block. The user sets it to `accepted` or `rejected: <why>`.
  - At the start of a session, check `backlog decision list --plain` and act on answered proposals.
- Tasks run as subagents from one Claude session (decision-7). That session is the orchestrator: it reads the board, launches agents and checks their work, and never writes code itself. Start `claude` inside `nix-shell`. Sonnet is enough for the orchestrator, since the agents set their own model.
  - Each task has a `tier:1` to `tier:4` label and runs on `.claude/agents/worker-tier<N>`, which fixes the model and effort. The orchestrator creates, edits, tiers, closes and reassigns single tasks itself with the `backlog` CLI (read the matching guide first). Launch `planner` only when the user asks, or when it is obviously the job: planning a whole milestone, turning several drafts into tasks, or splitting a big task into many. Workers put work they find outside their task in `backlog draft create`, and set the task to Blocked with a comment when they need a call.
  - When the user says "continue", take the first To Do task in the current milestone that has a tier and whose dependencies are Done. Launch its worker in the background with the task ID, one task at a time. Afterwards, check that the task is Done with every criterion ticked, `git status` is clean, and the checks pass for the parts it touched: `cd cli && go test ./... && golangci-lint run ./...`, `cd server && make test && make lint`, `cd ui && make lint && make typecheck && make test`. If not, run it once more a tier up. Report in a few lines and go on to the next task.
  - A tier 4 worker first writes a plan into the task and stops. Show the plan to the user, and once they approve it, launch the worker again with "plan approved".
  - Stop and ask the user when a task is Blocked, a tier 4 plan is waiting, a decision is proposed, or planning is due: tasks without a tier, drafts waiting, or the milestone done. On their go, launch `planner` with the milestone and show its report.
  - State lives in Backlog, so `/clear` loses nothing. Clear after a milestone or when the session gets long, then "continue".
- When starting a task, set its assignee to the agent doing the work (`@codex`, `@opencode` or `@claude`). These are Backlog labels, not routing.
- Work the current milestone. Build only what it needs.
- Commits: `feat:`, `fix:`, `chore:`, `docs:`, `build:`, `test:`, optional scope like `feat(cli):`. Short subject, no body, no trailers.


<!-- BACKLOG.MD GUIDELINES START -->
<!-- backlog.md-instructions-version: 1.55.3 -->
<CRITICAL_INSTRUCTION>

## Backlog.md Workflow

This project uses Backlog.md for task and project management.

**At the beginning of each conversation in this project, run `backlog instructions overview` before answering or taking action. Re-read it only if you have not read it yet in the current conversation.**

Use the overview to decide whether to search, read, create, or update Backlog tasks.

Before task lifecycle actions, read the matching detailed guide:
- `backlog instructions task-creation` before creating or splitting tasks
- `backlog instructions task-execution` before planning, changing status or assignee, adding a plan or implementation notes, or implementing task work
- `backlog instructions task-finalization` before checking acceptance criteria, writing final summaries, or moving tasks to terminal statuses

Use `backlog <command> --help` before running unfamiliar commands. Help shows options, fields, and examples.

Do not edit Backlog task, draft, document, decision, or milestone markdown files directly. Use the `backlog` CLI so metadata, relationships, and history stay consistent.

</CRITICAL_INSTRUCTION>
<!-- BACKLOG.MD GUIDELINES END -->
