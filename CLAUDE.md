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
- The user owns the direction (doc-1): the goal, the milestone order, pivots. Bring them only direction calls, as one question with a recommendation, and record each answer with `backlog decision create`. Only the user edits doc-1.
- Agents own the tasks: create, split, update and close them without asking.
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
