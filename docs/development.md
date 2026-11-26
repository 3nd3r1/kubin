# Development

All tools come from `nix-shell` (or `direnv allow` once). That includes the pinned `backlog` CLI in `tools/bin/`.

## What works today

```bash
make check                            # lint, typecheck, test and build every part; run before committing (after `make -C ui install` once)
cd cli && make test                   # unit tests
cd server && make test                # unit tests
cd server && make run                 # server on :8080, data in server/data/
cd cli && go run . create --server http://localhost:8080   # snapshot the current kube context, print the link
cd cli && go run . create -o cluster.tar.gz                # or keep the tarball instead
cd ui && npm ci && make dev          # UI on :5173 with hot reload, /api proxied to :8080
cd ui && make build                   # build the UI into the server; rebuild the server to embed it
make image                            # the release image as kubin:dev, UI included
docker run -p 8080:8080 -v kubin-data:/data kubin:dev   # run it like a self-hoster, see self-hosting.md
```

The `Dockerfile` is at the repo root because the image needs both `ui/` and `server/`.

## Local cluster

```bash
kind create cluster --name kubin
skaffold dev                 # builds the image, deploys charts/kubin to kubin-dev and port-forwards :8080
```

Use `skaffold run` for a one-shot deploy, then `kubectl -n kubin-dev port-forward svc/kubin-dev 8080:8080`.

## Board

```bash
backlog board                       # kanban in the terminal
backlog browser                     # same in a browser
backlog task list -s "To Do" --plain
backlog milestone list
```

The direction is in `backlog doc view doc-1`. Decisions are in `backlog decision list`.
