---
id: TASK-24
title: Replace the Next.js UI with a Vite app embedded in the server
status: Done
assignee:
  - '@claude'
created_date: '2026-11-17T10:11:00Z'
updated_date: '2026-11-17T14:34:00Z'
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
- [x] #1 `ui/` is a Vite React app with Tailwind, shadcn/ui and client-side routing
- [x] #2 The server embeds the built UI with go:embed and serves it at `/`, with unknown non-API paths falling back to index.html
- [x] #3 `npm run dev` in ui proxies `/api` to a locally running server
- [x] #4 Next.js files and dependencies are gone
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. ui: drop Next.js (config, app dir, next deps, next eslint config, placeholder svgs). Vite + React + TypeScript, Tailwind v4 through @tailwindcss/vite with the theme moved into CSS, react-router for /, /s/:id and a not-found page. Keep the shadcn components, cn helper and prettier config; replace next-themes with a small theme provider; Geist from @fontsource so the self-hosted UI loads no external fonts.
2. vite build writes to server/internal/web/dist; a .gitkeep keeps the embed pattern valid so the server builds and tests without node.
3. server/internal/web: serve the embedded files, index.html for any other non-API path (no-cache), long cache for hashed /assets. Unknown /api paths stay JSON 404s. Without a built UI, / says how to build it.
4. vite dev server proxies /api to localhost:8080.
5. /s/:id shows the cluster name and counts from the API so the same-origin wiring is exercised; TASK-6 builds the real page.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Stack: Vite 8, React 19, react-router 8, Tailwind 4.3 via @tailwindcss/vite with the old HSL theme moved into CSS (@theme inline, no tailwind.config), TypeScript 6, eslint 10 flat config. Geist comes from @fontsource so the UI makes no requests to other origins. next-themes replaced by a small provider; index.html sets the theme class before first paint. The build writes to server/internal/web/dist and touches dist/.gitkeep; only .gitkeep is committed, and the root and server .gitignore rules were narrowed so git can see it. A server built without the UI answers non-API paths with 503 and how to build it. Unknown /api paths are JSON 404s, never index.html. Verified: server tests, a clean checkout of the index builds and tests the server without node, ui typecheck/lint/build, the Vite dev proxy reaches the server (KUBIN_API overrides the target), and in Chrome: /s/<id> from a kind upload shows cluster and counts, unknown IDs show the not-found page, client-side navigation and the theme toggle work, no console errors.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
ui/ is now a Vite React app (Tailwind 4, shadcn/ui, react-router) with routes for /, /s/:id and a not-found page; the Next.js files and dependencies are gone. make build in ui writes into server/internal/web/dist, the server embeds it with go:embed and serves it for every non-API path, falling back to index.html (no-cache) and caching hashed assets for a year. npm run dev proxies /api to localhost:8080. /s/:id shows the cluster name, Kubernetes version, capture time and counts as a first view; TASK-6 builds the real page on it. The server Dockerfile still builds without the UI until TASK-25 adds the node stage.
<!-- SECTION:FINAL_SUMMARY:END -->
