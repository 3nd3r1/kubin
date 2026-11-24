---
id: TASK-26
title: Postgres as an optional database
status: To Do
assignee: []
created_date: '2026-11-17T10:11:00Z'
updated_date: '2026-11-24T09:36:00Z'
labels:
  - server
  - tier:3
milestone: m-3
dependencies:
  - TASK-3
priority: medium
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
SQLite is the default (decision-2), but several server replicas can't share one SQLite file. The Helm chart (TASK-7), docker compose (TASK-27) and the public instance (TASK-23) need Postgres. CLAUDE.md: SQL stays portable, and server tests use real databases.

## Where

- `server/internal/store/`: `sqlite.go` is the pattern (open, `migrate(ctx, db, dialect)`, return `*SQL`). Add `postgres.go` with `OpenPostgres(ctx, url string) (*SQL, error)` on `github.com/jackc/pgx/v5/stdlib`, which is pure Go, so the image keeps `CGO_ENABLED=0`.
- `server/internal/store/migrations/postgres/`: the same file numbers and names as `migrations/sqlite/`.
- `server/internal/config/config.go`: `DatabaseURL string` from `SERVER_DATABASE_URL`. Empty means SQLite at `<SERVER_DATA_DIR>/kubin.db`, as today.
- `server/cmd/main.go`: pick the store from config and log which one. Never log the URL, which holds the password.
- Tests: `sqlite_test.go` runs every store test on SQLite through `openTest`. Run the shared store tests against each dialect (a table of openers, or subtests); SQLite-only tests can stay SQLite-only. Postgres comes from a container the test starts with testcontainers-go (`github.com/testcontainers/testcontainers-go/modules/postgres`), each test on a fresh database or schema. `KUBIN_TEST_POSTGRES_URL` points the tests at an existing server instead.
- `docs/self-hosting.md`: config table, and when to pick Postgres over SQLite.

## Out of scope

- The S3 blob store (TASK-20), the chart and compose files.
- Copying data from SQLite to Postgres. Document that switching starts empty.
- The `store.Store` interface and the API stay as they are.

## Traps

- Postgres sorts TEXT by the database collation and SQLite compares bytes. `ListResources` and `CountResources` promise group, version, kind, namespace, name order, and tests compare exact order. Declare those columns `COLLATE "C"` in the Postgres migration so both sort the same.
- Use `TIMESTAMPTZ` in Postgres. The store writes `.UTC()` times; check they read back equal in both dialects.
- `$N` placeholders already work in both. `boolInt` stores booleans as integers; keep `previous` an INTEGER in Postgres so the same queries work.
- `migrate` sends a whole file in one Exec, and `0002_logs.sql` has two statements. With pgx that works only without arguments (simple protocol); check it.
- Replicas start together and each runs `migrate`. In Postgres, hold `pg_advisory_lock` around it so a migration isn't applied twice.
- Docker-backed tests may skip only when Docker is unreachable and `CI` is unset. With `CI=true` they fail instead of skipping, so CI can't pass without running them.
- If TASK-21 has landed, its migration needs a Postgres version too.
- TASK-20 also starts a container in tests. Whichever lands second reuses the first one's helper.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 With `SERVER_DATABASE_URL=postgres://...` the server uses Postgres; unset, it uses SQLite in SERVER_DATA_DIR as today
- [ ] #2 `migrations/postgres` has the same versions as `migrations/sqlite`, and a test fails when they differ
- [ ] #3 The shared store tests run against SQLite and against Postgres in a container the test starts, and pass on both, including the sort order of ListResources
- [ ] #4 Two OpenPostgres calls at once against an empty database both succeed and each migration is applied once (test)
- [ ] #5 /readyz answers 503 when Postgres is unreachable
- [ ] #6 docs/self-hosting.md documents SERVER_DATABASE_URL and when to use Postgres instead of SQLite
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 cd server && CI=true make test && make lint passes, with Docker running so the Postgres tests run rather than skip
- [ ] #2 Manual: docker run -d --name kubin-pg -e POSTGRES_PASSWORD=kubin -p 15432:5432 postgres:17-alpine, then cd server && SERVER_DATABASE_URL="postgres://postgres:kubin@localhost:15432/postgres?sslmode=disable" make run, and curl localhost:8080/readyz prints ok
<!-- DOD:END -->
