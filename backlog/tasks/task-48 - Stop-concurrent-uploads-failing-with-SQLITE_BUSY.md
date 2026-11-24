---
id: TASK-48
title: Stop concurrent uploads failing with SQLITE_BUSY
status: Done
assignee:
  - '@claude'
created_date: '2026-11-24T10:29:00Z'
updated_date: '2026-11-24T11:06:00Z'
labels:
  - server
  - tier:2
milestone: m-3
dependencies: []
priority: high
type: bug
ordinal: 1500
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Concurrent uploads of medium and big snapshots fail with 500 `database is locked (5) (SQLITE_BUSY)`. Measured on 2026-11-24 (docs/capacity.md): 7 of 40 uploads of a 600-resource snapshot at 8 concurrent, and 4 of 8 uploads of a 2,400-resource snapshot at 4 concurrent. One at a time, every upload succeeds.

Why: `OpenSQLite` (`server/internal/store/sqlite.go`) opens one `*sql.DB` with `_txlock=immediate` and `busy_timeout(5000)`, and `database/sql` hands each upload its own connection. Each upload writes its whole resource index in one transaction (`server/internal/store/sql.go`), which holds the SQLite write lock for about 0.3 s (600 resources) to 4 s (2,400 resources). The other uploads wait inside SQLite and give up after 5 s. WAL reads are not affected.

The public instance (TASK-23) runs on SQLite (decision-12), so this has to be fixed first. The usual fix is to queue writers in Go instead of in SQLite: a second `*sql.DB` for writes with `SetMaxOpenConns(1)`, reads staying on the pooled one. Raising busy_timeout alone only moves the limit. The `store.Store` interface stays the same.

Out of scope: Postgres, splitting the upload transaction, the upload rate limit (TASK-22).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A store test runs at least 8 concurrent uploads of a snapshot with at least 500 resources against a temp SQLite file and all of them succeed
- [x] #2 Reads keep working while an upload transaction is open (test)
- [x] #3 No SQLITE_BUSY on concurrent uploads in the server log when repeating the docs/capacity.md upload run (8 concurrent medium uploads)
- [x] #4 docs/capacity.md no longer lists the SQLITE_BUSY failures as open, and states the result of the re-run
<!-- AC:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a second SQLite pool (wdb, MaxOpenConns 1) used by BeginSnapshot so uploads queue in Go; reads stay on the pooled db. Store interface unchanged. Tests: 8 concurrent 600-resource uploads succeed (race clean), reads work while an upload tx is open. Re-ran docs/capacity.md upload run: 40 medium uploads at 8 concurrent, 0 failed, no SQLITE_BUSY in log. Caveat: used a synthetic seed snapshot (no demo cluster), big case not re-run; docs say so.
<!-- SECTION:FINAL_SUMMARY:END -->
