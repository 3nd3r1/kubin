---
id: decision-11
title: kubin create uploads to the public instance when no server is given
date: '2026-11-24T09:21:00Z'
status: proposed
---
## Context

Today `kubin create` fails unless `--server` or `KUBIN_SERVER` is set. TASK-23 as first written said "the CLI default server points at" the public instance (decision-10). That changes where a cluster dump goes when the user doesn't say: from nowhere to a public server. doc-1's goal is "run `kubin create`, get a link", which a default server delivers with no setup. TASK-47 makes the change and waits for this.

## Decision

Proposed, waiting for the user. Recommended: yes. With no `--server` and no `KUBIN_SERVER`, `kubin create` uploads to the public instance. Before collecting, it says on stderr that the snapshot goes to a public instance, that anyone with the link can open it until it expires, and how to use your own server instead. `--output` still writes a file and uploads nothing.

The alternative is to keep requiring a server and print the public instance's URL in the error.

## Consequences

The first run needs no setup, which is the point of the product. The risk is someone uploading a private cluster without noticing. The notice, short default expiry (TASK-21), `kubin delete` and redaction of secrets and env values (decision-5) limit the damage, but logs, container args and ConfigMaps are not redacted (docs/architecture.md, Security). If rejected, TASK-47 is archived.

