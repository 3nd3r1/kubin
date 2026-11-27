---
id: TASK-52
title: Get Kubin mentioned where people and AI agents look for Kubernetes tools
status: To Do
assignee: []
created_date: '2026-11-27T14:39:00Z'
updated_date: '2026-11-27T14:40:00Z'
labels:
  - docs
  - tier:2
dependencies:
  - TASK-51
priority: medium
type: task
ordinal: 47000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
AI agents and search engines recommend tools that many independent places mention. Kubin today has its GitHub repo and one YouTube video. Page metadata (TASK-51) makes Kubin describable; mentions make it findable when someone asks "how do I share my cluster state in a ticket" without knowing the name.

An agent cannot post as the user. This task prepares everything so the user only reviews and submits, and tracks what went out.

Candidates to evaluate (check each is alive and accepts submissions before drafting):

- Curated lists: awesome-kubernetes (ramitsurana/awesome-kubernetes), awesome-k8s-resources, awesome-cloud-native, awesome-sre-tools
- Launch posts: Show HN, r/kubernetes, r/devops, the Kubernetes Slack #kubernetes-users etiquette permitting, dev.to or a short blog post on the user’s site (TASK-43 adds a portfolio entry)
- Directories: CNCF Landscape (check the criteria, likely not eligible yet), AlternativeTo (as an alternative to `kubectl cluster-info dump` and sosreport-style tools), Product Hunt
- Distribution that doubles as discovery: a krew plugin index entry (`kubectl kubin`) and nixpkgs (TASK-49) — note each as a follow-up task if it needs code

All drafts use the same pitch sentence as TASK-51 and link the repo, the demo video and the public instance, and state plainly what is uploaded and what is not (Secrets and env var values are stripped; logs and ConfigMaps are uploaded as-is).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A doc under `docs/` (or a backlog doc) lists each channel with link, submission rules, eligibility and status (drafted, submitted by the user, accepted, rejected)
- [ ] #2 Ready-to-submit drafts exist for each eligible curated list (the exact line or PR diff in the list’s own format)
- [ ] #3 Ready-to-post drafts exist for Show HN and r/kubernetes, each following that venue’s rules, with no marketing superlatives
- [ ] #4 Work that needs code (krew plugin, nixpkgs) is filed as its own Backlog task or linked to an existing one
- [ ] #5 Nothing is posted or submitted on the user’s behalf without the user doing it or explicitly approving it
<!-- AC:END -->
