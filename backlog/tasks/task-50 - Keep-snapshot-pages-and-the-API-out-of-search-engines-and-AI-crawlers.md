---
id: TASK-50
title: Keep snapshot pages and the API out of search engines and AI crawlers
status: To Do
assignee: []
created_date: '2026-11-27T14:39:00Z'
updated_date: '2026-11-27T14:40:00Z'
labels:
  - server
  - tier:2
dependencies: []
priority: high
type: bug
ordinal: 45000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Snapshot links on the public instance (`https://kubin.ender.fi/s/<id>`) hold real cluster data: resource YAML, events and container logs. Anyone with the link can open it, which is the point, but nothing stops a search engine or an AI crawler that finds a link (in a public ticket, a chat log, a GitHub issue) from indexing it and serving it to strangers. Today the server sends no `robots.txt` and no `X-Robots-Tag`. This has to land before TASK-51 (page metadata), which makes the site more attractive to crawl.

It applies to every kubin, self-hosted included, not only the public instance.

## Where

- `server/internal/web/` (the embedded UI handler serves `index.html` for every client route, `/s/...` included)
- `server/internal/api/router.go` (`/api/...` routes)

## Traps

- A `Disallow: /s/` in `robots.txt` alone is not enough: crawlers that obey it never fetch the page, so they never see a `noindex`, and can still list the bare URL from links. The `X-Robots-Tag` header on the responses is what keeps them out of the index; `robots.txt` is the polite extra.
- The SPA serves the same `index.html` for `/` and `/s/...`, so a `<meta name="robots">` in `index.html` would hide the home page too. Decide by request path on the server.
- Hashed assets under `/assets/` are shared by the home page and snapshot pages; they do not need the header.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every response under `/s/` and `/api/` carries `X-Robots-Tag: noindex, nofollow, noarchive`, including 404s and expired snapshots
- [ ] #2 The home page `/` and static files like `/icon.svg` do not carry the header
- [ ] #3 The server serves `/robots.txt` that disallows `/api/` and names the common AI crawlers (GPTBot, ClaudeBot, Google-Extended, PerplexityBot, CCBot) as disallowed from `/s/`, without relying on it for the noindex
- [ ] #4 Snapshot pages send `Referrer-Policy: no-referrer` so the snapshot URL does not leak to sites linked from it
- [ ] #5 Tests in `server/` cover the header on snapshot, API and home routes and the robots.txt body
- [ ] #6 `docs/self-hosting.md` says snapshots are kept out of search engines and how
<!-- AC:END -->
