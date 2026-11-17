# Kubin (WIP)

![Kubin Logo](docs/assets/logo_nobg.png)

Pastebin for Kubernetes clusters. Run one command, get a link, and anyone with the link can browse the cluster as it was, in a Lens-like web UI. Self-hosting is one container.

```bash
kubin create
# https://kubin.example.com/s/7fK2q...
```

It replaces attaching cluster dumps to tickets.

Planned self-hosting (not working yet, see m-0):

```bash
docker run -p 8080:8080 -v kubin-data:/data ghcr.io/3nd3r1/kubin
kubin create --server http://localhost:8080
```

## Status

Early. The CLI collects namespaces and pods into a local `.tar.gz`. The server and UI are skeletons. Nothing uploads yet. The work is on the board, see [Development](#development).

## Layout

- `cli/`: the `kubin` command (Go, client-go)
- `server/`: the server (Go, chi)
- `ui/`: the web viewer, built into the server binary
- `charts/`, `skaffold.yaml`: Helm chart and local dev loop
- `backlog/`: tasks, milestones and decisions

How the parts fit: [docs/architecture.md](docs/architecture.md).

## Development

```bash
nix-shell            # or direnv: go, node, kind, helm, skaffold, backlog
backlog board        # what's being worked on
```

More in [docs/development.md](docs/development.md).

## License

MIT
