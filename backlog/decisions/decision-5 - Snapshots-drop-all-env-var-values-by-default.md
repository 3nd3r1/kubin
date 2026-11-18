---
id: decision-5
title: Snapshots drop all env var values by default
date: '2026-11-18T08:47:00Z'
status: accepted
---
## Context

A snapshot link is effectively public, so a leaked credential is the worst bug kubin can have (TASK-12). Pod and workload specs often carry credentials as literal env values. The choice was between dropping every literal env value, or keeping them and hiding only names that look sensitive, like PASSWORD or TOKEN. Name matching misses values like `DATABASE_URL=postgres://user:pass@...`.

## Decision

The user chose this on 2026-11-18: `kubin create` drops every literal env value by default. Env var names and `valueFrom` references stay, since they say where a value comes from without leaking it. A flag keeps the values for users who know their specs are clean.

## Consequences

A snapshot can't answer "what was FOO set to" unless the user opts in. The manifest says that env values were dropped, so the viewer can show them as redacted rather than empty.

