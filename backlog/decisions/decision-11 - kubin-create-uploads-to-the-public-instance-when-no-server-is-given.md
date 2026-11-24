---
id: decision-11
title: kubin create uploads to the public instance when no server is given
date: '2026-11-24T09:21:00Z'
status: accepted
---
## Context

Today `kubin create` fails unless `--server` or `KUBIN_SERVER` is set. doc-1's goal is "run `kubin create`, get a link", which a default server delivers with no setup. The public instance is `https://kubin.ender.fi` on the user's Coolify server (decision-12). A default changes where a cluster dump goes when the user doesn't say: from nowhere to a public server. The likely accident is not a stranger uploading on purpose, but someone on the wrong kube context uploading a cluster they didn't mean to share.

A login was considered and rejected: it adds accounts and a sign-up step, which undoes the point of the default, and auth is parked in doc-1. It also wouldn't catch the wrong-context accident.

## Decision

The user agreed to this on 2026-11-24.

- With no `--server` and no `KUBIN_SERVER`, `kubin create` uploads to `https://kubin.ender.fi`.
- The first time, after collecting and before uploading, it shows what was collected (context, API server, namespaces, resources, logs, size) and asks on stderr whether to upload to the public instance, that anyone with the link can open it until it expires, and that `--server` or `KUBIN_SERVER` picks your own kubin. The default answer is No.
- A yes is remembered in the user config dir, so later runs don't ask. They still print the summary line naming the context and the public instance before uploading. Deleting the stored answer brings the prompt back.
- Without a terminal (CI, scripts, pipes) and without a remembered yes, it refuses unless `--yes` is given.
- An explicit `--server` or `KUBIN_SERVER` never prompts, even when it names the public instance. `--output` writes a file and uploads nothing.

## Consequences

The first run needs no setup, and the one question is asked once. A remembered yes means a later run on the wrong context uploads without asking, so the summary line before the upload is the remaining guard; with the short default expiry (TASK-21), `kubin delete` (TASK-45) and redaction of secrets and env values (decision-5) it limits the damage. Logs, container args and ConfigMaps are not redacted (docs/architecture.md, Security). TASK-47 implements this.
