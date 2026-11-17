---
id: TASK-24
title: Replace the Next.js UI with a Vite app embedded in the server
status: To Do
assignee: []
created_date: '2026-11-17T10:11:00Z'
labels:
  - ui
  - server
milestone: m-0
dependencies:
  - TASK-1
priority: high
type: chore
ordinal: 4500
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Self-hosting should be one container (decision-2), so the server binary serves the UI itself. Next.js static export handles `/s/{id}` routes badly, and the current UI is only the create-next-app placeholder, so switching now costs almost nothing. Tailwind and shadcn/ui stay.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `ui/` is a Vite React app with Tailwind, shadcn/ui and client-side routing
- [ ] #2 The server embeds the built UI with go:embed and serves it at `/`, with unknown non-API paths falling back to index.html
- [ ] #3 `npm run dev` in ui proxies `/api` to a locally running server
- [ ] #4 Next.js files and dependencies are gone
<!-- AC:END -->
