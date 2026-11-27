---
id: TASK-51
title: >-
  Describe Kubin to search engines and AI agents: page metadata, llms.txt and
  repo metadata
status: To Do
assignee: []
created_date: '2026-11-27T14:39:00Z'
updated_date: '2026-11-27T14:40:00Z'
labels:
  - ui
  - server
  - tier:3
dependencies:
  - TASK-50
priority: high
type: feature
ordinal: 46000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
When someone asks an AI agent or a search engine for a way to share a Kubernetes cluster as a link, or asks what Kubin is, Kubin should come up with a correct description. Today it mostly does not:

- The public instance `https://kubin.ender.fi` is a React app. Its HTML is `<title>Kubin</title>` and an empty `<div id="root">`. Crawlers and agent fetchers that do not run JavaScript see nothing: no description, no Open Graph tags, no text.
- The GitHub repo `3nd3r1/kubin` has no topics, no homepage, and a description that never says "Kubernetes".
- The name is ambiguous (a surname, and voice input turns it into "Cuban"), so every surface must say "Kubin" next to "Kubernetes" in the first sentence.

The one-line pitch to use everywhere: "Kubin is a pastebin for Kubernetes clusters: `kubin create` captures a cluster and prints a link anyone can open to browse resources, events and logs." Keep the wording consistent across the page, `llms.txt`, the README and the repo.

Depends on TASK-50: snapshot pages must be kept out of indexes before the site is made easier to crawl.

## Where

- `ui/index.html`, `ui/src/pages/home.tsx`, `ui/public/`
- `server/internal/web/` if a file needs serving with a specific type
- `README.md` first lines
- GitHub repo settings through `gh api -X PATCH repos/3nd3r1/kubin` and `PUT repos/3nd3r1/kubin/topics` (ask the user before changing the live repo)

## Out of scope

A separate marketing site or docs site, a new service, server-side rendering of snapshot pages. One server, one image (decision-1, decision-2).

## Traps

- Self-hosted kubins serve the same `index.html`. Metadata must describe Kubin the project, not claim to be the public instance; absolute URLs (canonical, `og:url`, `og:image`) either point at the project (GitHub) or are left relative.
- Static text placed inside `#root` for non-JS readers must be replaced cleanly when React mounts, with no flash on normal loads.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `curl` of `/` (no JavaScript) returns a title naming Kubernetes, a meta description, Open Graph and Twitter card tags with an image, and a short readable paragraph with the pitch, the install command and a link to GitHub
- [ ] #2 The page has JSON-LD `SoftwareApplication` data: name, description, license MIT, `codeRepository`, operating systems, free
- [ ] #3 `/llms.txt` exists, follows the llmstxt.org layout, and covers what Kubin is, install, `kubin create`, self-hosting, privacy (what is and is not collected) and links to the docs on GitHub
- [ ] #4 `/robots.txt` from the noindex task allows `/` and `/llms.txt` for all crawlers, AI crawlers included
- [ ] #5 README opens with the same pitch sentence; repo description, homepage (`https://kubin.ender.fi`) and topics (kubernetes, k8s, kubectl, debugging, troubleshooting, devops, pastebin, cli) are set after the user agrees
- [ ] #6 Snapshot pages still render normally and still carry the noindex header
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 make lint in ui/ and make test in server/ pass
<!-- DOD:END -->
