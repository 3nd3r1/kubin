---
id: decision-4
title: How agents bring questions to the user
date: '2026-11-18T08:44:00Z'
status: accepted
---
## Context

On 2026-11-18 the user asked how questions and decisions should flow between agents and them. Options considered: always ask in chat, a queue of proposed decisions the user accepts or rejects, batching questions at milestone ends, and having the user place every new task in a milestone.

## Decision

The user chose this on 2026-11-18:

- A direction call that blocks the work at hand is asked in chat, as one question with a recommendation, and recorded as an accepted decision.
- A direction call that can wait is written as a `proposed` decision with the context, the recommended choice and its consequences. The agent keeps working on things it doesn't block. The user answers by setting the status to `accepted` or `rejected: <why>`.
- Agents keep owning tasks, including which milestone they go in.

## Consequences

Agents check `backlog decision list` at the start of a session and act on answered proposals. Questions the user isn't needed for, like implementation choices, are still not asked. The Workflow section of CLAUDE.md says this.

