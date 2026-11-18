# Snapshot format v2

A snapshot is a gzipped tar written by `kubin create` and read by the server. This file is the contract between them. Any change to it bumps `format_version`.

## Layout

```
manifest.json
resources/<group>/<version>/<Kind>/<namespace>/<name>.json
resources/<group>/<version>/<Kind>/_cluster/<name>.json
logs/<namespace>/<pod>/<container>.log
logs/<namespace>/<pod>/<container>.previous.log
```

- `manifest.json` is the first entry, so a reader can reject an unknown version before reading the rest.
- `<group>` is the API group, or `core` for the core group (Pods, Namespaces, ...).
- `<Kind>` is the kind as the API writes it, like `Pod` or `Deployment`.
- Cluster-scoped objects use `_cluster` in place of a namespace. `_cluster` can't clash with a real namespace because namespace names can't contain `_`.
- After the manifest, entries are sorted by path.
- Only regular files. No directories, links or other entry types.
- `<container>.log` is the current log of a container, or of its last run if it is waiting to restart. `<container>.previous.log` is the run before that, written for containers that restarted when the kubelet still has it. The kubelet keeps one dead container per name, so a container that is crash looping usually has its crash in `<container>.log` and no previous log. Container names can't contain dots, so the two never clash. Init containers are included.
- A container that never started has no log file. An empty file means it ran and wrote nothing.

## manifest.json

```json
{
  "format_version": 2,
  "created_at": "2026-11-17T10:00:00Z",
  "cli_version": "v0.1.0",
  "cluster": {
    "context": "kind-kubin",
    "name": "kind-kubin",
    "server_version": "v1.33.1"
  },
  "resources": [
    { "group": "core", "version": "v1", "kind": "Namespace", "count": 5 },
    { "group": "core", "version": "v1", "kind": "Pod", "count": 12 }
  ],
  "logs": {
    "tail_lines": 1000,
    "max_bytes": 1048576,
    "files": [
      { "namespace": "prod", "pod": "web-0", "container": "app", "previous": false, "truncated": true },
      { "namespace": "prod", "pod": "web-0", "container": "app", "previous": true, "truncated": false }
    ]
  }
}
```

| Field | Meaning |
|---|---|
| `format_version` | always `2` for this spec |
| `created_at` | when collection started, RFC 3339 in UTC |
| `cli_version` | version of the `kubin` binary, `dev` for local builds |
| `cluster.context` | kubeconfig context used |
| `cluster.name` | cluster name from that context in the kubeconfig |
| `cluster.server_version` | `gitVersion` the API server reported |
| `resources` | one entry per kind present, sorted by group, version, kind. `count` is the number of files under that kind |
| `logs.tail_lines` | the most lines kept per log file (`kubin create --log-tail`) |
| `logs.max_bytes` | the most bytes kept per log file |
| `logs.files` | one entry per log file, sorted by path. `truncated` means older lines were dropped to stay within the limits; the file then starts at a line boundary unless one line was longer than `max_bytes` |

The API server URL isn't recorded, because it can be an internal hostname that shouldn't end up in a shared link.

## Resource files

Each file is one object as JSON, as the API server returned it, with these changes:

- `apiVersion` and `kind` are always set. List calls leave them empty on items, so the CLI fills them in.
- `metadata.managedFields` is removed. It is large and not useful when reading a snapshot.

What else gets removed for safety, like Secret values, is defined by the redaction rules in TASK-12. Those rules will be listed in the manifest.

## Logs

Each log file is the end of a container log as the API server returned it, cut to `tail_lines` lines and `max_bytes` bytes. Logs are not redacted: kubin can't tell a credential in a log line from other text.

## Reading rules for the server

- Reject the snapshot if the first entry isn't `manifest.json` or `format_version` is unknown.
- Read v1 too. It is v2 without `.previous.log` files and without the `logs` section in the manifest.
- Reject any path that isn't in the layout above, including any path containing `..`.
- Treat counts in the manifest as a hint. Index what is actually in the archive.

## Changes

- v2 (TASK-11): logs are written, `<container>.previous.log` is added, and the manifest has a `logs` section.
- v1 (TASK-2): first version. `logs/` was reserved but never written.
